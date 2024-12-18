package telegram

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/box1bs/pur/pur_bot/pkg/sdk/auth"
	"github.com/box1bs/pur/pur_bot/pkg/sdk/resources"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const commandStart = "start"

func(b *Bot) handleCommand(message *tgbotapi.Message) error {

	switch message.Command() {
	case commandStart:
		return b.handleStartCommand(message)
	case "delete":
		return b.handleDeleteCommand(message)
	case "shareLink":
		return b.handleShareCommand(message)
	case "getAllLinks":
		return b.handleGetCommand(message)
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
	if b.auth {
		return nil
	}
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
			b.auth = true
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

	b.auth = false

	_, err = b.bot.Send(msg)
	return err
}

func (b *Bot) handleShareCommand(message *tgbotapi.Message) error {
	args := strings.TrimSpace(message.CommandArguments())

	if args == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please type like:\n/shareLink <url> <description>")
		b.bot.Send(msg)
		return fmt.Errorf("empty args")
	}

	parts := strings.SplitN(args, " ", 2)
	link, desc := parts[0], parts[1]

	req := &resources.ReqResource{Addr: "http://localhost:8080/link", Client: &http.Client{Timeout: 200 * time.Millisecond}}

	id, err := b.lc.GetSyncId(message.Chat.ID)
	if err != nil {
		return err
	}

	if err := req.SaveLink(id, link, desc); err != nil {
		return err
	}

	return nil
}

func (b *Bot) handleGetCommand(message *tgbotapi.Message) error {
	id, err := b.lc.GetSyncId(message.Chat.ID)
	if err != nil {
		return err
	}
	req := &resources.ReqResource{Addr: fmt.Sprintf("http://localhost:8080/link/%s", id), Client: &http.Client{Timeout: 200 * time.Millisecond}}
	links, err := req.GetAllLinks()
	if err != nil {
		return err
	}

	var messageStr string

	for _, link := range links {
		messageStr += link.PresentLink()
	}

	msg := tgbotapi.NewMessage(message.Chat.ID, messageStr)
	_, err = b.bot.Send(msg)

	return err
}

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Not allowed command")

	_, err := b.bot.Send(msg)
	return err
}