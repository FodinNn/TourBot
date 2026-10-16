package main

import (
	"context"
	"log"
	"os"

	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"
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
}
