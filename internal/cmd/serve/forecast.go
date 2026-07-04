// internal/cmd/serve/forecast.go
package serve

import (
	"sort"
	"time"

	"github.com/asano69/hashcards/internal/collection"
	"github.com/asano69/hashcards/internal/db"
	"github.com/asano69/hashcards/internal/types"
)

// Forecast horizons: how many future buckets each chart covers.
const (
	forecastDailyDays     = 7
	forecastWeeklyWeeks   = 12
	forecastMonthlyMonths = 12
)

// forecastBucket is one point on a forecast chart: a time bucket label and
// the number of cards due in that bucket.
type forecastBucket struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

// deckForecast holds the three forecast charts (daily/weekly/monthly) for
// a single deck.
type deckForecast struct {
	Deck    string           `json:"deck"`
	Daily   []forecastBucket `json:"daily"`
	Weekly  []forecastBucket `json:"weekly"`
	Monthly []forecastBucket `json:"monthly"`
}

// buildForecast computes, for every deck in the collection, how many cards
// are scheduled to come due over the next 7 days, 12 weeks, and 12 months.
// Overdue cards and new (never-reviewed) cards are excluded: this is a
// forward-looking forecast of the existing review schedule, not a count of
// what's due right now (see collection.Collection.DueToday for that).
func buildForecast(col *collection.Collection, database *db.Database) ([]deckForecast, error) {
	today := types.Today()

	deckNames, cardsByDeck := groupCardsByDeck(col.Cards)

	forecasts := make([]deckForecast, 0, len(deckNames))
	for _, name := range deckNames {
		df := deckForecast{
			Deck:    name,
			Daily:   newDailyBuckets(today),
			Weekly:  newWeeklyBuckets(today),
			Monthly: newMonthlyBuckets(today),
		}
		for _, card := range cardsByDeck[name] {
			perf, err := database.GetCardPerformance(card.Hash())
			if err != nil {
				return nil, err
			}
			if perf.IsNew() {
				continue
			}
			addToBuckets(&df, today, perf.Reviewed().DueDate)
		}
		forecasts = append(forecasts, df)
	}
	return forecasts, nil
}

// groupCardsByDeck groups cards by deck name, returning the deck names in
// sorted order alongside the grouping.
func groupCardsByDeck(cards []types.Card) ([]string, map[string][]types.Card) {
	byDeck := make(map[string][]types.Card)
	for _, c := range cards {
		byDeck[c.DeckName()] = append(byDeck[c.DeckName()], c)
	}
	names := make([]string, 0, len(byDeck))
	for name := range byDeck {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, byDeck
}

func newDailyBuckets(today types.Date) []forecastBucket {
	buckets := make([]forecastBucket, forecastDailyDays)
	for i := range buckets {
		buckets[i].Label = types.NewDate(today.Time().AddDate(0, 0, i)).String()
	}
	return buckets
}

func newWeeklyBuckets(today types.Date) []forecastBucket {
	buckets := make([]forecastBucket, forecastWeeklyWeeks)
	for i := range buckets {
		buckets[i].Label = types.NewDate(today.Time().AddDate(0, 0, 7*i)).String()
	}
	return buckets
}

func newMonthlyBuckets(today types.Date) []forecastBucket {
	buckets := make([]forecastBucket, forecastMonthlyMonths)
	for i := range buckets {
		buckets[i].Label = today.Time().AddDate(0, i, 0).Format("2006-01")
	}
	return buckets
}

// addToBuckets increments the daily/weekly/monthly bucket that dueDate
// falls into, relative to today. Cards due in the past are ignored.
func addToBuckets(df *deckForecast, today, dueDate types.Date) {
	days := int(dueDate.Time().Sub(today.Time()) / (24 * time.Hour))
	if days < 0 {
		return
	}

	if days < forecastDailyDays {
		df.Daily[days].Count++
	}
	if week := days / 7; week < forecastWeeklyWeeks {
		df.Weekly[week].Count++
	}
	if month := monthsBetween(today.Time(), dueDate.Time()); month >= 0 && month < forecastMonthlyMonths {
		df.Monthly[month].Count++
	}
}

// monthsBetween returns the number of calendar months between from and to
// (e.g. January to March is 2), which may be negative if to precedes from.
func monthsBetween(from, to time.Time) int {
	return (to.Year()-from.Year())*12 + int(to.Month()) - int(from.Month())
}
