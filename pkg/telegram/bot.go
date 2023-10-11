package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maxik12233/english-helper-telegrambot/pkg/db"
	gTranslate "github.com/maxik12233/english-helper-telegrambot/pkg/google-translate-sdk"
	"github.com/maxik12233/english-helper-telegrambot/pkg/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type Bot struct {
	bot              *tgbotapi.BotAPI
	repo             db.IRepository
	translateService gTranslate.IClient
}

func NewBot(botAPI *tgbotapi.BotAPI, repo db.IRepository, translateService gTranslate.IClient) Bot {
	return Bot{
		bot:              botAPI,
		repo:             repo,
		translateService: translateService,
	}
}

func (b *Bot) Start() {
	b.bot.Debug = true
	updates := b.CreateUpdateChan()
	b.handleUpdates(updates)
}

func (b *Bot) CreateUpdateChan() tgbotapi.UpdatesChannel {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30
	return b.bot.GetUpdatesChan(updateConfig)
}

func (b *Bot) handleUpdates(updates tgbotapi.UpdatesChannel) {
	log := logger.GetLogger().With(zap.String("place", "Inside handleUpdates"))
	for update := range updates {
		if update.Message == nil {
			log.Info("Updated message in nil")
			continue
		}

		if update.Message.IsCommand() {
			log.Info("Handling command")
			err := b.handleCommand(update.Message)
			if err != nil {
				log.Error("Error while handling a command.", zap.Error(err))
				b.handleError(update.Message.Chat.ID, err)
			}
			continue
		}

		err := b.handleMessage(update.Message)
		if err != nil {
			log.Error("Error while handling a message.", zap.Error(err))
			b.handleError(update.Message.Chat.ID, err)
		}
	}
}

func (b *Bot) GetOrCreateUserConfig(userid uint) (*db.Config, error) {

	cfg, err := b.repo.GetConfig(userid)
	if err == mongo.ErrNoDocuments {
		cfg = &defaultCfg
		cfg.UserID = userid
		err := b.repo.CreateConfig(cfg)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return cfg, nil
}
