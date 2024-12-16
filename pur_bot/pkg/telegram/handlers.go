package telegram

import (
	"log"
	"net/http"
	"time"

	"github.com/box1bs/pur/pur_bot/pkg/sdk/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const commandStart = "start"

func(b *Bot) handleCommand(message *tgbotapi.Message) error {

	switch message.Command() {
	case commandStart:
		return b.handleStartCommand(message)
	case "delete":
		return b.handleDeleteCommand(message)
	default:
		return b.handleUnknownCommand(message)
	}
}

func(b *Bot) handleMessage(message *tgbotapi.Message) {
	log.Printf("[%s] %s", message.From.UserName, message.Text)
	
	msg := tgbotapi.NewMessage(message.Chat.ID, message.Text)
	msg.ReplyToMessageID = message.MessageID
	b.bot.Send(msg)
}

func (b *Bot) handleStartCommand(message *tgbotapi.Message) error {
	controlChan := make(chan error)
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		if err := b.lc.SyncId(message.Chat.ID); err != nil {
			log.Printf("error sync id: %v", err)
			return
		}

		uid, err := b.lc.GetSyncId(message.Chat.ID)
		if err != nil {
			log.Printf("error getting id: %v", err)
			return
		}

		user := &auth.AccountData{
			Name: message.From.UserName,
			Id: uid,
			Client: &http.Client{
				Timeout: time.Second * 10,
			},
		}
		controlChan <- user.Authorizate()
	}()
	
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
        if err := <-controlChan; err != nil {
            errorMsg := tgbotapi.NewMessage(message.Chat.ID, "Ошибка авторизации")
            b.bot.Send(errorMsg)
            log.Println("Authorization error:", err)
        } else {
            successMsg := tgbotapi.NewMessage(message.Chat.ID, "Вы успешно авторизовались!")
            b.bot.Send(successMsg)
        }
    }()

	b.wg.Wait()

	return nil
}

func (b *Bot) handleDeleteCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "data deleted")
	uid, err := b.lc.GetSyncId(message.Chat.ID)
	if err != nil {
		return err
	}

	user := &auth.AccountData{
		Name: message.From.UserName,
		Id: uid,
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
	if err := user.DeleteAccount(); err != nil {
		return err
	}

	_, err = b.bot.Send(msg)
	return err
}

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Not allowed command")

	_, err := b.bot.Send(msg)
	return err
}