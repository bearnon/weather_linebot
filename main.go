package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

func getWeather() string {
	stationID := "C0D660" // Hsinchu East District
	authorization := os.Getenv("CWB_AUTHORIZATION")
	res, err := http.Get("https://opendata.cwb.gov.tw/api/v1/rest/datastore/O-A0001-001?Authorization=" + authorization + "&stationId=" + stationID)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var weather map[string]interface{}
	if err := json.Unmarshal(body, &weather); err != nil {
		fmt.Println(err)
	}
	// indent, err := json.MarshalIndent(weather, "", "\t")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	var temp string
	var pres string

	for _, ele := range weather["records"].(map[string]interface{})["location"].([]interface{})[0].(map[string]interface{})["weatherElement"].([]interface{}) {
		eleValue := ele.(map[string]interface{})["elementValue"].(string)
		switch ele.(map[string]interface{})["elementName"].(string) {
		case "TEMP":
			fmt.Println("Temperature:", eleValue)
			temp = eleValue
			// weatherInformation = append(weatherInformation, "Temperature:"+eleValue)
		case "PRES":
			fmt.Println("Pressure:", eleValue)
			pres = eleValue
			// weatherInformation = append(weatherInformation, "Pressure:"+eleValue)
		}
	}
	return fmt.Sprintf("Temperature: %s\nPressure: %s", temp, pres)
}

func main() {
	port := "80"
	if envPort := os.Getenv("PORT"); len(envPort) > 0 {
		port = envPort
	}
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	bot, err := linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("CHANNEL_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	RecvLine := func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					fmt.Println(message.Text)
					if strings.Contains(message.Text, "weather") {
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(getWeather())).Do(); err != nil {
							log.Print(err)
						}
					} else {
						if _, err := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
							log.Print(err)
						}
					}
					// case *linebot.StickerMessage:
				}
			}
		}
	}

	http.HandleFunc("/callback", RecvLine)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
