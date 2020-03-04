package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client

func main() {
	// load configuration
	var err error
	bot, err = linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
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

func callBack(w http.ResponseWriter, req *http.Request) {
	events, err := bot.ParseRequest(req)

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
			message := event.Message.Message()
			if message == "help" {
				// extract message details
				fmt.Printf("reply channel:%s, msg:%s, user id:%s\n", event.ReplyToken, message.Text, event.Source.UserID)

				// reply
				// if _, err = bot.ReplyMessage(event.ReplyToken,
				// 	linebot.NewTextMessage(message.Text)).Do(); err != nil {
				// 	log.Print(err)
				// }
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Halooo!")).Do(); err != nil {
					log.Print(err)
				}
			}
		}

	}
	return
}
