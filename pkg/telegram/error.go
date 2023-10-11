package telegram

import (
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maxik12233/english-helper-telegrambot/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrInternal            = errors.New("Internal error occured, try later.")
	ErrTranslationApi      = errors.New("Outer API error, try later.")
	ErrCreatingTranslation = errors.New("Internal gateway error.")
	ErrSending             = errors.New("Error occured while sending your results.")
)

func (b *Bot) handleError(chatid int64, err error) {
	log := logger.GetLogger()
	msg := tgbotapi.NewMessage(chatid, "Sorry, something went wrong.")

	switch err {
	case ErrTranslationApi:
		msg.Text = err.Error()
	case ErrInternal:
		msg.Text = err.Error()
	case ErrCreatingTranslation:
		msg.Text = err.Error()
	case ErrSending:
		msg.Text = err.Error()
	}

	_, err = b.bot.Send(msg)
	if err != nil {
		log.Error("Error in error handling method, cannot send message", zap.Error(err))
	}
}
