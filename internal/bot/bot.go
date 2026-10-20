package bot

import (
	"context"
	"log"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"

	"github.com/FodinNn/TourBot/internal/handlers"
)

// HandlerFunc обрабатывает текстовое сообщение пользователя.
type HandlerFunc func(string) []handlers.Reply

// LocationFunc обрабатывает геопозицию пользователя (вложение-локация).
type LocationFunc func(lat, lon float64) []handlers.Reply

type Bot struct {
	api      *maxbot.Api
	handler  HandlerFunc
	location LocationFunc
}

func New(api *maxbot.Api, handler HandlerFunc, location LocationFunc) *Bot {
	return &Bot{
		api:      api,
		handler:  handler,
		location: location,
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

		for _, u := range updates {
			switch u.UpdateType {
			case model.UpdateBotStarted:
				log.Println("Пользователь запустил бота", u.ChatID)
				b.send(ctx, u.ChatID, b.handler("старт"))

			case model.UpdateMessageCreated:
				body := u.GetMessage().Body

				if lat, lon, ok := extractLocation(body.Attachments); ok {
					log.Printf("Получена геопозиция: %f, %f", lat, lon)
					b.send(ctx, u.ChatID, b.location(lat, lon))
					continue
				}

				log.Printf("Получено сообщение: %s", body.Text)
				b.send(ctx, u.ChatID, b.handler(body.Text))
			}
		}
	}
}

// extractLocation ищет в вложениях сообщения геопозицию.
func extractLocation(attachments []model.Attachment) (float64, float64, bool) {
	for _, a := range attachments {
		if a.Type == model.AttachLocation {
			return a.Latitude, a.Longitude, true
		}
	}
	return 0, 0, false
}

func (b *Bot) send(ctx context.Context, chatID int64, replies []handlers.Reply) {
	for _, r := range replies {
		msg := maxbot.NewMessage().SetChat(chatID).SetText(r.Text)
		if r.Keyboard != nil {
			msg.AddKeyboard(r.Keyboard)
		}

		if _, err := b.api.Messages.Send(ctx, msg); err != nil {
			log.Println("Ошибка отправки сообщения", err)
		}
	}
}
