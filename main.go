package main

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/go-sdk-core/core"
	"github.com/gofiber/fiber"
	"github.com/jeffotoni/gconcat"
	"github.com/watson-developer-cloud/go-sdk/speechtotextv1"
	"io"
	"os"
)

type Dados struct {
	Car string `form:"car"`
	Text string `form:"text"`
}

func main() {
	app := fiber.New()
	app.Settings.BodyLimit = 12 * 1024 * 1024 // 12 megabytes
	app.Post("/behinthecode8", func(c *fiber.Ctx) {
		println("INICIANDO")
		d := new(Dados)
		err := c.BodyParser(d)
		switch err {
		case nil:
			break
		default:
			c.Status(400).Send(err)
			return
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
					c.Status(500).Send(err)
					return
				}
			default:
				fmt.Println("ERROR:", err)
				c.Status(500).Send(err)
				return
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
			audioFile, audioFileErr = os.Open(gconcat.Build("./audio/" , file.Filename))
			if audioFileErr != nil {
				panic(audioFileErr)
			}
			result, _, responseErr := speechToText.Recognize(
				&speechtotextv1.RecognizeOptions{
					Audio:                     audioFile,
					ContentType:               core.StringPtr("application/octet-stream"),
					Model: core.StringPtr("pt-BR_BroadbandModel"),
				},
			)
			if responseErr != nil {
				//println(responseErr.Error())
				panic(responseErr)
			}
			for i,result1:=range result.Results{
				print(*result1.Alternatives[i].Confidence,"\n",*result1.Alternatives[i].Transcript)
			}
			b, err := json.MarshalIndent(result, "", "  ")
			switch err{
			case nil:
			break
			default:
				println("Err:",err)
			}
			println(string(b))

		default:
			println(d.Text)
		}
		c.Status(200)
		return
	})

	app.Listen("8081")

}
