package bot

import (
	"context"
	"log"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

type HandlerFunc func(string) string

type Bot struct {
	api     *maxbot.Api
	handler HandlerFunc
}

func New(api *maxbot.Api, handler HandlerFunc) *Bot {
	return &Bot{
		api:     api,
		handler: handler,
	}
}

func (b *Bot) Run(ctx context.Context) error {
	var marker int64
	for {
		updates, newMarker, err := b.api.Subscriptions.GetUpdates(ctx, marker)
		if err != nil {
			log.Println("Ошибка получения обновлений", err)
			continue
		}
		marker = newMarker

		for _, update := range updates {
			switch update.UpdateType {
			case model.UpdateMessageCreated:
				messageText := update.GetMessage().Body.Text
				chatID := update.ChatID

				log.Printf("Получено сообщение: %s", messageText)

				reply := b.handler(messageText)
				response := maxbot.NewMessage().SetChat(chatID).SetText(reply)

				_, err := b.api.Messages.Send(ctx, response)
				if err != nil {
					log.Println("Ошибки отправки сообщения", err)
				}
			}
		}
	}
}
