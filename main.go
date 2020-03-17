package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	app, err := NewTamakoBot(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
		os.Getenv("APP_BASE_URL"),
	)
	if err != nil {
		log.Fatal(err)
	}

	// serve /static/** files
	staticFileServer := http.FileServer(http.Dir("static"))
	http.HandleFunc("/static/", http.StripPrefix("/static/", staticFileServer).ServeHTTP)
	// serve /downloaded/** files
	downloadedFileServer := http.FileServer(http.Dir(app.downloadDir))
	http.HandleFunc("/downloaded/", http.StripPrefix("/downloaded/", downloadedFileServer).ServeHTTP)

	http.HandleFunc("/callback", app.Callback)
	// This is just a sample code.
	// For actually use, you must support HTTPS by using `ListenAndServeTLS`, reverse proxy or etc.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}

}

// TamakoBot app
type TamakoBot struct {
	bot         *linebot.Client
	appBaseURL  string
	downloadDir string
}

// NewTamakoBot function
func NewTamakoBot(channelSecret, channelToken, appBaseURL string) (*TamakoBot, error) {

	apiEndpointBase := os.Getenv("ENDPOINT_BASE")
	if apiEndpointBase == "" {
		apiEndpointBase = linebot.APIEndpointBase
	}
	fmt.Println(apiEndpointBase)
	bot, err := linebot.New(
		channelSecret,
		channelToken,
		linebot.WithEndpointBase(apiEndpointBase), // Usually you omit this.
	)
	if err != nil {
		return nil, err
	}
	downloadDir := filepath.Join(filepath.Dir(os.Args[0]), "line-bot")
	_, err = os.Stat(downloadDir)
	if err != nil {
		if err := os.Mkdir(downloadDir, 0777); err != nil {
			return nil, err
		}
	}
	return &TamakoBot{
		bot:         bot,
		appBaseURL:  appBaseURL,
		downloadDir: downloadDir,
	}, nil
}

// Callback function for http server
func (app *TamakoBot) Callback(w http.ResponseWriter, r *http.Request) {
	events, err := app.bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		log.Printf("Got event %v", event)
		switch event.Type {
		case linebot.EventTypeMessage:
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if err := app.handleText(message, event.ReplyToken, event.Source); err != nil {
					log.Print(err)
				}
			case *linebot.ImageMessage:
				if err := app.handleImage(message, event.ReplyToken); err != nil {
					log.Print(err)
				}
			case *linebot.VideoMessage:
				if err := app.handleVideo(message, event.ReplyToken); err != nil {
					log.Print(err)
				}
			case *linebot.AudioMessage:
				if err := app.handleAudio(message, event.ReplyToken); err != nil {
					log.Print(err)
				}
			case *linebot.FileMessage:
				if err := app.handleFile(message, event.ReplyToken); err != nil {
					log.Print(err)
				}
			case *linebot.LocationMessage:
				if err := app.handleLocation(message, event.ReplyToken); err != nil {
					log.Print(err)
				}
			case *linebot.StickerMessage:
				if err := app.handleSticker(message, event.ReplyToken); err != nil {
					log.Print(err)
				}
			default:
				log.Printf("Unknown message: %v", message)
			}
		case linebot.EventTypeFollow:
			if err := app.replyText(event.ReplyToken, "Got followed event"); err != nil {
				log.Print(err)
			}
		case linebot.EventTypeUnfollow:
			log.Printf("Unfollowed this bot: %v", event)
		case linebot.EventTypeJoin:
			if err := app.replyText(event.ReplyToken, "Joined "+string(event.Source.Type)); err != nil {
				log.Print(err)
			}
		case linebot.EventTypeLeave:
			log.Printf("Left: %v", event)
		case linebot.EventTypePostback:
			data := event.Postback.Data
			if data == "dmr" {
				song := "https://sites.google.com/site/untukaudio1/directory/Dramatic%20Market%20Ride.m4a"
				if _, err := app.bot.ReplyMessage(
					event.ReplyToken,
					linebot.NewAudioMessage(song, 100),
					linebot.NewTextMessage("Tamako Market Opening Song \nPerformed by : \nKitashirakawa Tamako (CV: Suzaki Aya)"),
				).Do(); err != nil {
					log.Print(err)
				}
			}
			if data == "neguse" {
				song := "https://sites.google.com/site/untukaudio1/directory/Neguse.m4a"
				if _, err := app.bot.ReplyMessage(
					event.ReplyToken,
					linebot.NewAudioMessage(song, 100),
					linebot.NewTextMessage("Tamako Market Ending Song \nPerformed by : \nKitashirakawa Tamako (CV: Suzaki Aya)"),
				).Do(); err != nil {
					log.Print(err)
				}
			}
			if data == "principle" {
				song := "https://sites.google.com/site/untukaudio1/directory/principle%20half.m4a"
				if _, err := app.bot.ReplyMessage(
					event.ReplyToken,
					linebot.NewAudioMessage(song, 100),
					linebot.NewTextMessage("Tamako Love Story\nPerformed by : \nKitashirakawa Tamako (CV: Suzaki Aya)"),
				).Do(); err != nil {
					log.Print(err)
				}
			}
			if data == "koinouta" {
				song := "https://sites.google.com/site/untukaudio1/directory/koinouta%20half.m4a"
				if _, err := app.bot.ReplyMessage(
					event.ReplyToken,
					linebot.NewAudioMessage(song, 100),
					linebot.NewTextMessage("Tamako Love Story Insert Song \nPerformed by : \nKitashirakawa Tamako (CV: Suzaki Aya)"),
				).Do(); err != nil {
					log.Print(err)
				}
			}
			if data == "DATE" || data == "TIME" || data == "DATETIME" {
				data += fmt.Sprintf("(%v)", *event.Postback.Params)
			}
			if err := app.replyText(event.ReplyToken, "Got postback: "+data); err != nil {
				log.Print(err)
			}
		case linebot.EventTypeBeacon:
			if err := app.replyText(event.ReplyToken, "Got beacon: "+event.Beacon.Hwid); err != nil {
				log.Print(err)
			}
		default:
			log.Printf("Unknown event: %v", event)
		}
	}
}

