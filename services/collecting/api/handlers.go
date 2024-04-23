package collectingapi

import (
	"fmt"
	"log"

	"blinders/packages/collecting"
	"blinders/packages/transport"
	"blinders/packages/utils"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s Service) HandlePushEvent(ctx *fiber.Ctx) error {
	req, err := utils.ParseJSON[transport.CollectEventRequest](ctx.Body())
	if err != nil {
		log.Printf("collecting: cannot get collect event from request's body, err: %v\n", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot get event from body"})
	}

	if req.Request.Type != transport.CollectEvent {
		log.Printf("collecting: event type mismatch, type: %v\n", req.Request.Type)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot get event from body"})
	}

	eventID, err := s.HandleGenericEvent(req.Data)
	if err != nil {
		log.Printf("collecting: cannot process generic event, err: %v\n", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"id": eventID})
}

// TODO: aws transport.push fail to this endpoint, fix
func (s Service) HandleGetEvent(ctx *fiber.Ctx) error {
	req, err := utils.ParseJSON[transport.GetEventRequest](ctx.Body())
	if err != nil {
		log.Printf("collecting: cannot get event request from body, err: %v\n", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot get request body"})
	}

	userOID, err := primitive.ObjectIDFromHex(req.UserID)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot get userID from request"})
	}

	switch req.Type {
	case collecting.EventTypeExplain:
		logs, err := s.Collector.GetExplainLogByUserID(userOID)
		if err != nil {
			log.Println("collecting: cannot get logs of user", err)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot get user's log event"})
		}

		return ctx.Status(fiber.StatusOK).JSON(logs[:req.NumReturn])

	case collecting.EventTypeTranslate:
		logs, err := s.Collector.GetTranslateLogByUserID(userOID)
		if err != nil {
			log.Println("collecting: cannot get logs of user", err)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot get user's log event"})
		}

		return ctx.Status(fiber.StatusOK).JSON(logs[:req.NumReturn])

	case collecting.EventTypeSuggestPracticeUnit:
		logs, err := s.Collector.GetSuggestPracticeUnitLogByUserID(userOID)
		if err != nil {
			log.Println("collecting: cannot get logs of user", err)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot get user's log event"})
		}

		return ctx.Status(fiber.StatusOK).JSON(logs[:req.NumReturn])

	default:
		log.Printf("collecting: received undefined event type (%v)\n", req.Type)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "unsupported event"})
	}
}

// HandleGeneric will check the generic event types and then add event to correspond storage,
//
// This method return id of new added event and error if occurs.
// Error returns from this method is ready to response to client
func (s Service) HandleGenericEvent(event collecting.GenericEvent) (string, error) {
	switch event.Type {
	case collecting.EventTypeSuggestPracticeUnit:
		event, err := utils.JSONConvert[collecting.SuggestPracticeUnitEvent](event.Payload)
		if err != nil {
			log.Printf("collecting: cannot get suggest practice unit event from payload, err: %v\n", err)
			return "", fmt.Errorf("mismatch event type and event payload")
		}

		eventLog, err := s.Collector.AddRawSuggestPracticeUnitLog(&collecting.SuggestPracticeUnitEventLog{
			SuggestPracticeUnitEvent: *event,
		})
		if err != nil {
			log.Printf("collecting: cannot add raw translate log, err: %v", err)
			return "", fmt.Errorf("cannot append translate log")
		}

		return eventLog.ID.Hex(), nil

	case collecting.EventTypeTranslate:
		event, err := utils.JSONConvert[collecting.TranslateEvent](event.Payload)
		if err != nil {
			log.Printf("collecting: cannot get translate event from payload, err: %v\n", err)
			return "", fmt.Errorf("mismatch event type and event payload")
		}

		eventLog, err := s.Collector.AddRawTranslateLog(&collecting.TranslateEventLog{
			TranslateEvent: *event,
		})
		if err != nil {
			log.Printf("collecting: cannot add raw translate log, err: %v\n", err)
			return "", fmt.Errorf("cannot append translate log")
		}

		return eventLog.ID.Hex(), nil

	case collecting.EventTypeExplain:
		event, err := utils.JSONConvert[collecting.ExplainEvent](event.Payload)
		if err != nil {
			log.Printf("collecting: cannot get explain event from payload, err :%v\n", err)
			return "", fmt.Errorf("mismatch event type and event payload")
		}

		eventLog, err := s.Collector.AddRawExplainLog(&collecting.ExplainEventLog{
			ExplainEvent: *event,
		})
		if err != nil {
			log.Printf("collecting: cannot add raw explain log, err: %v\n", err)
			return "", fmt.Errorf("cannot append explain log")
		}

		return eventLog.ID.Hex(), nil

	default:
		log.Printf("collecting: receive unsupport event, type: %v", event.Type)
		return "", fmt.Errorf("unsupported event type: %v", event.Type)
	}
}