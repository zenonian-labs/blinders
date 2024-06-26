package restapi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"blinders/packages/auth"
	"blinders/packages/db/chatdb"
	"blinders/packages/db/usersdb"
	"blinders/packages/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ConversationsService struct {
	ConversationsRepo *chatdb.ConversationsRepo
	MessagesRepo      *chatdb.MessagesRepo
	UsersRepo         *usersdb.UsersRepo
}

func NewConversationsService(
	convRepo *chatdb.ConversationsRepo,
	messagesRepo *chatdb.MessagesRepo,
	usersRepo *usersdb.UsersRepo,
) *ConversationsService {
	return &ConversationsService{
		ConversationsRepo: convRepo,
		MessagesRepo:      messagesRepo,
		UsersRepo:         usersRepo,
	}
}

func (s ConversationsService) GetConversationByID(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("invalid id:", err)
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"error": "invalid id",
		})
	}

	conversation, err := s.ConversationsRepo.GetConversationByID(oid)
	if err != nil {
		log.Println("can not get conversation:", err)
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"error": "can not get conversation",
		})
	}

	return ctx.Status(http.StatusOK).JSON(conversation)
}

func (s ConversationsService) GetConversationsOfUser(ctx *fiber.Ctx) error {
	userAuth := ctx.Locals(auth.UserAuthKey).(*auth.UserAuth)
	if userAuth == nil {
		return fmt.Errorf("required user auth")
	}

	userID, _ := primitive.ObjectIDFromHex(userAuth.ID)

	queryType := ctx.Query("type", "all")
	switch queryType {
	case "all":
		conversations, err := s.ConversationsRepo.GetConversationByMembers(
			[]primitive.ObjectID{userID})
		if err != nil {
			log.Println("can not get conversations:", err)
			return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": "can not get conversations",
			})
		}
		return ctx.Status(http.StatusOK).JSON(conversations)
	case "individual":
		friendID, err := primitive.ObjectIDFromHex(
			ctx.Query("friendId", ""))
		if err != nil {
			return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": "friend id is required",
			})
		}
		conversations, err := s.ConversationsRepo.GetConversationByMembers(
			[]primitive.ObjectID{userID, friendID},
			chatdb.IndividualConversation)
		if err != nil {
			log.Println("can not get conversations:", err)
			return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": "can not get conversations",
			})
		}
		return ctx.Status(http.StatusOK).JSON(conversations)
	case "group":
		conversations, err := s.ConversationsRepo.GetConversationByMembers(
			[]primitive.ObjectID{userID}, chatdb.GroupConversation)
		if err != nil {
			log.Println("can not get conversations:", err)
			return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
				"error": "can not get conversations",
			})
		}
		return ctx.Status(http.StatusOK).JSON(conversations)
	default:
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"error": "invalid query type, must be 'all', 'group' or 'individual'",
		})
	}
}

type CreateConversationDTO struct {
	Type chatdb.ConversationType `json:"type"`
}

type CreateGroupConvDTO struct {
	CreateConversationDTO `json:",inline"`
}

type CreateIndividualConvDTO struct {
	CreateConversationDTO `json:",inline"`
	FriendID              string `json:"friendId"`
}

func (s ConversationsService) CreateNewIndividualConversation(ctx *fiber.Ctx) error {
	convDTO, err := utils.ParseJSON[CreateConversationDTO](ctx.Body())
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"error": "invalid payload to create conversation",
		})
	}

	switch convDTO.Type {
	case chatdb.IndividualConversation:
		{
			convDTO, err := utils.ParseJSON[CreateIndividualConvDTO](ctx.Body())
			if err != nil {
				return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
					"error": "invalid payload to create individual conversation",
				})
			}

			authUser := ctx.Locals(auth.UserAuthKey).(*auth.UserAuth)
			userID, _ := primitive.ObjectIDFromHex(authUser.ID)

			friendID, err := primitive.ObjectIDFromHex(convDTO.FriendID)
			if err != nil {
				return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
					"error": "invalid friend id",
				})
			}

			err = s.CheckFriendRelationship(userID, friendID)
			if err != nil {
				return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
					"error": err.Error(),
				})
			}

			conv, err := s.ConversationsRepo.InsertIndividualConversation(userID, friendID)
			if err != nil {
				return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
					"error": err.Error(),
				})
			}

			return ctx.Status(http.StatusCreated).JSON(conv)

		}
	}

	return nil
}

func (s ConversationsService) CheckFriendRelationship(
	userID primitive.ObjectID,
	friendID primitive.ObjectID,
) error {
	var user usersdb.User
	err := s.UsersRepo.FindOne(context.Background(), bson.M{
		"_id":     userID,
		"friends": bson.M{"$all": []primitive.ObjectID{friendID}},
	}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("do not have friend relationship with this user")
	}

	var friend usersdb.User
	err = s.UsersRepo.FindOne(context.Background(), bson.M{
		"_id": friendID,
	}).Decode(&friend)
	if err == mongo.ErrNoDocuments {
		return fmt.Errorf("not found friend user")
	}

	return nil
}

func (s ConversationsService) GetMessagesOfConversation(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("invalid id:", err)
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"error": "invalid id",
		})
	}

	limit, err := strconv.Atoi(ctx.Query("limit", "30"))
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"error": "invalid limit",
		})
	}
	messages, err := s.MessagesRepo.GetMessagesOfConversation(oid, int64(limit))
	if err != nil {
		log.Println("can not get messages:", err)
		return ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"error": "can not get messages",
		})
	}

	return ctx.Status(http.StatusOK).JSON(messages)
}