func (app *TamakoBot) handleText(message *linebot.TextMessage, replyToken string, source *linebot.EventSource) error {
	prefix := "!"
	lowerText := strings.ToLower(message.Text)
	if strings.HasPrefix(lowerText, prefix) {
		keyword := string(lowerText[1:])
		arg1 := strings.Split(keyword, " ")
		//arg2 := arg1[1]
		switch arg1[0] {
		case "help":
			profile, err := app.bot.GetProfile(source.UserID).Do()
			if err != nil {
				return app.replyText(replyToken, "add me as a friend")
			}

			var help = "Hello " + profile.DisplayName + ", nice to meet you ^_^ \nKeywords: help, sing, about, write, dota, games, manga, motw, ynm, chs, osu, steam, urban, lovecalc, anime, weather, stalk, music, youtubemp3, yt-dl, usage, leave.\n\n	For help type : \n!usage <available keyword>"
			if _, err := app.bot.ReplyMessage(replyToken, linebot.NewTextMessage(help)).Do(); err != nil {
				return err
			}
		case "sing":
			imageURL := "https://s-media-cache-ak0.pinimg.com/564x/9e/fa/18/9efa18b56cd5057101bf72a0b023ad7f.jpg"
			template := linebot.NewButtonsTemplate(
				imageURL, "Choose Tamako Song", "CV : Suzaki Aya",
				linebot.NewPostbackAction("Dramatic Market Ride", "dmr", "", ""),
				linebot.NewPostbackAction("Principle", "principle", "", ""),
				linebot.NewPostbackAction("Koi no Uta", "koinouta", "", ""),
				linebot.NewPostbackAction("Neguse", "neguse", "", ""),
			)
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewTemplateMessage("Song List", template),
			).Do(); err != nil {
				return err
			}
		case "about":
			imageURL := "https://01d54fec-a-62cb3a1a-s-sites.googlegroups.com/site/untukaudio1/directory/tamakomem.jpg"
			template := linebot.NewButtonsTemplate(
				imageURL, "About Developer", "Mr. Rojokundo",
				linebot.NewURIAction("Youtube", "https://www.youtube.com/rojofactory"),
				linebot.NewURIAction("Line@", "https://line.me/R/ti/p/%40wfq6948b"),
				linebot.NewURIAction("Instagram", "https://www.instagram.com/afifmakarim88"),
			)
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewTemplateMessage("About Developer", template),
			).Do(); err != nil {
				return err
			}
		case "write":
			sentence := string(keyword[5:])
			rawEncoded := Rawurlencode(sentence)
			var imageUrl string

			if rawEncoded == "" {
				return app.replyText(replyToken, "Nothing to write")
			}

			if len(rawEncoded) >= 8 && len(rawEncoded) <= 55 {
				imageUrl = "https://res.cloudinary.com/dftovjqdo/image/upload/a_-27,g_west,l_text:dark_name:" + rawEncoded + ",w_450,x_280,y_100/anime_notebook_yhekwa.jpg"
			} else if len(rawEncoded) <= 8 && len(rawEncoded) <= 55 {
				imageUrl = "https://res.cloudinary.com/dftovjqdo/image/upload/a_-27,g_west,l_text:dark_name:" + rawEncoded + ",w_200,x_250,y_100/anime_notebook_yhekwa.jpg"
			} else {
				return app.replyText(replyToken, "Text too long :(")
			}

			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewImageMessage(imageUrl, imageUrl),
			).Do(); err != nil {
				return err
			}
		case "dota":
			username := arg1[1]
			if err := app.dotaMessage(username, replyToken); err != nil {
				log.Print(err)
			}

		case "games":
			gamesKeyword := string(keyword[6:])
			if err := app.gameMessage(gamesKeyword, replyToken); err != nil {
				log.Print(err)
			}
		case "manga":
			mangaKeyword := string(keyword[6:])
			if err := app.mangaMessage(mangaKeyword, replyToken); err != nil {
				log.Print(err)
			}
		case "motw":

			if err := app.motwMessage(replyToken); err != nil {
				log.Print(err)
			}

		case "ynm":
			ynmKeyword := string(keyword[4:])
			array := []string{"Yes", "No", "Maybe"}
			rand.Seed(time.Now().UnixNano())
			randomInt := randomInt(0, 3)
			return app.replyText(replyToken, ynmKeyword+"\n"+array[randomInt])
		case "chs":
			chsKeyword := string(keyword[4:])
			explode := strings.Split(chsKeyword, "-")
			rand.Seed(time.Now().UnixNano())
			randomInt := randomInt(0, len(explode))
			random_str := explode[randomInt]
			return app.replyText(replyToken, "I choose "+random_str)
		case "osu":
			osuUsername := string(keyword[4:])
			if err := app.osuMessage(osuUsername, replyToken); err != nil {
				log.Print(err)
			}
		case "steam":
			steamUsername := string(keyword[6:])
			if err := app.steamMessage(steamUsername, replyToken); err != nil {
				log.Print(err)
			}
		case "bye":
			switch source.Type {
			case linebot.EventSourceTypeUser:
				return app.replyText(replyToken, "Bot can't leave from 1:1 chat")
			case linebot.EventSourceTypeGroup:
				if err := app.replyText(replyToken, "Leaving group"); err != nil {
					return err
				}
				if _, err := app.bot.LeaveGroup(source.GroupID).Do(); err != nil {
					return app.replyText(replyToken, err.Error())
				}
			case linebot.EventSourceTypeRoom:
				if err := app.replyText(replyToken, "Leaving room"); err != nil {
					return err
				}
				if _, err := app.bot.LeaveRoom(source.RoomID).Do(); err != nil {
					return app.replyText(replyToken, err.Error())
				}
			}
		default:
			log.Printf("Echo message to %s: %s", replyToken, message.Text)
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewTextMessage(message.Text),
			).Do(); err != nil {
				return err
			}
		}

	}
	return nil
}
func (app *TamakoBot) handleImage(message *linebot.ImageMessage, replyToken string) error {
	return app.handleHeavyContent(message.ID, func(originalContent *os.File) error {
		// You need to install ImageMagick.
		// And you should consider about security and scalability.
		previewImagePath := originalContent.Name() + "-preview"
		_, err := exec.Command("convert", "-resize", "240x", "jpeg:"+originalContent.Name(), "jpeg:"+previewImagePath).Output()
		if err != nil {
			return err
		}

		originalContentURL := app.appBaseURL + "/downloaded/" + filepath.Base(originalContent.Name())
		previewImageURL := app.appBaseURL + "/downloaded/" + filepath.Base(previewImagePath)
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewImageMessage(originalContentURL, previewImageURL),
		).Do(); err != nil {
			return err
		}
		return nil
	})
}

