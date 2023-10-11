package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maxik12233/english-helper-telegrambot/pkg/db"
	"github.com/maxik12233/english-helper-telegrambot/pkg/logger"
	"go.uber.org/zap"
)

const (
	commandStart        = "start"
	commandChooseMode   = "mode"
	commandLanguageSwap = "swap"
	commandRepeat       = "repeat"
	commandStopRepeat   = "stop"

	modeLearn     = "Learn"
	modeTranslate = "Translate"
	modeRepeat    = "Repeat"
)

const (
	sourceDefault = "en"
	targetDefault = "ru"
	modeDefault   = modeLearn
)

var defaultCfg = db.Config{
	UserID: 0, // Rewrite when used to create db instance
	Target: targetDefault,
	Source: sourceDefault,
	Mode:   modeDefault,
}

func (b *Bot) saveMessagesInDb(botmsg *tgbotapi.Message, message *tgbotapi.Message) error {
	log := logger.GetLogger()

	log.Info("Saving messages in database.", zap.Any("botmsg", botmsg), zap.Any("usermsg", message))
	// Save bot's and user's message in db
	if err := b.repo.CreateMessage(&db.Message{
		UserID:     uint(message.From.ID),
		ChatID:     uint(message.Chat.ID),
		Text:       message.Text,
		BotMessage: false,
	}); err != nil {
		log.Error("Error saving message in database", zap.Error(err))
		return err
	}
	if err := b.repo.CreateMessage(&db.Message{
		UserID:     uint(message.From.ID),
		ChatID:     uint(message.Chat.ID),
		Text:       botmsg.Text,
		BotMessage: true,
	}); err != nil {
		log.Error("Error saving message in database", zap.Error(err))
		return err
	}

	return nil
}

func (b *Bot) handleMessage(message *tgbotapi.Message) error {
	log := logger.GetLogger()

	cfg, err := b.GetOrCreateUserConfig(uint(message.From.ID))
	if err != nil {
		return ErrInternal
	}
	log.Info("Obtained config", zap.Any("Config", cfg))

	var botmsg *tgbotapi.Message
	switch cfg.Mode {
	case modeRepeat:
		log.Info("Starting repeat seesion.")
		botmsg, err = b.handleRepeatMessage(message)
		if err != nil {
			return err
		}
	default:
		log.Info("Starting translation.")
		botmsg, err = b.handleTranslateMessage(message)
		if err != nil {
			return err
		}
	}

	b.saveMessagesInDb(botmsg, message)

	return nil
}

func (b *Bot) handleRepeatMessage(message *tgbotapi.Message) (*tgbotapi.Message, error) {
	log := logger.GetLogger()

	msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, something went wrong.")

	cfg, err := b.GetOrCreateUserConfig(uint(message.From.ID))
	if err != nil {
		return nil, ErrInternal
	}
	log.Info("Obtained config", zap.Any("Config", cfg))

	ok := false
	if message.Text != cfg.TranslationWord {
		msg.Text = "Incorrect."
	} else {
		msg.Text = "Excellent!"
		ok = true
	}

	sendmsg, err := b.bot.Send(msg)
	if err != nil {
		return nil, ErrSending
	}

	if ok {
		b.repeatWord(message)
	}

	return &sendmsg, nil
}

func (b *Bot) handleTranslateMessage(message *tgbotapi.Message) (*tgbotapi.Message, error) {
	log := logger.GetLogger()

	msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, something went wrong.")

	cfg, err := b.GetOrCreateUserConfig(uint(message.From.ID))
	if err != nil {
		return nil, ErrInternal
	}
	log.Info("Obtained config", zap.Any("Config", cfg))

	resultText, err := b.translateService.TranslateText(message.Text, cfg.Target, cfg.Source)
	if err != nil {
		return nil, ErrTranslationApi
	}

	if resultText != "" {
		msg.Text = resultText

		if cfg.Mode == modeLearn {
			// Save result of translation operation in db if mode learn
			// If translation was performed (dont depends on send error)
			if err := b.repo.CreateTranslation(&db.Translation{
				UserID:     uint(message.From.ID),
				ChatID:     uint(message.Chat.ID),
				SourceText: message.Text,
				TargetText: msg.Text,
				Source:     cfg.Source,
				Target:     cfg.Target,
			}); err != nil {
				return nil, ErrCreatingTranslation
			}
		}
	}

	sendmsg, err := b.bot.Send(msg)
	if err != nil {
		return nil, ErrSending
	}

	return &sendmsg, nil
}

