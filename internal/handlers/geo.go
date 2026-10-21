package handlers

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/FodinNn/TourBot/internal/data"
)

const (
	earthRadiusM = 6371000.0
	nearbyCount  = 3
)

func DistanceMeters(lat1, lon1, lat2, lon2 float64) int {
	phi1 := lat1 * math.Pi / 180
	phi2 := lat2 * math.Pi / 180
	dPhi := (lat2 - lat1) * math.Pi / 180
	dLambda := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dPhi/2)*math.Sin(dPhi/2) +
		math.Cos(phi1)*math.Cos(phi2)*math.Sin(dLambda/2)*math.Sin(dLambda/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return int(earthRadiusM * c)
}

func FormatDistance(m int) string {
	if m < 1000 {
		return strconv.Itoa(roundToTens(m)) + " м"
	}
	km := float64(m) / 1000
	return strings.Replace(strconv.FormatFloat(km, 'f', 1, 64), ".", ",", 1) + " км"
}

func roundToTens(m int) int {
	return int(math.Round(float64(m)/10)) * 10
}

type nearbyPlace struct {
	place data.Place
	dist  int
}

func HandleLocation(lat, lon float64) []Reply {
	return []Reply{GeoReply(lat, lon)}
}

func GeoReply(lat, lon float64) Reply {
	nearby := make([]nearbyPlace, 0, len(data.Places))
	for _, p := range data.Places {
		nearby = append(nearby, nearbyPlace{
			place: p,
			dist:  DistanceMeters(lat, lon, p.Lat, p.Lon),
		})
	}

	sort.Slice(nearby, func(i, j int) bool {
		return nearby[i].dist < nearby[j].dist
	})

	if len(nearby) > nearbyCount {
		nearby = nearby[:nearbyCount]
	}

	var sb strings.Builder
	sb.WriteString("📍 Вот что поблизости:\n\n")
	for i, n := range nearby {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(n.place.Emoji)
		sb.WriteString(" **")
		sb.WriteString(n.place.Name)
		sb.WriteString("** — ")
		sb.WriteString(FormatDistance(n.dist))
		sb.WriteString("\n")
		sb.WriteString(n.place.Description)
		sb.WriteString("\n")
	}

	return Reply{
		Text:     sb.String() + "\n" + data.GeoNote,
		Keyboard: BackKeyboard(),
		Markdown: true,
	}
}
