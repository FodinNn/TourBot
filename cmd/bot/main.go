package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"

	"github.com/FodinNn/TourBot/internal/bot"
	"github.com/FodinNn/TourBot/internal/handlers"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, использую переменные окружения")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Переменная окружения BOT_TOKEN не установлена")
	}

	api, err := maxbot.NewApi(
		token,
		maxbot.WithBaseURL("https://botapi.max.ru"),
		maxbot.WithPollingTimeout(25*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	info, err := api.Bots.GetMyInfo(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Бот запущен: %s", info.FirstName)

	b := bot.New(api, handlers.Handle, handlers.HandleLocation)

	log.Println("Запускаю цикл получения обновлений")
	if err := b.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
