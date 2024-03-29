package explore

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"blinders/packages/db"
	"blinders/packages/db/models"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Explorer interface {
	// Suggest returns list of users that maybe match with given user
	Suggest(userID string) ([]models.MatchInfo, error)
	// AddUserMatchInformation adds user match information to the database.
	AddUserMatchInformation(info models.MatchInfo) (models.MatchInfo, error)
	// AddEmbedding adds user embed vector to the vector database.
	AddEmbedding(userID primitive.ObjectID, embed EmbeddingVector) error
}

type MongoExplorer struct {
	Db          *db.MongoManager
	RedisClient *redis.Client
}

func NewMongoExplorer(Db *db.MongoManager, RedisClient *redis.Client) *MongoExplorer {
	return &MongoExplorer{
		Db:          Db,
		RedisClient: RedisClient,
	}
}

/*
Suggest  recommends 5 users who are not friends of the current user.

TODO: The goal is to recommend users with whom the current user may communicate effectively.
These users should either be fluent in the language the current user is learning or actively learning the same language.
To achieve this, we will filter the Users database to extract users who are native speakers of the language the current user is learning,
or users who are currently learning the same language as the current user.

We will then use KNN-search in the filtered space to identify 5 users that may match with the current user.
*/
func (m *MongoExplorer) Suggest(userID string) ([]models.MatchInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf(
			"explore: cannot parse object id with given hex string(%s), err: %v\n",
			userID,
			err,
		)
		return nil, err
	}

	user, err := m.Db.Users.GetUserByID(oid)
	if err != nil {
		log.Println("explore: cannot get user by id, err", err)
		return nil, err
	}

	// JSONGet return value wrapped in an array.
	jsonStr, err := m.RedisClient.JSONGet(ctx, CreateMatchKeyWithUserID(userID), "$.embed").Result()
	if err != nil {
		log.Println("explore: cannot get explore entry in redis, err", err)
		return []models.MatchInfo{}, fmt.Errorf(
			"explore profile not found, might need to check onboarding status",
		)
	}
	var embedArr []EmbeddingVector
	if err := json.Unmarshal([]byte(jsonStr), &embedArr); err != nil {
		log.Println("explore: cannot unmarshall embed vector, err", err)
		return []models.MatchInfo{}, fmt.Errorf("something went wrong")
	}
	embed := embedArr[0]

	// exclude friends of current user
	excludeFilter := userID
	for _, friendID := range user.FriendIDs {
		excludeFilter += " | " + friendID.Hex()
	}
	excludeFilter = fmt.Sprintf("-@id:(%s)", excludeFilter)

	candidates, err := m.Db.Matches.GetUsersByLanguage(user.ID, 1000)
	if err != nil {
		log.Println("explore: cannot explore candidates, err", err)
		return nil, err
	}

	includeFilter := ""
	if len(candidates) != 0 {
		includeFilter = candidates[0]
		for idx := 1; idx < len(candidates); idx++ {
			includeFilter += " | " + candidates[idx]
		}
		includeFilter = fmt.Sprintf("@id:(%s)", includeFilter)
	}

	prefilter := fmt.Sprintf("(%s %s)", excludeFilter, includeFilter)

	cmd := m.RedisClient.Do(ctx,
		"FT.SEARCH",
		"idx:match_vss",
		fmt.Sprintf("%s=>[KNN 5 @embed $query_vector as vector_score]", prefilter),
		"SORTBY", "vector_score",
		"PARAMS", "2",
		"query_vector", &embed,
		"DIALECT", "2",
		"RETURN", "1", "id",
	)
	if err := cmd.Err(); err != nil {
		log.Println("explore: cannot perform knn search in vector database, err", err)
		return nil, err
	}

	var res []models.MatchInfo
	for _, doc := range cmd.Val().(map[any]any)["results"].([]any) {
		userID := doc.(map[any]any)["extra_attributes"].(map[any]any)["id"].(string)
		oid, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			return nil, err
		}
		user, err := m.Db.Matches.GetMatchInfoByUserID(oid)
		if err != nil {
			return nil, err
		}
		res = append(res, user)
	}

	// TODO: After the suggestion process, mark these users as suggested to prevent them from being recommended in future suggestions.
	// Idea: Recommended users will be assigned extra points, which will be added to their vector space during the vector search, making their vectors more distant from the current vector.
	// Redis does not support sorting by expression.
	return res, nil
}

/*
AddUserMatchInformation inserts information into the match database.

Currently, embedding will be handled by another service. The caller of this method must trigger a new event
to notify that a new user has been created. This allows the embedding service to update the embedding vector
in the vector database.
*/
func (m *MongoExplorer) AddUserMatchInformation(info models.MatchInfo) (models.MatchInfo, error) {
	_, err := m.Db.Users.GetUserByID(info.UserID)
	if err != nil {
		return models.MatchInfo{}, err
	}

	// duplicated match information will be handled by the repository since we have already indexed the collection with firebaseUID.
	matchInfo, err := m.Db.Matches.InsertNewRawMatchInfo(info)
	if err != nil {
		return models.MatchInfo{}, err
	}
	return matchInfo, nil
}

func (m *MongoExplorer) AddEmbedding(userID primitive.ObjectID, embed EmbeddingVector) error {
	_, err := m.Db.Matches.GetMatchInfoByUserID(userID)
	if err != nil {
		return err
	}
	fmt.Println("checkpoint")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = m.RedisClient.JSONSet(ctx,
		CreateMatchKeyWithUserID(userID.Hex()),
		"$",
		map[string]any{"embed": embed, "id": userID},
	).Err()
	return err
}
