package telegram

import (
	"log"
	"sync"

	localstorage "github.com/box1bs/pur/pur_bot/pkg/localStorage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot *tgbotapi.BotAPI
	wg  *sync.WaitGroup
	lc 	localstorage.LocalStorage
}

func NewBot(bot *tgbotapi.BotAPI) *Bot {
	return &Bot{bot: bot, wg: new(sync.WaitGroup), lc: localstorage.NewRedisStorage(0, "localhost:6379", "")}
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.bot.Self.UserName)

	b.handleUpdates()

	return nil
}

func (b *Bot) handleUpdates() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			if update.Message.IsCommand() {
				log.Println(b.handleCommand(update.Message))
				return
			}
		
			b.handleMessage(update.Message)
		}
		u.Offset = update.UpdateID + 1
	}
}