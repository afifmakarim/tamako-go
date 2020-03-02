package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	os.Setenv("CHANNEL_SECRET", "1be4fbd94dd56738610507b0c20de840")
	os.Setenv("CHANNEL_TOKEN", "SnCXDhD8ZOMoDJv4+JdRIT/XVjkOIhdY2/v/YECnPH1hJx+ttwK0pqXsXyfemvWPadHCnqZNEIRWx/j6VzxGw4zBM1ahAJIIJyfODtGCTS2YipB0r+1NF1efb/NvGGuCFc6Kut6x6oa4J4zoX0jTyQdB04t89/1O/w1cDnyilFU=")

	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
				fmt.Fprint(w, "Invalid Signature")
			} else {
				w.WriteHeader(500)
				fmt.Fprint(w, "ok")
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				case *linebot.StickerMessage:
					replyMessage := fmt.Sprintf(
						"sticker id is %s, stickerResourceType is %s", message.StickerID, message.StickerResourceType)
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})
	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
}
