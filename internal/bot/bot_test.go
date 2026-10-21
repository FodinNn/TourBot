package bot

import (
	"path/filepath"
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"
)

func TestAcquireLockIsExclusive(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "bot.lock")

	first, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatalf("первый экземпляр не получил блокировку: %v", err)
	}
	defer func() {
		_ = first.Close()
	}()

	second, err := AcquireLock(lockPath)
	if err == nil {
		_ = second.Close()
		t.Fatal("второй экземпляр получил блокировку — два процесса будут делить очередь обновлений")
	}
}

func TestAcquireLockAfterRelease(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "bot.lock")

	first, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatalf("первый экземпляр не получил блокировку: %v", err)
	}
	_ = first.Close()

	second, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatalf("после освобождения блокировка не выдана: %v", err)
	}
	_ = second.Close()
}

func TestExtractLocation(t *testing.T) {
	lat, lon, ok := extractLocation([]model.Attachment{
		{Type: model.AttachImage},
		{Type: model.AttachLocation, Latitude: 55.7885, Longitude: 49.1217},
	})
	if !ok {
		t.Fatal("геопозиция не найдена")
	}
	if lat != 55.7885 || lon != 49.1217 {
		t.Fatalf("координаты %f, %f не совпали", lat, lon)
	}

	if _, _, ok := extractLocation([]model.Attachment{{Type: model.AttachImage}}); ok {
		t.Fatal("лишнее вложение распознано как геопозиция")
	}
}
