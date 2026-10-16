package main

import (
	"context"
	"fmt"
	"log"
	"os"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

func main() {
	ctx := context.Background()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Переменная окружения BOT_TOKEN не установлена")
	}

	api, err := maxbot.NewApi(token)
	if err != nil {
		log.Fatal(err)
	}

	info, err := api.Bots.GetMyInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Бот запущен %s", info.FirstName)

	var marker int64
	for {
		updates, newMarker, err := api.Subscriptions.GetUpdates(ctx, marker)
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

				response := maxbot.NewMessage().SetChat(chatID).SetText(fmt.Sprint("Ты написал: ", messageText))

				_, err := api.Messages.Send(ctx, response)
				if err != nil {
					log.Println("Ошибки отправки сообщения", err)
				}
			}
		}
	}
}
