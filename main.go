package main

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/go-sdk-core/core"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jeffotoni/gconcat"
	"github.com/watson-developer-cloud/go-sdk/naturallanguageunderstandingv1"
	"github.com/watson-developer-cloud/go-sdk/speechtotextv1"
	"io"
	"os"
)

type Dados struct {
	Car  string `form:"car"`
	Text string `form:"text"`
}

func main() {
	var urlnnul string

	app := fiber.New(fiber.Config{BodyLimit: 12 * 1024 * 1024})
	app.Use(logger.New(logger.Config{
		Format:     "${pid} ${status} - ${method} ${ip} ${path} ${time} \n",
		TimeFormat: "02-Jan-2006",
		Output:     os.Stdout}))

	app.Post("/behinthecode8", func(c *fiber.Ctx) error {
		println("INICIANDO")
		d := new(Dados)
		err := c.BodyParser(d)
		switch err {
		case nil:
			fmt.Println(d)
			break
		default:
			return c.Status(400).SendString(err.Error())
		}
		switch d.Text {
		case "":
			println("INICIANDO SALVAMENTO DO ARQUIVO DE AUDIO")
			file, err := c.FormFile("audio")
			switch err {
			case nil:
				err := c.SaveFile(file, gconcat.Build("./audio/", file.Filename))
				switch err {
				case nil:
					break
				default:
					fmt.Println("ERROR:", err)
					return c.Status(500).SendString(err.Error())
				}
			default:
				fmt.Println("ERROR:", err)
				return c.Status(500).SendString(err.Error())
			}
			println("ENVIANDO PARA O SST")
			authenticator := &core.IamAuthenticator{
				ApiKey: os.Getenv("APIKEY_BEHINDCODE8"),
			}

			options := &speechtotextv1.SpeechToTextV1Options{
				Authenticator: authenticator,
			}

			speechToText, speechToTextErr := speechtotextv1.NewSpeechToTextV1(options)

			if speechToTextErr != nil {
				panic(speechToTextErr)
			}

			speechToText.SetServiceURL(os.Getenv("URL1_BEHINDCODE8"))
			var audioFile io.ReadCloser
			var audioFileErr error
			audioFile, audioFileErr = os.Open(gconcat.Build("./audio/", file.Filename))
			if audioFileErr != nil {
				panic(audioFileErr)
			}
			result, _, responseErr := speechToText.Recognize(
				&speechtotextv1.RecognizeOptions{
					Audio:       audioFile,
					ContentType: core.StringPtr("application/octet-stream"),
					Model:       core.StringPtr("pt-BR_BroadbandModel"),
				},
			)
			if responseErr != nil {
				//println(responseErr.Error())
				panic(responseErr)
			}

			for i, result1 := range result.Results {
				urlnnul = gconcat.Build(urlnnul, *result1.Alternatives[i].Transcript, "\n")
				fmt.Println(*result1.Alternatives[i].Transcript, "\n")
			}
			break
		default:
			urlnnul = d.Text
		}

		println("ENVIANDO PARA O NLU")
		authenticator1 := &core.IamAuthenticator{
			ApiKey: os.Getenv("APIKEY1_BEHINDCODE8"),
		}

		options1 := &naturallanguageunderstandingv1.NaturalLanguageUnderstandingV1Options{
			Version:       "2020-09-17",
			Authenticator: authenticator1,
		}

		naturalLanguageUnderstanding, naturalLanguageUnderstandingErr := naturallanguageunderstandingv1.NewNaturalLanguageUnderstandingV1(options1)

		if naturalLanguageUnderstandingErr != nil {
			fmt.Println("ERROR:", naturalLanguageUnderstandingErr)
			return c.Status(500).SendString("deu ruim")
		}
		naturalLanguageUnderstanding.Service.SetServiceURL(os.Getenv("URL2_BEHINDCODE8"))
		id := os.Getenv("MODELOID")
		//print(id)
		/*
			result, _, responseErr := naturalLanguageUnderstanding.ListModels(
				&naturallanguageunderstandingv1.ListModelsOptions{},
			)
			if responseErr != nil {
				panic(responseErr)
			}
			b, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(b))

		*/
		result1, _, responseErr1 := naturalLanguageUnderstanding.Analyze(
			&naturallanguageunderstandingv1.AnalyzeOptions{
				Text: &urlnnul,
				Features: &naturallanguageunderstandingv1.Features{
					Entities: &naturallanguageunderstandingv1.EntitiesOptions{
						Mentions:  core.BoolPtr(true),
						Model:     core.StringPtr(id),
						Sentiment: core.BoolPtr(true),
					},
				},
			},
		)
		switch responseErr1 {
		case nil:
			b, err := json.MarshalIndent(result1, "", "   ")
			switch err {
			case nil:
				fmt.Println(string(b))
				break
			default:
				fmt.Println(err.Error())
				return c.Status(500).SendString(err.Error())
			}
			break
		default:
			return c.Status(500).SendString(responseErr1.Error())
		}
		return nil
	})

	app.Listen(":8081")
}