func (b *Bot) handleCommand(message *tgbotapi.Message) error {

	var botmsg *tgbotapi.Message
	var err error
	switch message.Command() {
	case commandStart:
		botmsg, err = b.handleStartCommand(message)
		if err != nil {
			return err
		}
	case commandChooseMode:
		botmsg, err = b.handleChooseModeCommand(message)
		if err != nil {
			return err
		}
	case commandLanguageSwap:
		botmsg, err = b.handleSwapCommand(message)
		if err != nil {
			return err
		}
	case commandRepeat:
		botmsg, err = b.handleRepeatCommand(message)
		if err != nil {
			return err
		}
	case commandStopRepeat:
		botmsg, err = b.handleStopRepeatCommand(message)
		if err != nil {
			return err
		}
	default:
		botmsg, err = b.handleUnknownCommand(message)
		if err != nil {
			return err
		}
	}

	b.saveMessagesInDb(botmsg, message)

	return nil
}

func (b *Bot) repeatWord(message *tgbotapi.Message) (*tgbotapi.Message, error) {
	cfg, err := b.GetOrCreateUserConfig(uint(message.From.ID))
	if err != nil {
		return nil, ErrInternal
	}

	trnsl, err := b.repo.GetRandomTranslation()
	if err != nil {
		return nil, err
	}

	cfg.TranslationWord = trnsl.TargetText
	err = b.repo.UpdateConfig(cfg)
	if err != nil {
		return nil, err
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, trnsl.SourceText)
	botmsg, err := b.bot.Send(msg)
	if err != nil {
		return nil, ErrSending
	}

	return &botmsg, nil
}

func (b *Bot) handleStopRepeatCommand(message *tgbotapi.Message) (*tgbotapi.Message, error) {

	cfg, err := b.GetOrCreateUserConfig(uint(message.From.ID))
	if err != nil {
		return nil, ErrInternal
	}
	cfg.Mode = modeDefault
	b.repo.UpdateConfig(cfg)
	if err != nil {
		return nil, ErrInternal
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Repeat sessiong is off. Current mode - %v.", modeDefault))
	botmsg, err := b.bot.Send(msg)
	return &botmsg, nil
}

func (b *Bot) handleRepeatCommand(message *tgbotapi.Message) (*tgbotapi.Message, error) {

	cfg, err := b.GetOrCreateUserConfig(uint(message.From.ID))
	if err != nil {
		return nil, ErrInternal
	}
	cfg.Mode = modeRepeat
	b.repo.UpdateConfig(cfg)
	if err != nil {
		return nil, err
	}

	botmsg, err := b.repeatWord(message)
	if err != nil {
		return nil, err
	}

	return botmsg, nil
}

func (b *Bot) handleStartCommand(message *tgbotapi.Message) (*tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, something went wrong.")

	cfg, err := b.GetOrCreateUserConfig(uint(message.From.ID))
	if err != nil {
		return nil, ErrInternal
	}

	msg.Text = fmt.Sprintf("Let's begin. Type any word or phrase you want to translate. Default translate setting: %v -> %v", cfg.Source, cfg.Target)
	botmsg, err := b.bot.Send(msg)
	if err != nil {
		return nil, ErrSending
	}

	return &botmsg, nil
}

func (b *Bot) handleChooseModeCommand(message *tgbotapi.Message) (*tgbotapi.Message, error) {

	cfg, err := b.GetOrCreateUserConfig(uint(message.From.ID))
	if err != nil {
		return nil, ErrInternal
	}

	if cfg.Mode == modeLearn {
		cfg.Mode = modeTranslate
	} else {
		cfg.Mode = modeLearn
	}
	err = b.repo.UpdateConfig(cfg)
	if err != nil {
		return nil, ErrInternal
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Mode saved to - %v.", cfg.Mode))
	botmsg, err := b.bot.Send(msg)
	if err != nil {
		return nil, ErrSending
	}

	return &botmsg, nil
}

func (b *Bot) handleSwapCommand(message *tgbotapi.Message) (*tgbotapi.Message, error) {

	cfg, err := b.GetOrCreateUserConfig(uint(message.From.ID))
	if err != nil {
		return nil, ErrInternal
	}

	templ := cfg.Target
	cfg.Target = cfg.Source
	cfg.Source = templ

	err = b.repo.UpdateConfig(cfg)
	if err != nil {
		return nil, ErrInternal
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, fmt.Sprintf("Languages saved. Current settings: %v -> %v.", cfg.Source, cfg.Target))
	botmsg, err := b.bot.Send(msg)
	if err != nil {
		return nil, ErrSending
	}

	return &botmsg, nil
}

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message) (*tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Sorry, something went wrong.")

	msg.Text = "Invalid command."
	botmsg, err := b.bot.Send(msg)
	if err != nil {
		return nil, ErrSending
	}

	return &botmsg, nil
}
