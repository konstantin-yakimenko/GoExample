package main

import (
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func getToken() string {
	return "1991242971:AAGo27nj7DrqBRH2EZc8BvVitFMZSFX32CQ"
}

type wallet map[string]float64

var db = map[int]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI(getToken())
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		command := strings.Split(update.Message.Text, " ")
		userId := update.Message.From.ID

		fmt.Println(command)
		switch command[0] {
		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные агрументы"))
				continue
			}
			money, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
			}
			_, err = getPrice(command[1])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная валюта"))
				continue
			}

			if _, ok := db[userId]; !ok {
				db[userId] = make(wallet)
			}
			db[userId][command[1]] += money

			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "команда ADD"))
		case "SUB":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные агрументы"))
				continue
			}
			money, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}

			if _, ok := db[userId]; !ok {
				db[userId] = make(wallet)
			}
			db[userId][command[1]] -= money
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "команда SUB"))
		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные агрументы"))
				continue
			}
			delete(db[userId], command[1])
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "команда DEL"))
		case "SHOW":
			resp := ""
			for key, value := range db[userId] {
				usdPrice, err := getPrice(key)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					continue
				}
				rubPrice, err := getUsdRub()
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					continue
				}

				resp += fmt.Sprintf("%s: %.2f рублей\n", key, value * usdPrice * rubPrice)
			}
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, resp))
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда"))
		}

		//log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		//msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		//msg.ReplyToMessageID = update.Message.MessageID
		//bot.Send(msg)
	}
}

type bResponse struct {
	Symbol 	string 	`json:"symbol"`
	Price 	float64	`json:"price,string"`
}

func getPrice(symbol string) (float64, error) {
	url := fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	var bRes bResponse
	err = json.NewDecoder(resp.Body).Decode(&bRes)
	if err != nil {
		return 0, err
	}
	if bRes.Symbol == "" {
		return 0, errors.New("неверная валюта ")
	}
	return bRes.Price, nil
}

func getUsdRub() (float64, error) {
	resp, err := http.Get("https://api.binance.com/api/v3/ticker/price?symbol=USDTRUB")
	if err != nil {
		return 0, err
	}

	var bRes bResponse
	err = json.NewDecoder(resp.Body).Decode(&bRes)
	if err != nil {
		return 0, err
	}
	if bRes.Symbol == "" {
		return 0, errors.New("Ошибка при получении курса доллара")
	}
	return bRes.Price, nil
}
/*
func main() {
	http.HandleFunc("/", hello)
	_ = http.ListenAndServe("localhost:8080", nil)
}

func hello(writer http.ResponseWriter, request *http.Request) {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	name := nameGenerator.Generate()
	fmt.Println(name)
	fmt.Fprintf(writer, "Hello %s", request.URL.Path[1:])
}
*/
