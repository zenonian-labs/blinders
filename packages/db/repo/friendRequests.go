package repo

import (
	"context"
	"fmt"
	"log"
	"time"

	"blinders/packages/db/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FriendRequestsRepo struct {
	Col *mongo.Collection
}

func NewFriendRequestsRepo(col *mongo.Collection) *FriendRequestsRepo {
	ctx, cal := context.WithTimeout(context.Background(), time.Second)
	defer cal()

	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.M{"firebaseUID": 1},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Println("can not create index for firebaseUID:", err)
		return nil
	}

	return &FriendRequestsRepo{
		Col: col,
	}
}

// this function creates new ID and time and insert the document to database
func (r *FriendRequestsRepo) InsertNewRawFriendRequest(
	request models.FriendRequest,
) (*models.FriendRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	request.ID = primitive.NewObjectID()
	now := primitive.NewDateTimeFromTime(time.Now())
	request.CreatedAt = now
	request.UpdatedAt = now

	upsert := true
	result, err := r.Col.UpdateOne(ctx,
		bson.M{
			"from": request.From,
			"to":   request.To,
		},
		bson.M{"$setOnInsert": request},
		&options.UpdateOptions{Upsert: &upsert})
	if err != nil {
		log.Println("can not insert friend request:", err)
		return nil, fmt.Errorf("something went wrong")
	} else if result.UpsertedCount == 0 {
		return nil, fmt.Errorf("request already existed")
	}

	return &request, err
}

func (r *FriendRequestsRepo) GetFriendRequestByFrom(
	from primitive.ObjectID,
	status models.FriendRequestStatus,
) ([]models.FriendRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var filter bson.M
	switch status {
	case models.FriendStatusPending, models.FriendStatusAccepted, models.FriendStatusDenied:
		filter = bson.M{"from": from, "status": status}
	default:
		filter = bson.M{"from": from}
	}

	cursor, err := r.Col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var requests []models.FriendRequest
	err = cursor.All(ctx, &requests)
	if err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *FriendRequestsRepo) GetFriendRequestByTo(
	to primitive.ObjectID,
	status models.FriendRequestStatus,
) ([]models.FriendRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var filter bson.M
	switch status {
	case models.FriendStatusPending, models.FriendStatusAccepted, models.FriendStatusDenied:
		filter = bson.M{"to": to, "status": status}
	default:
		filter = bson.M{"to": to}
	}

	cursor, err := r.Col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var requests []models.FriendRequest
	err = cursor.All(ctx, &requests)
	if err != nil {
		return nil, err
	}

	return requests, nil
}

func (r *FriendRequestsRepo) GetFriendRequestByID(
	id primitive.ObjectID,
) (*models.FriendRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var request models.FriendRequest
	err := r.Col.FindOne(ctx, bson.M{"_id": id}).Decode(&request)
	if err == mongo.ErrNoDocuments {
		return nil, fmt.Errorf("not found this friend request")
	} else if err != nil {
		log.Println("can not get friend request:", err)
		return nil, fmt.Errorf("something went wrong when get friend request")
	}

	return &request, nil
}

func (r *FriendRequestsRepo) UpdateFriendRequestStatusByID(
	id primitive.ObjectID,
	status models.FriendRequestStatus,
) (*models.FriendRequest, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var request models.FriendRequest
	result, err := r.Col.UpdateByID(ctx, id, bson.M{"$set": bson.M{"status": status}})
	if err != nil {
		log.Println("can not update friend request:", err)
		return nil, fmt.Errorf("something went wrong when update friend request")
	}
	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("not found this friend request")
	}

	return &request, nil
}