func (app *TamakoBot) handleVideo(message *linebot.VideoMessage, replyToken string) error {
	return app.handleHeavyContent(message.ID, func(originalContent *os.File) error {
		// You need to install FFmpeg and ImageMagick.
		// And you should consider about security and scalability.
		previewImagePath := originalContent.Name() + "-preview"
		_, err := exec.Command("convert", "mp4:"+originalContent.Name()+"[0]", "jpeg:"+previewImagePath).Output()
		if err != nil {
			return err
		}

		originalContentURL := app.appBaseURL + "/downloaded/" + filepath.Base(originalContent.Name())
		previewImageURL := app.appBaseURL + "/downloaded/" + filepath.Base(previewImagePath)
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewVideoMessage(originalContentURL, previewImageURL),
		).Do(); err != nil {
			return err
		}
		return nil
	})
}

func (app *TamakoBot) handleAudio(message *linebot.AudioMessage, replyToken string) error {
	return app.handleHeavyContent(message.ID, func(originalContent *os.File) error {
		originalContentURL := app.appBaseURL + "/downloaded/" + filepath.Base(originalContent.Name())
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewAudioMessage(originalContentURL, 100),
		).Do(); err != nil {
			return err
		}
		return nil
	})
}

func (app *TamakoBot) handleFile(message *linebot.FileMessage, replyToken string) error {
	return app.replyText(replyToken, fmt.Sprintf("File `%s` (%d bytes) received.", message.FileName, message.FileSize))
}

func (app *TamakoBot) handleLocation(message *linebot.LocationMessage, replyToken string) error {
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewLocationMessage(message.Title, message.Address, message.Latitude, message.Longitude),
	).Do(); err != nil {
		return err
	}
	return nil
}

func (app *TamakoBot) handleSticker(message *linebot.StickerMessage, replyToken string) error {
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewStickerMessage(message.PackageID, message.StickerID),
	).Do(); err != nil {
		return err
	}
	return nil
}

// func joinTag(article string )[]string {
// 	tags := strings.Split(article, ",")
//     return tags
// }

func defaultValue(message string) string {

	if message == "" {
		return "-"
	}
	potong := strings.ReplaceAll(message, "\n", "")
	potongLagi := strings.ReplaceAll(potong, `"`, "")
	potonglagilagi := strings.ReplaceAll(potongLagi, "\r", "")
	return potonglagilagi
}

func defaultImage(message string) string {
	if message == "" {
		return "https://forum.dbaclass.com/wp-content/themes/qaengine/img/default-thumbnail.jpg"
	}
	return message
}

