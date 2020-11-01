package bot

import (
	"database/sql"
	"fmt"
	"github.com/Avimitin/go-bot/internal/auth"
	"github.com/Avimitin/go-bot/internal/bot/internal/KaR"
	"github.com/Avimitin/go-bot/internal/conf"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
)

const (
	VERSION = "0.5.8"
	CREATOR = 649191333
)

var (
	DB   *sql.DB
	bot  *tgbotapi.BotAPI
	data *conf.BotData
)

func Run(CleanMode bool) {
	// Bot init
	log.Printf("Bot initializing... Version: %v", VERSION)
	bot = NewBot()
	bot.Debug = true
	log.Printf("Authorized on accout %s", bot.Self.UserName)

	// DB init
	log.Print("Fetching database connection...")
	DB = NewDB()
	defer DB.Close()

	log.Print("Initializing data")
	data = NewData()

	log.Print("Fetching authorized groups...")
	data.Groups = NewAuthGroups()

	log.Print("Fetching keywords and replies...")
	NewKeywordReplies()

	updateMsg := tgbotapi.NewUpdate(0)
	updateMsg.Timeout = 20

	updates, err := bot.GetUpdatesChan(updateMsg)

	if err != nil {
		log.Printf("Some error occur when getting update.\nDescriptions: %v", err)
	}

	// 清理模式
	for CleanMode {
		log.Printf("Cleaning MSG...")
		<-updates
		os.Exit(0)
	}

	for update := range updates {

		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		if update.Message.Chat.Type == "supergroup" && !auth.CFGIsAuthGroups(data, update.Message.Chat.ID) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "你们这啥群啊，别乱拉人，爬爬爬！")
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("[ERROR] %s", err)
			}

			_, err = bot.LeaveChat(update.Message.Chat.ChatConfig())
			if err != nil {
				log.Printf("[ERROR] Error happen when leave chat %s", err)
			}
		}

		if update.Message.IsCommand() {
			go commandHandler(update.Message)
			continue
		}

		go keywordHandler(update.Message)
	}
}

func commandHandler(message *tgbotapi.Message) {
	cmd, hasElem := COMMAND[message.Command()]
	if hasElem {
		_, err := cmd(bot, message)

		if err != nil {
			_, _ = SendParseTextMsg(bot, message.Chat.ID,
				fmt.Sprintf("<b>Some error happen when sending message.</b> \n\nDescriptions: \n\n<code>%v</code>", err),
				"html")
		}
	}
}

func keywordHandler(message *tgbotapi.Message) {
	reply, e := KaR.RegexKAR(message.Text, data)
	if e {
		_, _ = SendTextMsg(bot, message.Chat.ID, reply)
	}
}
