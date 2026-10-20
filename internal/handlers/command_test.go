package handlers

import (
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/v2/model"

	"github.com/FodinNn/TourBot/internal/data"
)

func buttonTexts(r Reply) []string {
	if r.Keyboard == nil {
		return nil
	}
	var texts []string
	for _, row := range r.Keyboard.Build().Payload.Buttons {
		for _, b := range row {
			texts = append(texts, b.Text)
		}
	}
	return texts
}

func hasButton(r Reply, text string) bool {
	for _, t := range buttonTexts(r) {
		if t == text {
			return true
		}
	}
	return false
}

func TestMenuCommandVariants(t *testing.T) {
	for _, input := range []string{"/start", "старт", "меню", "назад", "привет", "МЕНЮ!", "ПРИВЕТ !", "назад.", " menu "} {
		replies := Handle(input)
		if len(replies) != 1 {
			t.Fatalf("%q: ожидался 1 ответ, получено %d", input, len(replies))
		}
		if replies[0].Text != data.Greeting() {
			t.Fatalf("%q: ожидалось меню, получено: %q", input, replies[0].Text[:30])
		}
		if !hasButton(replies[0], "Гастрономия") {
			t.Fatalf("%q: в меню нет кнопки «Гастрономия»", input)
		}
		if !hasButton(replies[0], data.GeoButton) {
			t.Fatalf("%q: в меню нет гео-кнопки", input)
		}
	}
}

func TestSectionNumbers(t *testing.T) {
	for n := 1; n <= 6; n++ {
		replies := Handle(string(rune('0' + n)))
		if len(replies) != 1 {
			t.Fatalf("%d: ожидался 1 ответ", n)
		}
		if replies[0].Text != data.Section(n) {
			t.Fatalf("%d: текст раздела не совпал", n)
		}
		if !hasButton(replies[0], "Назад") {
			t.Fatalf("%d: нет кнопки «Назад»", n)
		}
	}
}

func TestSectionNames(t *testing.T) {
	cases := map[string]int{
		"Главные must-see":   1,
		"Культура и религия": 2,
		"Современная Казань": 3,
		"Гастрономия":        4,
		"Маршрут на 1 день":  5,
		"С детьми":           6,
	}
	for input, want := range cases {
		replies := Handle(input)
		if replies[0].Text != data.Section(want) {
			t.Fatalf("%q: ожидался раздел %d", input, want)
		}
	}
}

func TestFarewellExactMatchOnly(t *testing.T) {
	for _, input := range []string{"спасибо", "пока", "пока.", "до свидания", "стоп"} {
		replies := Handle(input)
		if replies[0].Text != data.Farewell {
			t.Fatalf("%q: ожидалось прощание", input)
		}
	}
	replies := Handle("покажи кремль")
	if replies[len(replies)-1].Text == data.Farewell {
		t.Fatalf("«покажи кремль» ошибочно распознано как прощание")
	}
}

func TestBaumanKeyword(t *testing.T) {
	for _, input := range []string{"баумана", "Улица Баумана", "БАУМАН?"} {
		replies := Handle(input)
		if replies[0].Text != data.BaumanAnswer {
			t.Fatalf("%q: ожидался ответ про Баумана", input)
		}
		if !hasButton(replies[0], "Назад") {
			t.Fatalf("%q: нет кнопки «Назад»", input)
		}
	}
}

func TestKeywordDispatch(t *testing.T) {
	cases := map[string]int{
		"кул шариф":           1,
		"марджани":            2,
		"храм всех религий":   2,
		"набережная":          3,
		"озеро кабан":         3,
		"где поесть":          4,
		"чак-чак":             4,
		"маршрут на выходные": 5,
		"куда с ребенком":     6,
	}
	for input, want := range cases {
		replies := Handle(input)
		if replies[0].Text != data.Section(want) {
			t.Fatalf("%q: ожидался раздел %d, получено: %q", input, want, replies[0].Text[:25])
		}
	}
}

