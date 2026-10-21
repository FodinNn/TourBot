package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	maxbot "github.com/max-messenger/max-bot-api-client-go/v2"

	"github.com/FodinNn/TourBot/internal/bot"
	"github.com/FodinNn/TourBot/internal/handlers"
)

const defaultLockFile = "bot.lock"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := godotenv.Load(); err != nil {
		log.Println("Файл .env не найден, использую переменные окружения")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("Переменная окружения BOT_TOKEN не установлена")
	}

	lockFile := os.Getenv("BOT_LOCK_FILE")
	if lockFile == "" {
		lockFile = defaultLockFile
	}

	lock, err := bot.AcquireLock(lockFile)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = lock.Close()
	}()

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
	log.Printf("Бот запущен: %s (@%s)", info.FirstName, info.Username)

	if err := bot.DropWebhooks(ctx, api); err != nil {
		log.Fatal(err)
	}

	b := bot.New(api, handlers.Handle, handlers.HandleLocation)

	log.Println("Запускаю цикл получения обновлений")
	if err := b.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
