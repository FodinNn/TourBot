package bot

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"

	"github.com/FodinNn/TourBot/internal/handlers"
)

const (
	minRetryDelay = time.Second
	maxRetryDelay = 30 * time.Second
)

type HandlerFunc func(string) []handlers.Reply

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
	retry := minRetryDelay

	for ctx.Err() == nil {
		updates, newMarker, err := b.api.Subscriptions.GetUpdates(ctx, marker)
		if err != nil {
			if ctx.Err() != nil {
				break
			}

			log.Printf("Ошибка получения обновлений: %v (повтор через %s)", err, retry)

			select {
			case <-ctx.Done():
			case <-time.After(retry):
			}

			retry = min(retry*2, maxRetryDelay)
			continue
		}

		retry = minRetryDelay
		marker = newMarker

		for _, u := range updates {
			if ctx.Err() != nil {
				break
			}
			b.handle(ctx, u)
		}
	}

	log.Println("Остановка: получен сигнал завершения")
	return nil
}

func (b *Bot) handle(ctx context.Context, u model.Update) {
	switch u.UpdateType {
	case model.UpdateBotStarted:
		log.Printf("Пользователь %d запустил бота (чат %d)", u.UserID, u.ChatID)
		b.send(ctx, u.ChatID, u.UserID, b.handler("старт"))

	case model.UpdateMessageCreated:
		msg := u.GetMessage()

		if msg.Sender.IsBot {
			return
		}

		if lat, lon, ok := extractLocation(msg.Body.Attachments); ok {
			log.Printf("Геопозиция из чата %d: %.6f, %.6f", u.ChatID, lat, lon)
			b.send(ctx, u.ChatID, u.UserID, b.location(lat, lon))
			return
		}

		log.Printf("Сообщение из чата %d от %d: %q", u.ChatID, u.UserID, msg.Body.Text)
		b.send(ctx, u.ChatID, u.UserID, b.handler(msg.Body.Text))

	default:
		log.Printf("Пропущено обновление %q (чат %d)", u.UpdateType, u.ChatID)
	}
}

func extractLocation(attachments []model.Attachment) (float64, float64, bool) {
	for _, a := range attachments {
		if a.Type == model.AttachLocation {
			return a.Latitude, a.Longitude, true
		}
	}
	return 0, 0, false
}

func (b *Bot) send(ctx context.Context, chatID, userID int64, replies []handlers.Reply) {
	for _, r := range replies {
		msg := maxbot.NewMessage().SetText(r.Text)

		switch {
		case chatID != 0:
			msg.SetChat(chatID)
		case userID != 0:
			msg.SetUser(userID)
		default:
			log.Println("Ошибка отправки: в обновлении нет ни chat_id, ни user_id")
			return
		}

		if r.Markdown {
			msg.SetFormat(model.FormatMarkdown)
		}
		if r.Keyboard != nil {
			msg.AddKeyboard(r.Keyboard)
		}

		if _, err := b.api.Messages.Send(ctx, msg); err != nil {
			log.Printf("Ошибка отправки сообщения в чат %d: %v", chatID, err)
		}
	}
}

func DropWebhooks(ctx context.Context, api *maxbot.Api) error {
	subs, err := api.Subscriptions.GetSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("не удалось получить список подписок: %w", err)
	}

	for _, s := range subs.Subscriptions {
		log.Printf("Найден webhook %s — удаляю: пока он есть, long polling не получает сообщения", s.URL)

		if _, err := api.Subscriptions.Unsubscribe(ctx, s.URL); err != nil {
			return fmt.Errorf("не удалось удалить подписку %s: %w", s.URL, err)
		}
	}

	return nil
}

func AcquireLock(path string) (io.Closer, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("не удалось создать каталог %s: %w", dir, err)
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл блокировки %s: %w", path, err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("другой экземпляр бота уже запущен (блокировка %s): %w", path, err)
	}

	_ = f.Truncate(0)
	_, _ = f.WriteString(strconv.Itoa(os.Getpid()) + "\n")

	return f, nil
}
