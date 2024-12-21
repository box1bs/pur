package telegram

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/box1bs/pur/pur_bot/pkg/sdk/auth"
	"github.com/box1bs/pur/pur_bot/pkg/sdk/resources"
	"github.com/box1bs/pur/pur_bot/pkg/telegram/messanges"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func(b *Bot) handleCommand(message *tgbotapi.Message) error {

	switch message.Command() {
	case messanges.Start:
		return b.handleStartCommand(message)
	case messanges.Del:
		return b.handleDeleteCommand(message)
	case messanges.DeleteLink:
		return b.HandleDeleteLinkCommand(message)
	case messanges.Save:
		return b.handleShareCommand(message)
	case messanges.Get:
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
		b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, messanges.GetWelcomeMessange()))
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
            log.Println("Authorization error:", err)
        } else {
			b.auth = true
            successMsg := tgbotapi.NewMessage(message.Chat.ID, messanges.GetWelcomeMessange())
            b.bot.Send(successMsg)
        }
    }()

	b.wg.Wait()

	return nil
}

func (b *Bot) handleDeleteCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Success")
	uid, err := b.lc.GetSyncId(message.Chat.ID)
	if err != nil {
		return err
	}

	user := &auth.AccountData{
		Name: message.From.UserName,
		Id: uid,
		Client: &http.Client{
			Timeout: time.Second * 1,
		},
	}
	if err := user.DeleteAccount(); err != nil {
		return err
	}

	b.auth = false

	_, err = b.bot.Send(msg)
	return err
}

func (b *Bot) HandleDeleteLinkCommand(message *tgbotapi.Message) error {
	url := strings.TrimSpace(message.CommandArguments())
	if url == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please type like:\n/delete_link <url>")
		b.bot.Send(msg)
		return fmt.Errorf("empty args")
	}

	uid, err := b.lc.GetSyncId(message.Chat.ID)
	if err != nil {
		return err
	}

	req := &resources.ReqResource{Addr: fmt.Sprintf("http://localhost:8080/link/%s", uid.String()), Client: &http.Client{Timeout: 200 * time.Millisecond}}
	if err := req.DeleteLink(); err != nil {
		return err
	}

	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "link successfully deleted"))
	return nil
}

func (b *Bot) handleShareCommand(message *tgbotapi.Message) error {
	args := strings.TrimSpace(message.CommandArguments())

	if args == "" {
		msg := tgbotapi.NewMessage(message.Chat.ID, "Please type like:\n/share_link <url> <description>")
		b.bot.Send(msg)
		return fmt.Errorf("empty args")
	}

	parts := strings.SplitN(args, " ", 2)
	link, desc := parts[0], parts[1]

	req := &resources.ReqResource{Addr: "http://localhost:8080/link", Client: &http.Client{Timeout: 2 * time.Second}}

	id, err := b.lc.GetSyncId(message.Chat.ID)
	if err != nil {
		log.Printf("error getting id from redis: %v", err)
		return err
	}

	if err := req.SaveLink(id, link, desc); err != nil {
		log.Printf("error saving link: %v", err)
		return err
	}

	b.bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Saved"))

	return nil
}

func (b *Bot) handleGetCommand(message *tgbotapi.Message) error {
	id, err := b.lc.GetSyncId(message.Chat.ID)
	if err != nil {
		return err
	}
	req := &resources.ReqResource{Addr: fmt.Sprintf("http://localhost:8080/link/%s", id.String()), Client: &http.Client{Timeout: 200 * time.Millisecond}}
	links, err := req.GetAllLinks()
	if err != nil {
		log.Println(err)
		return err
	}

	for _, link := range links {
		msg := tgbotapi.NewMessage(message.Chat.ID, link.PresentLink())
		b.bot.Send(msg)
	}

	return nil
}

func (b *Bot) handleUnknownCommand(message *tgbotapi.Message) error {
	msg := tgbotapi.NewMessage(message.Chat.ID, "Not allowed command")

	_, err := b.bot.Send(msg)
	return err
}