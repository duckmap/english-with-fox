// Package quiz contains the game rules without graphics, audio or storage.
package quiz

import (
	"fmt"
	"math/rand/v2"
	"time"
)

type Word struct {
	EN string `json:"en"`
	RU string `json:"ru"`
}
type Topic struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Words []Word `json:"words"`
}
type State string

const (
	Playing State = "playing"
	Won     State = "won"
	Lost    State = "lost"
)

type Round struct {
	Topic                        Topic
	Words                        []Word
	Index, Score, Correct, Wrong int
	Options                      []string
	State                        State
	answered                     bool
	rng                          *rand.Rand
}

func Validate(topics []Topic) error {
	ids := map[string]bool{}
	for _, t := range topics {
		if t.ID == "" || ids[t.ID] || t.Name == "" {
			return fmt.Errorf("invalid topic %q", t.ID)
		}
		ids[t.ID] = true
		if len(t.Words) != 30 {
			return fmt.Errorf("%s: expected 30 words", t.ID)
		}
		en, ru := map[string]bool{}, map[string]bool{}
		for _, w := range t.Words {
			if w.EN == "" || w.RU == "" || en[w.EN] || ru[w.RU] {
				return fmt.Errorf("%s: duplicate or empty word", t.ID)
			}
			en[w.EN] = true
			ru[w.RU] = true
		}
	}
	return nil
}
func New(t Topic, rng *rand.Rand) *Round {
	r := &Round{Topic: t, Words: append([]Word(nil), t.Words...), Score: 10, State: Playing, rng: rng}
	rng.Shuffle(len(r.Words), func(i, j int) { r.Words[i], r.Words[j] = r.Words[j], r.Words[i] })
	r.options()
	return r
}
func (r *Round) Current() Word { return r.Words[r.Index] }
func (r *Round) options() {
	wrong := []string{}
	for _, w := range r.Topic.Words {
		if w.RU != r.Current().RU {
			wrong = append(wrong, w.RU)
		}
	}
	r.rng.Shuffle(len(wrong), func(i, j int) { wrong[i], wrong[j] = wrong[j], wrong[i] })
	r.Options = append([]string{r.Current().RU}, wrong[:4]...)
	r.rng.Shuffle(5, func(i, j int) { r.Options[i], r.Options[j] = r.Options[j], r.Options[i] })
}

// Answer scores once; the UI locks input until Advance, to prevent double clicks.
func (r *Round) Answer(option int) (bool, error) {
	if r.answered || r.State != Playing || option < 0 || option >= len(r.Options) {
		return false, fmt.Errorf("invalid answer")
	}
	r.answered = true
	ok := r.Options[option] == r.Current().RU
	if ok {
		r.Score++
		r.Correct++
	} else {
		r.Score -= 2
		r.Wrong++
	}
	if r.Score <= 0 {
		r.Score = 0
		r.State = Lost
	} else if r.Correct+r.Wrong == len(r.Words) {
		r.State = Won
	}
	return ok, nil
}
func (r *Round) Advance() {
	if r.State == Playing && r.answered {
		r.Index++
		r.answered = false
		r.options()
	}
}

type Result struct {
	Name    string    `json:"name"`
	Topic   string    `json:"topic"`
	Score   int       `json:"score"`
	Correct int       `json:"correct"`
	Wrong   int       `json:"wrong"`
	Won     bool      `json:"won"`
	Date    time.Time `json:"date"`
}
