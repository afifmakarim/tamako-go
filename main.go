package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

func Rawurlencode(str string) string {
	return strings.Replace(url.QueryEscape(str), "+", "%20", -1)
}

func (app *TamakoBot) handleText(message *linebot.TextMessage, replyToken string, source *linebot.EventSource) error {
	prefix := "!"
	if strings.HasPrefix(message.Text, prefix) {
		keyword := string(message.Text[1:])
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
			sentence := string(keyword[4:])
			rawEncoded := Rawurlencode(string(sentence[1:]))
			var imageUrl string

			if sentence == "" {
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
		case "image carousel":
			imageURL := app.appBaseURL + "/static/buttons/1040.jpg"
			template := linebot.NewImageCarouselTemplate(
				linebot.NewImageCarouselColumn(
					imageURL,
					linebot.NewURIAction("Go to LINE", "https://line.me"),
				),
				linebot.NewImageCarouselColumn(
					imageURL,
					linebot.NewPostbackAction("Say hello1", "hello こんにちは", "", ""),
				),
				linebot.NewImageCarouselColumn(
					imageURL,
					linebot.NewMessageAction("Say message", "Rice=米"),
				),
				linebot.NewImageCarouselColumn(
					imageURL,
					linebot.NewDatetimePickerAction("datetime", "DATETIME", "datetime", "", "", ""),
				),
			)
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewTemplateMessage("Image carousel alt text", template),
			).Do(); err != nil {
				return err
			}
		case "datetime":
			template := linebot.NewButtonsTemplate(
				"", "", "Select date / time !",
				linebot.NewDatetimePickerAction("date", "DATE", "date", "", "", ""),
				linebot.NewDatetimePickerAction("time", "TIME", "time", "", "", ""),
				linebot.NewDatetimePickerAction("datetime", "DATETIME", "datetime", "", "", ""),
			)
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewTemplateMessage("Datetime pickers alt text", template),
			).Do(); err != nil {
				return err
			}
		case "flex":
			// {
			//   "type": "bubble",
			//   "body": {
			//     "type": "box",
			//     "layout": "horizontal",
			//     "contents": [
			//       {
			//         "type": "text",
			//         "text": "Hello,"
			//       },
			//       {
			//         "type": "text",
			//         "text": "World!"
			//       }
			//     ]
			//   }
			// }
			contents := &linebot.BubbleContainer{
				Type: linebot.FlexContainerTypeBubble,
				Body: &linebot.BoxComponent{
					Type:   linebot.FlexComponentTypeBox,
					Layout: linebot.FlexBoxLayoutTypeHorizontal,
					Contents: []linebot.FlexComponent{
						&linebot.TextComponent{
							Type: linebot.FlexComponentTypeText,
							Text: "Hello,",
						},
						&linebot.TextComponent{
							Type: linebot.FlexComponentTypeText,
							Text: "World!",
						},
					},
				},
			}
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewFlexMessage("Flex message alt text", contents),
			).Do(); err != nil {
				return err
			}
		case "flex carousel":
			// {
			//   "type": "carousel",
			//   "contents": [
			//     {
			//       "type": "bubble",
			//       "body": {
			//         "type": "box",
			//         "layout": "vertical",
			//         "contents": [
			//           {
			//             "type": "text",
			//             "text": "First bubble"
			//           }
			//         ]
			//       }
			//     },
			//     {
			//       "type": "bubble",
			//       "body": {
			//         "type": "box",
			//         "layout": "vertical",
			//         "contents": [
			//           {
			//             "type": "text",
			//             "text": "Second bubble"
			//           }
			//         ]
			//       }
			//     }
			//   ]
			// }
			contents := &linebot.CarouselContainer{
				Type: linebot.FlexContainerTypeCarousel,
				Contents: []*linebot.BubbleContainer{
					{
						Type: linebot.FlexContainerTypeBubble,
						Body: &linebot.BoxComponent{
							Type:   linebot.FlexComponentTypeBox,
							Layout: linebot.FlexBoxLayoutTypeVertical,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type: linebot.FlexComponentTypeText,
									Text: "First bubble",
								},
							},
						},
					},
					{
						Type: linebot.FlexContainerTypeBubble,
						Body: &linebot.BoxComponent{
							Type:   linebot.FlexComponentTypeBox,
							Layout: linebot.FlexBoxLayoutTypeVertical,
							Contents: []linebot.FlexComponent{
								&linebot.TextComponent{
									Type: linebot.FlexComponentTypeText,
									Text: "Second bubble",
								},
							},
						},
					},
				},
			}
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewFlexMessage("Flex message alt text", contents),
			).Do(); err != nil {
				return err
			}
		case "flex json":
			jsonString := `{
  "type": "bubble",
  "hero": {
    "type": "image",
    "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/01_1_cafe.png",
    "size": "full",
    "aspectRatio": "20:13",
    "aspectMode": "cover",
    "action": {
      "type": "uri",
      "uri": "http://linecorp.com/"
    }
  },
  "body": {
    "type": "box",
    "layout": "vertical",
    "contents": [
      {
        "type": "text",
        "text": "Brown Cafe",
        "weight": "bold",
        "size": "xl"
      },
      {
        "type": "box",
        "layout": "baseline",
        "margin": "md",
        "contents": [
          {
            "type": "icon",
            "size": "sm",
            "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
          },
          {
            "type": "icon",
            "size": "sm",
            "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
          },
          {
            "type": "icon",
            "size": "sm",
            "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
          },
          {
            "type": "icon",
            "size": "sm",
            "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gold_star_28.png"
          },
          {
            "type": "icon",
            "size": "sm",
            "url": "https://scdn.line-apps.com/n/channel_devcenter/img/fx/review_gray_star_28.png"
          },
          {
            "type": "text",
            "text": "4.0",
            "size": "sm",
            "color": "#999999",
            "margin": "md",
            "flex": 0
          }
        ]
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
                "text": "Place",
                "color": "#aaaaaa",
                "size": "sm",
                "flex": 1
              },
              {
                "type": "text",
                "text": "Miraina Tower, 4-1-6 Shinjuku, Tokyo",
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
                "text": "Time",
                "color": "#aaaaaa",
                "size": "sm",
                "flex": 1
              },
              {
                "type": "text",
                "text": "10:00 - 23:00",
                "wrap": true,
                "color": "#666666",
                "size": "sm",
                "flex": 5
              }
            ]
          }
        ]
      }
    ]
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
          "label": "CALL",
          "uri": "https://linecorp.com"
        }
      },
      {
        "type": "button",
        "style": "link",
        "height": "sm",
        "action": {
          "type": "uri",
          "label": "WEBSITE",
          "uri": "https://linecorp.com",
          "altUri": {
            "desktop": "https://line.me/ja/download"
          }
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
			contents, err := linebot.UnmarshalFlexMessageJSON([]byte(jsonString))
			if err != nil {
				return err
			}
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewFlexMessage("Flex message alt text", contents),
			).Do(); err != nil {
				return err
			}
		case "imagemap":
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewImagemapMessage(
					app.appBaseURL+"/static/rich",
					"Imagemap alt text",
					linebot.ImagemapBaseSize{Width: 1040, Height: 1040},
					linebot.NewURIImagemapAction("LINE Store Manga", "https://store.line.me/family/manga/en", linebot.ImagemapArea{X: 0, Y: 0, Width: 520, Height: 520}),
					linebot.NewURIImagemapAction("LINE Store Music", "https://store.line.me/family/music/en", linebot.ImagemapArea{X: 520, Y: 0, Width: 520, Height: 520}),
					linebot.NewURIImagemapAction("LINE Store Play", "https://store.line.me/family/play/en", linebot.ImagemapArea{X: 0, Y: 520, Width: 520, Height: 520}),
					linebot.NewMessageImagemapAction("URANAI!", "URANAI!", linebot.ImagemapArea{X: 520, Y: 520, Width: 520, Height: 520}),
				),
			).Do(); err != nil {
				return err
			}
		case "imagemap video":
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewImagemapMessage(
					app.appBaseURL+"/static/rich",
					"Imagemap with video alt text",
					linebot.ImagemapBaseSize{Width: 1040, Height: 1040},
					linebot.NewURIImagemapAction("LINE Store Manga", "https://store.line.me/family/manga/en", linebot.ImagemapArea{X: 0, Y: 0, Width: 520, Height: 520}),
					linebot.NewURIImagemapAction("LINE Store Music", "https://store.line.me/family/music/en", linebot.ImagemapArea{X: 520, Y: 0, Width: 520, Height: 520}),
					linebot.NewURIImagemapAction("LINE Store Play", "https://store.line.me/family/play/en", linebot.ImagemapArea{X: 0, Y: 520, Width: 520, Height: 520}),
					linebot.NewMessageImagemapAction("URANAI!", "URANAI!", linebot.ImagemapArea{X: 520, Y: 520, Width: 520, Height: 520}),
				).WithVideo(&linebot.ImagemapVideo{
					OriginalContentURL: app.appBaseURL + "/static/imagemap/video.mp4",
					PreviewImageURL:    app.appBaseURL + "/static/imagemap/preview.jpg",
					Area:               linebot.ImagemapArea{X: 280, Y: 385, Width: 480, Height: 270},
					ExternalLink:       &linebot.ImagemapVideoExternalLink{LinkURI: "https://line.me", Label: "LINE"},
				}),
			).Do(); err != nil {
				return err
			}
		case "quick":
			if _, err := app.bot.ReplyMessage(
				replyToken,
				linebot.NewTextMessage("Select your favorite food category or send me your location!").
					WithQuickReplies(linebot.NewQuickReplyItems(
						linebot.NewQuickReplyButton(
							app.appBaseURL+"/static/quick/sushi.png",
							linebot.NewMessageAction("Sushi", "Sushi")),
						linebot.NewQuickReplyButton(
							app.appBaseURL+"/static/quick/tempura.png",
							linebot.NewMessageAction("Tempura", "Tempura")),
						linebot.NewQuickReplyButton(
							"",
							linebot.NewLocationAction("Send location")),
					)),
			).Do(); err != nil {
				return err
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
