package restapi

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"blinders/packages/auth"
	"blinders/packages/db/matchingdb"
	"blinders/packages/db/usersdb"
	"blinders/packages/transport"
	"blinders/packages/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	OnboardingService struct {
		UserRepo     *usersdb.UsersRepo
		MatchingRepo *matchingdb.MatchingRepo
		Transport    transport.Transport
		ConsumerMap  transport.ConsumerMap
	}
	OnboardingForm struct {
		Name      string   `json:"name"      form:"name"`
		Major     string   `json:"major"     form:"major"`
		Gender    string   `json:"gender"    form:"gender"`
		Native    string   `json:"native"    form:"native"`
		Country   string   `json:"country"   form:"country"`
		Learnings []string `json:"learnings" form:"learnings"`
		Interests []string `json:"interests" form:"interests"`
		Age       int      `json:"age"       form:"age"`
	}
)

func NewOnboardingService(
	userRepo *usersdb.UsersRepo,
	matchingRepo *matchingdb.MatchingRepo,
	transporter transport.Transport,
	consumerMap transport.ConsumerMap,
) *OnboardingService {
	return &OnboardingService{
		UserRepo:     userRepo,
		MatchingRepo: matchingRepo,
		Transport:    transporter,
		ConsumerMap:  consumerMap,
	}
}

func (s *OnboardingService) PostOnboardingForm() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userAuth, ok := ctx.Locals(auth.UserAuthKey).(*auth.UserAuth)
		if !ok {
			log.Fatalln("cannot get auth user")
		}

		var formValue OnboardingForm
		if err := ctx.BodyParser(&formValue); err != nil {
			log.Println("cannot parse form value", err)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err})
		}

		uid, err := primitive.ObjectIDFromHex(userAuth.ID)
		if err != nil {
			log.Println("cannot get objectID from userID", err)
			return ctx.Status(fiber.StatusBadRequest).
				JSON(fiber.Map{"error": "cannot get objectID from userID " + err.Error()})
		}
		matchInfo, err := utils.JSONConvert[matchingdb.MatchInfo](formValue)
		if err != nil {
			log.Println("cannot unmarshal match info from form value", err)
			return ctx.Status(fiber.StatusBadRequest).
				JSON(fiber.Map{"error": "cannot unmarshal match info from form value" + err.Error()})
		}
		matchInfo.UserID = uid

		payload, _ := json.Marshal(transport.AddUserMatchInfoRequest{
			Request: transport.Request{Type: transport.AddUserMatchInfo},
			Payload: *matchInfo,
		})
		resBytes, err := s.Transport.Request(
			context.Background(),
			s.ConsumerMap[transport.Explore],
			payload,
		)
		if err != nil {
			log.Println("invoke explore error", err)
			return ctx.SendStatus(http.StatusBadRequest)
		}

		res, err := utils.ParseJSON[transport.AddUserMatchInfoResponse](resBytes)
		if err != nil {
			log.Println(string(resBytes))
			log.Println("explore response error", err)
			return ctx.SendStatus(http.StatusBadRequest)
		}

		if res != nil && res.Error != "" {
			log.Println("from explore response error", res.Error)
		}

		return nil
	}
}
