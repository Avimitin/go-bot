package bot

import (
	"fmt"
	"log"

	bapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func sendT(text string, chatID int64) (bapi.Message, error) {
	msg := bapi.NewMessage(chatID, text)
	return bot.Send(msg)
}

func sendP(text string, chatID int64, parseMode string) (bapi.Message, error) {
	msg := bapi.NewMessage(chatID, text)
	msg.ParseMode = parseMode
	return bot.Send(msg)
}

func editT(newText string, chatID int64, msgID int) (bapi.Message, error) {
	msg := bapi.NewEditMessageText(chatID, msgID, newText)
	return bot.Send(msg)
}

func errF(where string, err error, more string) error {
	log.Printf("[%s]%v", where, err)
	return fmt.Errorf("%s", more)
}
