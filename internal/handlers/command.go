package handlers

import (
	"log"
	"strings"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"

	"github.com/FodinNn/TourBot/internal/data"
)

type Reply struct {
	Text     string
	Keyboard *model.Keyboard
}

var menuCommands = map[string]bool{
	"/start":       true,
	"старт":        true,
	"меню":         true,
	"назад":        true,
	"привет":       true,
	"здравствуйте": true,
	"menu":         true,
}

var farewells = map[string]bool{
	"спасибо":     true,
	"пока":        true,
	"до свидания": true,
	"завершить":   true,
	"стоп":        true,
}

var sectionNames = map[string]int{
	"главные must-see":   1,
	"культура и религия": 2,
	"современная казань": 3,
	"гастрономия":        4,
	"маршрут на 1 день":  5,
	"с детьми":           6,
}

func Handle(text string) []Reply {
	q := normalize(text)

	if q == "" || menuCommands[q] {
		return []Reply{MenuReply()}
	}

	if farewells[q] {
		return []Reply{{Text: data.Farewell}}
	}

	if len(q) == 1 && q[0] >= '1' && q[0] <= '6' {
		return []Reply{SectionReply(int(q[0] - '0'))}
	}

	if n, ok := sectionNames[q]; ok {
		return []Reply{SectionReply(n)}
	}

	if strings.Contains(q, "бауман") {
		return []Reply{{Text: data.BaumanAnswer, Keyboard: BackKeyboard()}}
	}

	for _, kw := range data.Keywords {
		if strings.Contains(q, kw.Word) {
			return []Reply{SectionReply(kw.Section)}
		}
	}

	log.Printf("Неизвестный запрос: %s", q)
	return []Reply{{Text: data.Unknown}, MenuReply()}
}

func normalize(text string) string {
	s := strings.ToLower(text)
	s = strings.Join(strings.Fields(s), " ")
	return strings.Trim(s, " .,!?;:\"«»()-—")
}

func MenuReply() Reply {
	kb := model.NewKeyboard()

	row := kb.AddRow()
	row.AddMessage("Главные must-see")
	row.AddMessage("Культура и религия")

	row = kb.AddRow()
	row.AddMessage("Современная Казань")
	row.AddMessage("Гастрономия")

	row = kb.AddRow()
	row.AddMessage("Маршрут на 1 день")
	row.AddMessage("С детьми")

	geo := kb.AddRow()
	geo.AddGeoLocation(data.GeoButton, false)

	return Reply{Text: data.Greeting(), Keyboard: kb}
}

func SectionReply(n int) Reply {
	return Reply{Text: data.Section(n), Keyboard: BackKeyboard()}
}

func BackKeyboard() *model.Keyboard {
	kb := model.NewKeyboard()
	kb.AddRow().AddMessage("Назад")
	return kb
}
