package main

import (
	"context"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/maxik12233/english-helper-telegrambot/pkg/db"
	gTranslate "github.com/maxik12233/english-helper-telegrambot/pkg/google-translate-sdk"
	"github.com/maxik12233/english-helper-telegrambot/pkg/logger"
	"github.com/maxik12233/english-helper-telegrambot/pkg/telegram"
	"go.uber.org/zap"
)

func main() {

	err := godotenv.Load("../../.env")
	if err != nil {
		panic(err)
	}

	logger.Init()
	log := logger.GetLogger()

	botAPI, err := tgbotapi.NewBotAPI(os.Getenv("BOT_API_KEY"))
	if err != nil {
		log.Fatal("Failed creating new bot api instance", zap.Error(err))
		panic(err)
	}

	client, err := db.InitMongoConnection()
	if err != nil {
		log.Fatal("Error while initializing mongoDB connection", zap.Error(err))
		panic("DB error")
	}
	if client == nil {
		log.Fatal("Mongo connection is nil, failed creating mongo connection", zap.Error(err))
		panic("Mongo connection is nil")
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal("Failed closing mongo connection.", zap.Error(err))
			panic(err)
		}
	}()

	repo := db.NewMongoRepo(client.Database("bot"))

	translater, err := gTranslate.NewClient(gTranslate.Config{
		Key: os.Getenv("GTRANSLATE_API_KEY"),
	}, http.DefaultClient)
	if err != nil {
		log.Fatal("Failed creating new translater instance.", zap.Error(err))
		panic(err)
	}

	bot := telegram.NewBot(botAPI, repo, translater)

	log.Info("App initialized, starting bot service")
	bot.Start()
}
