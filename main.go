package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	os.Setenv("CHANNEL_SECRET", "b54b13e54172b7798711da6dc37089cf")
	os.Setenv("CHANNEL_TOKEN", "4KNSn8qgfy9/9XmP8Yqs3Vi6bRNa7Q4Xes/50bZvRZizpnDkaj6oaQSsTAP9N6DFadHCnqZNEIRWx/j6VzxGw4zBM1ahAJIIJyfODtGCTS2rAqGQfQT+/pn1Eyg8/7lZCaDtWciPmT6NOBaGrF46uwdB04t89/1O/w1cDnyilFU=")

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
				//fmt.Fprint(w, "ok")
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