func (app *TamakoBot) osuMessage(message string, replyToken string) error {

	var osuStd []OsuStd
	var osuMania []OsuMania
	var osuTaiko []OsuTaiko
	var osuCtb []OsuCtb

	// get osu standard api
	stdApi := getData("https://osu.ppy.sh/api/get_user?u=" + message + "&m=0&k=1958afa9967f399f1cd22f52be34d93bcf755212")
	json.Unmarshal([]byte(stdApi), &osuStd)
	stdAkurasi := defaultValue(osuStd[0].Accuracy)
	stdCountryRank := defaultValue(osuStd[0].Pp_country_rank)
	stdGlobalRank := defaultValue(osuStd[0].Pp_rank)

	username := osuStd[0].Username
	imageUrl := "https://a.ppy.sh/" + osuStd[0].User_id

	if len(username) == 0 || message == "" {
		return app.replyText(replyToken, "osu information not found")
	}

	maniaApi := getData("https://osu.ppy.sh/api/get_user?u=" + message + "&m=3&k=1958afa9967f399f1cd22f52be34d93bcf755212")
	json.Unmarshal([]byte(maniaApi), &osuMania)
	maniaAkurasi := defaultValue(osuMania[0].Accuracy)
	maniaCountryRank := defaultValue(osuMania[0].Pp_country_rank)
	maniaGlobalRank := defaultValue(osuMania[0].Pp_rank)

	taikoApi := getData("https://osu.ppy.sh/api/get_user?u=" + message + "&m=1&k=1958afa9967f399f1cd22f52be34d93bcf755212")
	json.Unmarshal([]byte(taikoApi), &osuTaiko)
	taikoAkurasi := defaultValue(osuTaiko[0].Accuracy)
	taikoCountryRank := defaultValue(osuTaiko[0].Pp_country_rank)
	taikoGlobalRank := defaultValue(osuTaiko[0].Pp_rank)

	ctbApi := getData("https://osu.ppy.sh/api/get_user?u=" + message + "&m=2&k=1958afa9967f399f1cd22f52be34d93bcf755212")
	json.Unmarshal([]byte(ctbApi), &osuCtb)
	ctbAkurasi := defaultValue(osuCtb[0].Accuracy)
	ctbCountryRank := defaultValue(osuCtb[0].Pp_country_rank)
	ctbGlobalRank := defaultValue(osuCtb[0].Pp_rank)

	jsonString := `{
		"type": "carousel",
		"contents": [
		  {
			"type": "bubble",
			"body": {
			  "type": "box",
			  "layout": "vertical",
			  "contents": [
				{
				  "type": "text",
				  "text": "osu!",
				  "weight": "bold",
				  "color": "#dc98a4",
				  "size": "sm"
				},
				{
				  "type": "text",
				  "text": "` + username + `",
				  "weight": "bold",
				  "size": "xxl",
				  "margin": "md"
				},
				{
				  "type": "text",
				  "text": "Miraina Tower, 4-1-6 Shinjuku, Tokyo",
				  "size": "xs",
				  "color": "#aaaaaa",
				  "wrap": true
				},
				{
				  "type": "separator",
				  "margin": "xxl"
				},
				{
				  "type": "box",
				  "layout": "vertical",
				  "margin": "xxl",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Standard",
						  "size": "md",
						  "color": "#555555",
						  "flex": 0,
						  "weight": "bold"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Country Rank",
						  "size": "sm",
						  "color": "#555555",
						  "flex": 0
						},
						{
						  "type": "text",
						  "text": "` + stdCountryRank + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Global Rank",
						  "size": "sm",
						  "color": "#555555",
						  "flex": 0
						},
						{
						  "type": "text",
						  "text": "` + stdGlobalRank + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Accuracy",
						  "size": "sm",
						  "color": "#555555",
						  "flex": 0
						},
						{
						  "type": "text",
						  "text": "` + stdAkurasi + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "separator",
					  "margin": "xxl"
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Mania",
						  "size": "md",
						  "color": "#555555",
						  "flex": 0,
						  "weight": "bold"
						}
					  ],
					  "margin": "xxl"
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Country Rank",
						  "size": "sm",
						  "color": "#555555"
						},
						{
						  "type": "text",
						  "text": "` + maniaCountryRank + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Global Rank",
						  "size": "sm",
						  "color": "#555555"
						},
						{
						  "type": "text",
						  "text": "` + maniaGlobalRank + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Accuracy",
						  "size": "sm",
						  "color": "#555555"
						},
						{
						  "type": "text",
						  "text": "` + maniaAkurasi + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					}
				  ]
				},
				{
				  "type": "separator",
				  "margin": "xxl"
				},
				{
				  "type": "box",
				  "layout": "horizontal",
				  "margin": "md",
				  "contents": [
					{
					  "type": "text",
					  "text": "Country",
					  "size": "xs",
					  "color": "#aaaaaa",
					  "flex": 0
					},
					{
					  "type": "text",
					  "text": "aaaa",
					  "color": "#aaaaaa",
					  "size": "xs",
					  "align": "end"
					}
				  ]
				}
			  ]
			},
			"styles": {
			  "footer": {
				"separator": true
			  }
			}
		  },
		  {
			"type": "bubble",
			"body": {
			  "type": "box",
			  "layout": "vertical",
			  "contents": [{
				"type": "box",
				"layout": "vertical",
				"contents": [
				  {
					"type": "image",
					"url": "` + imageUrl + `"
				  }
				]
			  },
				{
				  "type": "separator",
				  "margin": "xxl"
				},
				{
				  "type": "box",
				  "layout": "vertical",
				  "margin": "xxl",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Taiko",
						  "size": "md",
						  "color": "#555555",
						  "flex": 0,
						  "weight": "bold"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Country Rank",
						  "size": "sm",
						  "color": "#555555",
						  "flex": 0
						},
						{
						  "type": "text",
						  "text": "` + taikoCountryRank + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Global Rank",
						  "size": "sm",
						  "color": "#555555",
						  "flex": 0
						},
						{
						  "type": "text",
						  "text": "` + taikoGlobalRank + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Accuracy",
						  "size": "sm",
						  "color": "#555555",
						  "flex": 0
						},
						{
						  "type": "text",
						  "text": "` + taikoAkurasi + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "separator",
					  "margin": "xxl"
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Catch the beat",
						  "size": "md",
						  "color": "#555555",
						  "flex": 0,
						  "weight": "bold"
						}
					  ],
					  "margin": "xxl"
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Country Rank",
						  "size": "sm",
						  "color": "#555555"
						},
						{
						  "type": "text",
						  "text": "` + ctbCountryRank + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Global Rank",
						  "size": "sm",
						  "color": "#555555"
						},
						{
						  "type": "text",
						  "text": "` + ctbGlobalRank + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					},
					{
					  "type": "box",
					  "layout": "horizontal",
					  "contents": [
						{
						  "type": "text",
						  "text": "Accuracy",
						  "size": "sm",
						  "color": "#555555"
						},
						{
						  "type": "text",
						  "text": "` + ctbAkurasi + `",
						  "size": "sm",
						  "color": "#111111",
						  "align": "end"
						}
					  ]
					}
				  ]
				},
				{
				  "type": "separator",
				  "margin": "xxl"
				}
			  ]
			},
			"styles": {
			  "footer": {
				"separator": true
			  }
			}
		  }
		]
	  }`
	contents, err := linebot.UnmarshalFlexMessageJSON([]byte(jsonString))
	if err != nil {
		return err
	}
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewFlexMessage("osu! information", contents),
	).Do(); err != nil {
		return err
	}

	return nil

}

func (app *TamakoBot) motwMessage(replyToken string) error {
	var motwApi MotwApi

	motw := getData("https://rss.itunes.apple.com/api/v1/id/apple-music/top-songs/all/10/explicit.json")
	json.Unmarshal([]byte(motw), &motwApi)
	//var id []string
	var columns []*linebot.CarouselColumn
	//var actions []linebot.TemplateAction
	// Add Actions

	for _, details := range motwApi.Feed.Results {
		imgUrl := details.ArtworkUrl100
		artistName := details.ArtistName
		titleName := details.Name
		id := details.Id
		columns = append(columns, linebot.NewCarouselColumn(imgUrl, artistName, titleName, linebot.NewMessageAction("Preview", "!details "+id)))

	}

	template := linebot.NewCarouselTemplate(columns...)
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewTemplateMessage("Music of the week", template),
	).Do(); err != nil {
		return err
	}
	return nil
	// return app.replyText(replyToken, "Video game information not found")
}

func (app *TamakoBot) gameMessage(message string, replyToken string) error {
	var gameList GameList
	queryGame := Rawurlencode(message)
	url := "https://www.giantbomb.com/api/search/?api_key=a0bede1760f86f2f59ff3ac477c953fed643ea0b&resources=game&query=" + queryGame + "&format=json&limit=5"
	gameApi := Request(url, "lashaparesha api script")
	json.Unmarshal([]byte(gameApi), &gameList)

	var jsonString string

	hitung := len(gameList.Results)
	if hitung > 0 {
		result := make([]string, hitung)

		// exe := string(runes[hitung:]) // hapus comma

		for _, details := range gameList.Results {

			title := defaultValue(details.Name)
			release_date := defaultValue(details.Original_release_date)
			small_url := defaultImage(details.Image.Small_url)
			deck := defaultValue(details.Deck)
			link := details.Site_detail_url

			platforms := []string{}

			for _, plat := range details.Platforms {
				platform := plat.Name
				platforms = append(platforms, platform)
			}

			joinPlat := strings.Join(platforms, ", ")

			jsonString = `{
		  "type": "bubble",
		  "hero": {
			"type": "image",
			"url": "` + small_url + `",
			"size": "full",
			"aspectRatio": "8:9",
			"aspectMode": "cover"
		  },
		  "body": {
			"type": "box",
			"layout": "vertical",
			"contents": [
			  {
				"type": "text",
				"text": "` + title + `",
				"weight": "bold",
				"size": "xl",
				"wrap": true
			  },
			  {
				"type": "box",
				"layout": "vertical",
				"margin": "lg",
				"spacing": "sm",
				"contents": [
				  {
					"type": "box",
					"layout": "baseline",
					"spacing": "sm",
					"contents": [
					  {
						"type": "text",
						"text": "Release Date",
						"color": "#aaaaaa",
						"size": "sm",
						"flex": 3,
						"wrap": true
					  },
					  {
						"type": "text",
						"text": "` + release_date + `",
						"wrap": true,
						"color": "#666666",
						"size": "sm",
						"flex": 5
					  }
					]
				  },
				  {
					"type": "box",
					"layout": "baseline",
					"spacing": "sm",
					"contents": [
					  {
						"type": "text",
						"text": "Platform",
						"color": "#aaaaaa",
						"size": "sm",
						"flex": 3
					  },
					  {
						"type": "text",
						"text": "` + joinPlat + `",
						"wrap": true,
						"color": "#666666",
						"size": "sm",
						"flex": 5
					  }
					]
				  }
				]
			  },
			  {
				"type": "box",
				"layout": "vertical",
				"contents": [
				  {
					"type": "text",
					"text": "Description",
					"weight": "bold",
					"size": "sm"
				  },
				  {
					"type": "box",
					"layout": "vertical",
					"contents": [
					  {
						"type": "text",
						"text": "` + deck + `",
						"margin": "lg",
						"size": "sm",
						"wrap": true
					  }
					],
					"paddingTop": "5px"
				  }
				],
				"margin": "xl",
				"cornerRadius": "2px"
			  }
			],
			"backgroundColor": "#aaaaaa"
		  },
		  "footer": {
			"type": "box",
			"layout": "vertical",
			"spacing": "sm",
			"contents": [
			  {
				"type": "button",
				"style": "link",
				"height": "sm",
				"action": {
				  "type": "uri",
				  "label": "Open Browser",
				  "uri": "` + link + `"
				}
			  },
			  {
				"type": "spacer",
				"size": "sm"
			  }
			],
			"flex": 0
		  }
		}`

			result = append(result, jsonString)
		}

		joinString := strings.Join(result, ",") // join string
		runes := []rune(joinString)
		exe := string(runes[hitung:]) // hapus comma
		resultz := fmt.Sprintf(`{ 
		"type": "carousel",
		"contents": [%s]
	  }`, exe) // gabung json string ke type carousel

		fmt.Print("WADOOOHHH" + exe)
		fmt.Println("BAAASATTTT: " + url)
		contents, err := linebot.UnmarshalFlexMessageJSON([]byte(resultz))
		if err != nil {
			return err
		}
		if _, err := app.bot.ReplyMessage(
			replyToken,
			linebot.NewFlexMessage("Flex message alt text", contents),
		).Do(); err != nil {
			return err
		}
	} else {
		return app.replyText(replyToken, "Video game information not found")
	}
	return nil
}

func (app *TamakoBot) dotaMessage(message string, replyToken string) error {

	var steam Steam
	var dotaProfile DotaProfile
	var dotaWinrate DotaWinrate
	var signatureHero []DotaHero
	var recentMatch []DotaMatch

	if message == "" {
		return app.replyText(replyToken, "Dota 2 information not found")
	}

	// Get 64bit SteamId
	steamJson := getData("https://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key=7834436769DDB41F2D14A2F312377946&vanityurl=" + message)
	json.Unmarshal([]byte(steamJson), &steam)
	steam_64 := convert32bit(steam.Response.Steamid)

	// Get Dota 2 Player Profile
	get_info_dota := getData("https://api.opendota.com/api/players/" + steam_64)
	json.Unmarshal([]byte(get_info_dota), &dotaProfile)

	// Get Dota 2 Win Rate
	get_winrate := getData("https://api.opendota.com/api/players/" + steam_64 + "/wl")
	json.Unmarshal([]byte(get_winrate), &dotaWinrate)
	win := strconv.Itoa(dotaWinrate.Win)
	// lose := strconv.Itoa(dotaWinrate.Lose)
	totalMatch := strconv.Itoa(dotaWinrate.Win + dotaWinrate.Lose)

	// Get Dota 2 Signature Hero
	get_signature_hero := getData("https://api.opendota.com/api/players/" + steam_64 + "/heroes")
	json.Unmarshal([]byte(get_signature_hero), &signatureHero)
	signature_hero := hero_id_to_names(signatureHero[0].Hero_id)

	// Get Dota 2 Recent Match
	get_recent_match := getData("https://api.opendota.com/api/players/" + steam_64 + "/recentMatches")
	json.Unmarshal([]byte(get_recent_match), &recentMatch)
	matchId := "https://www.dotabuff.com/matches/" + strconv.Itoa(recentMatch[0].Match_id)
	hero := "Hero : " + hero_id_to_names(strconv.Itoa(recentMatch[0].Hero_id))
	kda := "K/D/A : " + strconv.Itoa(recentMatch[0].Kills) + "/" + strconv.Itoa(recentMatch[0].Deaths) + "/" + strconv.Itoa(recentMatch[0].Assists)
	lh_gpm := "LH/GPM : " + strconv.Itoa(recentMatch[0].Last_hits) + "/" + strconv.Itoa(recentMatch[0].Gold_per_min)

	steamUrl := "https://steamcommunity.com/id/" + message

	imageURL := dotaProfile.Profile.Avatarfull
	template := linebot.NewCarouselTemplate(
		linebot.NewCarouselColumn(
			imageURL, dotaProfile.Profile.Personaname, "Win : "+win+"\nTotal Match : "+totalMatch+"\nSignature Hero : "+signature_hero,
			linebot.NewURIAction("Open Steam", steamUrl),
		),
		linebot.NewCarouselColumn(
			imageURL, "Recent Match Played", hero+"\n"+kda+"\n"+lh_gpm,
			linebot.NewURIAction("Open Dotabuff", matchId),
		),
	)

	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewTemplateMessage("Info Dota", template),
	).Do(); err != nil {
		return err
	}
	return nil
}

func (app *TamakoBot) steamMessage(message string, replyToken string) error {
	var steam Steam
	var gameCount GameCount
	//var steamProfile SteamProfile
	steamJson := getData("https://api.steampowered.com/ISteamUser/ResolveVanityURL/v0001/?key=7834436769DDB41F2D14A2F312377946&vanityurl=" + message)
	json.Unmarshal([]byte(steamJson), &steam)
	steam_64 := convert32bit(steam.Response.Steamid)

	getGameCount := getData("http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=7834436769DDB41F2D14A2F312377946&steamid=" + steam_64 + "&format=json")
	json.Unmarshal([]byte(getGameCount), &gameCount)

	return app.replyText(replyToken, steam.Response.Steamid)
	//total_lib := gameCount.Game_count
	// 	total_lib := "xx"
	// 	getSteamProfile := getData("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=7834436769DDB41F2D14A2F312377946&steamids=" + steam_64)
	// 	json.Unmarshal([]byte(getSteamProfile), &steamProfile)
	// 	get_nickname := steamProfile.Players[0].Personaname
	// 	get_avatar := steamProfile.Players[0].Avatarfull
	// 	get_realname := steamProfile.Players[0].Personaname
	// 	//get_state := steamProfile.Players[0].Profilestate
	// 	get_state := "1"
	// 	jsonString := `{
	// 	"type": "bubble",
	// 	"body": {
	// 	  "type": "box",
	// 	  "layout": "vertical",
	// 	  "contents": [
	// 		{
	// 		  "type": "box",
	// 		  "layout": "horizontal",
	// 		  "contents": [
	// 			{
	// 			  "type": "box",
	// 			  "layout": "vertical",
	// 			  "contents": [
	// 				{
	// 				  "type": "image",
	// 				  "url": "` + get_avatar + `",
	// 				  "aspectMode": "cover",
	// 				  "size": "full"
	// 				}
	// 			  ],
	// 			  "cornerRadius": "100px",
	// 			  "width": "72px",
	// 			  "height": "72px"
	// 			},
	// 			{
	// 			  "type": "box",
	// 			  "layout": "vertical",
	// 			  "contents": [
	// 				{
	// 				  "type": "text",
	// 				  "contents": [
	// 					{
	// 					  "type": "span",
	// 					  "text": "` + get_nickname + `",
	// 					  "size": "lg",
	// 					  "weight": "bold"
	// 					}
	// 				  ],
	// 				  "size": "sm",
	// 				  "wrap": true
	// 				},
	// 				{
	// 				  "type": "box",
	// 				  "layout": "baseline",
	// 				  "contents": [
	// 					{
	// 					  "type": "text",
	// 					  "text": "` + get_realname + `",
	// 					  "size": "sm",
	// 					  "color": "#bcbcbc"
	// 					}
	// 				  ],
	// 				  "spacing": "sm",
	// 				  "margin": "md"
	// 				},
	// 				{
	// 				  "type": "box",
	// 				  "layout": "baseline",
	// 				  "contents": [
	// 					{
	// 					  "type": "text",
	// 					  "text": "` + get_state + `",
	// 					  "size": "sm",
	// 					  "color": "#bcbcbc"
	// 					}
	// 				  ],
	// 				  "spacing": "sm",
	// 				  "margin": "md"
	// 				}
	// 			  ]
	// 			}
	// 		  ],
	// 		  "spacing": "xl",
	// 		  "paddingAll": "20px"
	// 		},
	// 		{
	// 		  "type": "box",
	// 		  "layout": "baseline",
	// 		  "spacing": "sm",
	// 		  "contents": [
	// 			{
	// 			  "type": "text",
	// 			  "text": "Library",
	// 			  "color": "#aaaaaa",
	// 			  "size": "sm",
	// 			  "flex": 3
	// 			},
	// 			{
	// 			  "type": "text",
	// 			  "text": "` + total_lib + `",
	// 			  "wrap": true,
	// 			  "color": "#666666",
	// 			  "size": "sm",
	// 			  "flex": 5
	// 			}
	// 		  ]
	// 		},
	// 		{
	// 		  "type": "box",
	// 		  "layout": "baseline",
	// 		  "spacing": "sm",
	// 		  "contents": [
	// 			{
	// 			  "type": "text",
	// 			  "text": "Recent",
	// 			  "color": "#aaaaaa",
	// 			  "size": "sm",
	// 			  "flex": 3
	// 			},
	// 			{
	// 			  "type": "text",
	// 			  "text": "Miraina Tower, 4-1-6 Shinjuku, Tokyo",
	// 			  "wrap": true,
	// 			  "color": "#666666",
	// 			  "size": "sm",
	// 			  "flex": 5
	// 			}
	// 		  ]
	// 		}
	// 	  ]
	// 	},
	// 	"footer": {
	// 	  "type": "box",
	// 	  "layout": "vertical",
	// 	  "contents": [
	// 		{
	// 		  "type": "button",
	// 		  "action": {
	// 			"type": "uri",
	// 			"label": "action",
	// 			"uri": "http://linecorp.com/"
	// 		  },
	// 		  "style": "primary"
	// 		}
	// 	  ]
	// 	}
	//   }`

	// 	contents, err := linebot.UnmarshalFlexMessageJSON([]byte(jsonString))
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if _, err := app.bot.ReplyMessage(
	// 		replyToken,
	// 		linebot.NewFlexMessage("Manga Information", contents),
	// 	).Do(); err != nil {
	// 		return err
	// 	}

	// 	return nil
}

func (app *TamakoBot) mangaMessage(message string, replyToken string) error {
	var getManga MangaApi
	var getGenre GenreApi

	queryManga := Rawurlencode(message)

	get_manga := getData("https://kitsu.io/api/edge/manga?filter[text]=" + queryManga + "&page[limit]=3&page[offset]=0")
	json.Unmarshal([]byte(get_manga), &getManga)

	result := []string{}

	if len(getManga.Data) == 0 || message == "" {
		return app.replyText(replyToken, "Manga information not found")
	}

	for _, details := range getManga.Data {
		title := defaultValue(details.Attributes.CanonicalTitle)
		image := details.Attributes.PosterImage.Medium
		status := defaultValue(details.Attributes.Status)
		rating := defaultValue(details.Attributes.AverageRating)
		synopsis := defaultValue(details.Attributes.Synopsis)

		get_genre_endpoint := details.Relationships.Genres.Links.Related
		get_genre := getData(get_genre_endpoint)
		json.Unmarshal([]byte(get_genre), &getGenre)

		genresArray := []string{}

		for i := len(getGenre.Data) - 1; i >= 1; i-- {
			genres := getGenre.Data[i].Attributes.Name
			genresArray = append(genresArray, genres)
		}
		join_genre := defaultValue(strings.Join(genresArray, ", "))

		// fmt.Println(join_genre)
		// fmt.Println(get_genre_endpoint)

		jsonString := `{
			"type": "bubble",
			"hero": {
			  "type": "image",
			  "url": "` + image + `",
			  "size": "full",
			  "aspectRatio": "7:9",
			  "aspectMode": "cover"
			},
			"body": {
			  "type": "box",
			  "layout": "vertical",
			  "contents": [
				{
				  "type": "text",
				  "text": "` + title + `",
				  "weight": "bold",
				  "size": "xl",
				  "wrap": true
				},
				{
				  "type": "box",
				  "layout": "vertical",
				  "margin": "lg",
				  "spacing": "sm",
				  "contents": [
					{
					  "type": "box",
					  "layout": "baseline",
					  "spacing": "sm",
					  "contents": [
						{
						  "type": "text",
						  "text": "Status",
						  "color": "#aaaaaa",
						  "size": "sm",
						  "flex": 3,
						  "wrap": true
						},
						{
						  "type": "text",
						  "text": "` + status + `",
						  "wrap": true,
						  "color": "#666666",
						  "size": "sm",
						  "flex": 5
						}
					  ]
					},{
						"type": "box",
						"layout": "baseline",
						"spacing": "sm",
						"contents": [
						  {
							"type": "text",
							"text": "Rating",
							"color": "#aaaaaa",
							"size": "sm",
							"flex": 3,
							"wrap": true
						  },
						  {
							"type": "text",
							"text": "` + rating + `",
							"wrap": true,
							"color": "#666666",
							"size": "sm",
							"flex": 5
						  }
						]
					  },
					{
					  "type": "box",
					  "layout": "baseline",
					  "spacing": "sm",
					  "contents": [
						{
						  "type": "text",
						  "text": "Genre",
						  "color": "#aaaaaa",
						  "size": "sm",
						  "flex": 3
						},
						{
						  "type": "text",
						  "text": "` + join_genre + `",
						  "wrap": true,
						  "color": "#666666",
						  "size": "sm",
						  "flex": 5
						}
					  ]
					}
				  ]
				},
				{
				  "type": "box",
				  "layout": "vertical",
				  "contents": [
					{
					  "type": "text",
					  "text": "Synopsis",
					  "weight": "bold",
					  "size": "sm"
					},
					{
					  "type": "box",
					  "layout": "vertical",
					  "contents": [
						{
						  "type": "text",
						  "text": "` + synopsis + `",
						  "margin": "lg",
						  "size": "xs",
						  "wrap": true,
						  "align": "center"
						}
					  ],
					  "paddingTop": "5px"
					}
				  ],
				  "margin": "xl",
				  "cornerRadius": "2px"
				}
			  ]
			},
			"footer": {
			  "type": "box",
			  "layout": "vertical",
			  "spacing": "sm",
			  "contents": [
				{
				  "type": "spacer",
				  "size": "sm"
				}
			  ],
			  "flex": 0
			}
		  }`
		result = append(result, jsonString)
	}
	join_string := strings.Join(result, ", ")
	result_carousel := fmt.Sprintf(`{ 
		"type": "carousel",
		"contents": [%s]
	  }`, join_string) // gabung json string ke type carousel

	contents, err := linebot.UnmarshalFlexMessageJSON([]byte(result_carousel))
	if err != nil {
		return err
	}
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewFlexMessage("Manga Information", contents),
	).Do(); err != nil {
		return err
	}

	return nil
}

func (app *TamakoBot) replyText(replyToken, text string) error {
	if _, err := app.bot.ReplyMessage(
		replyToken,
		linebot.NewTextMessage(text),
	).Do(); err != nil {
		return err
	}
	return nil
}

func (app *TamakoBot) handleHeavyContent(messageID string, callback func(*os.File) error) error {
	content, err := app.bot.GetMessageContent(messageID).Do()
	if err != nil {
		return err
	}
	defer content.Content.Close()
	log.Printf("Got file: %s", content.ContentType)
	originalConent, err := app.saveContent(content.Content)
	if err != nil {
		return err
	}
	return callback(originalConent)
}

func (app *TamakoBot) saveContent(content io.ReadCloser) (*os.File, error) {
	file, err := ioutil.TempFile(app.downloadDir, "")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	if err != nil {
		return nil, err
	}
	log.Printf("Saved %s", file.Name())
	return file, nil
}
