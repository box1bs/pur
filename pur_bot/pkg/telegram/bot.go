package telegram

import (
	"log"
	"sync"

	ldb "github.com/box1bs/pur/pur_bot/pkg/localStorage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	bot 	*tgbotapi.BotAPI
	wg  	*sync.WaitGroup
	lc 		ldb.LocalStorage
	auth	bool
}

func NewBot(bot *tgbotapi.BotAPI) *Bot {
	lc, err := ldb.NewRedisStorage(0, "localhost:6379", "")
	if err != nil {
		panic(err)
	}

	purBot := &Bot{
		bot: bot,
		wg: new(sync.WaitGroup),
		lc: lc,
		auth: false,
	}

	return purBot
}

func (b *Bot) Start() error {
	log.Printf("Authorized on account %s", b.bot.Self.UserName)
	
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "start work with bot"},
		{Command: "share_link", Description: "save your link for a while for you"},
		{Command: "get_all_links", Description: "print out all your save links"},
		{Command: "delete_link", Description: "delete saved link by it's url"},
	}
	
	setCommandConf := tgbotapi.NewSetMyCommands(commands...)
	if _, err := b.bot.Request(setCommandConf); err != nil {
		return err
	}

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
				b.handleCommand(update.Message)
				continue
			}
		
			b.handleMessage(update.Message)
		}
	}
}