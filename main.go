package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

// KitchenSink app
type KitchenSink struct {
	bot         *linebot.Client
	appBaseURL  string
	downloadDir string
}

func (app *KitchenSink) main() {
	// load configuration
	var err error
	app.bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
	if err != nil {
		log.Println("Bot Initial Error:", err)
	}

	http.HandleFunc("/", hello)
	http.HandleFunc("/callback", callBack)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), nil))

}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello"))
	w.WriteHeader(200)
}

func (app *KitchenSink) callBack(w http.ResponseWriter, req *http.Request) {
	events, err := app.bot.ParseRequest(req)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
			fmt.Fprint(w, "ok")
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:

				// extract message details
				fmt.Printf("reply channel:%s, msg:%s, user id:%s\n", event.ReplyToken, message.Text,
					event.Source.UserID)

				// reply
				// if _, err = bot.ReplyMessage(event.ReplyToken,
				// 	linebot.NewTextMessage(message.Text)).Do(); err != nil {
				// 	log.Print(err)
				// }
				// if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Halooo!")).Do(); err != nil {
				// 	log.Print(err)
				// }
				if err := app.bot.handleText(message, event.ReplyToken, event.Source); err != nil {
					log.Print(err)
				}
			}
		}

	}
	return
}

func (app *KitchenSink) handleText(message *linebot.TextMessage, replyToken string, source *linebot.EventSource) error {
	switch message.Text {
	case "profile":
		return app.bot.replyText(replyToken, "Bot can't use profile API without user ID")
	}
}
