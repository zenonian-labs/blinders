package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"blinders/packages/translate"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type TranslatePayload struct {
	Text string `json:"text"`
}

type TranslateResponse struct {
	Text       string `json:"text"`
	Translated string `json:"translated"`
	Languages  string `json:"languages"`
}

var translator translate.Translator

func init() {
	translator = translate.YandexTranslator{APIKey: os.Getenv("YANDEX_API_KEY")}
}

func HandleRequest(
	_ context.Context,
	event events.APIGatewayV2HTTPRequest,
) (events.APIGatewayV2HTTPResponse, error) {
	text, ok := event.QueryStringParameters["text"]
	if !ok {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       "required text param",
			Headers:    map[string]string{"Access-Control-Allow-Origin": "*"},
		}, nil
	}

	langs, ok := event.QueryStringParameters["languages"]
	if !ok {
		langs = "en-vi"
	}

	translated, err := translator.Translate(text, translate.Languages(langs))
	if err != nil {
		log.Println("error translating: ", err)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf("cannot translate \"%s\"", text),
			Headers:    map[string]string{"Access-Control-Allow-Origin": "*"},
		}, nil
	}

	res := TranslateResponse{
		Text:       text,
		Translated: translated,
		Languages:  langs,
	}

	resInBytes, _ := json.Marshal(res)
	return events.APIGatewayV2HTTPResponse{
		StatusCode: 200,
		Body:       string(resInBytes),
		Headers:    map[string]string{"Access-Control-Allow-Origin": "*"},
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