func TestUnknownQuery(t *testing.T) {
	replies := Handle("foo bar baz")
	if len(replies) != 2 {
		t.Fatalf("ожидалось 2 сообщения (загалка + меню), получено %d", len(replies))
	}
	if replies[0].Text != data.Unknown {
		t.Fatalf("первое сообщение — не заглушка")
	}
	if replies[1].Text != data.Greeting() {
		t.Fatalf("второе сообщение — не меню")
	}
}

func TestEmptyInput(t *testing.T) {
	replies := Handle("   ")
	if len(replies) != 1 || replies[0].Text != data.Greeting() {
		t.Fatalf("пустой ввод должен давать меню")
	}
}

func TestDistanceMetersKnownValues(t *testing.T) {
	d := DistanceMeters(55.7885, 49.1217, 55.7988, 49.1052)
	if d < 1100 || d > 1600 {
		t.Fatalf("расстояние Бауман—Кремль=%d м, ожидалось 1100..1600", d)
	}

	if got := DistanceMeters(55.7, 49.1, 55.7, 49.1); got != 0 {
		t.Fatalf("нулевое расстояние=%d, ожидалось 0", got)
	}
}

func TestFormatDistance(t *testing.T) {
	if got := FormatDistance(350); got != "350 м" {
		t.Fatalf("FormatDistance(350)=%q, ожидалось %q", got, "350 м")
	}
	if got := FormatDistance(1300); got != "1,3 км" {
		t.Fatalf("FormatDistance(1300)=%q, ожидалось %q", got, "1,3 км")
	}
	if got := FormatDistance(1350); got != "1,4 км" {
		t.Fatalf("FormatDistance(1350)=%q, ожидалось %q", got, "1,4 км")
	}
	if got := FormatDistance(1000); got != "1,0 км" {
		t.Fatalf("FormatDistance(1000)=%q, ожидалось %q", got, "1,0 км")
	}
}

func TestGeoReplySortedAndComplete(t *testing.T) {
	reply := GeoReply(55.7988, 49.1052)

	for _, p := range data.Places {
		if !contains(reply.Text, p.Name) {
			t.Fatalf("в гео-ответе нет %q", p.Name)
		}
	}

	if idx := indexOf(reply.Text, "Казанский Кремль"); idx >= 0 {
		idxKaban := indexOf(reply.Text, "Озеро Кабан")
		if idx > idxKaban {
			t.Fatalf("Кремль должен идти раньше удалённого Кабана (сортировка по расстоянию)")
		}
	}

	if !contains(reply.Text, data.GeoNote) {
		t.Fatalf("в гео-ответе нет сноски")
	}
	if !hasButton(reply, "Назад") {
		t.Fatalf("в гео-ответе нет кнопки «Назад»")
	}
}

func TestHandleLocationReturnsGeoReply(t *testing.T) {
	replies := HandleLocation(55.7885, 49.1217)
	if len(replies) != 1 {
		t.Fatalf("ожидался 1 ответ гео-сервиса")
	}
	if !contains(replies[0].Text, "Улица Баумана") {
		t.Fatalf("в гео-ответе нет Баумана")
	}
}

func TestLocationButtonSendsTextThatMatchesSection(t *testing.T) {
	replies := Handle("меню")
	texts := buttonTexts(replies[0])
	if len(texts) == 0 {
		t.Fatalf("в меню нет кнопок")
	}
	foundGeo := false
	for _, row := range replies[0].Keyboard.Build().Payload.Buttons {
		for _, b := range row {
			if b.Text == data.GeoButton {
				if b.Type != model.ButtonRequestGeo {
					t.Fatalf("гео-кнопка должна быть request_geo_location, получено %q", b.Type)
				}
				foundGeo = true
			}
		}
	}
	if !foundGeo {
		t.Fatalf("в меню нет гео-кнопки")
	}
}

func contains(s, substr string) bool {
	return len(substr) == 0 || indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
