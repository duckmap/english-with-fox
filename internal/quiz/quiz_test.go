package quiz

import (
	"encoding/json"
	"math/rand/v2"
	"os"
	"testing"
)

func topics(t *testing.T) []Topic {
	t.Helper()
	b, e := os.ReadFile("../../assets/data/topics.json")
	if e != nil {
		t.Fatal(e)
	}
	var ts []Topic
	if e = json.Unmarshal(b, &ts); e != nil {
		t.Fatal(e)
	}
	if e = Validate(ts); e != nil {
		t.Fatal(e)
	}
	return ts
}
func round(t *testing.T) *Round { return New(topics(t)[0], rand.New(rand.NewPCG(1, 2))) }
func pick(r *Round, correct bool) int {
	for i, s := range r.Options {
		if (s == r.Current().RU) == correct {
			return i
		}
	}
	panic("answer not found")
}
func TestDictionary(t *testing.T) {
	ts := topics(t)
	if len(ts) != 12 {
		t.Fatal(len(ts))
	}
	for _, v := range ts {
		if len(v.Words) != 30 {
			t.Fatal(v.ID)
		}
	}
}
func TestOptionsAndPerfectGame(t *testing.T) {
	for seed := uint64(0); seed < 80; seed++ {
		r := New(topics(t)[0], rand.New(rand.NewPCG(seed, 1)))
		seen := map[string]bool{}
		for n := 0; n < 30; n++ {
			if len(r.Options) != 5 {
				t.Fatal("not five options")
			}
			unique := map[string]bool{}
			for _, s := range r.Options {
				if unique[s] {
					t.Fatal("duplicate option")
				}
				unique[s] = true
			}
			if !unique[r.Current().RU] {
				t.Fatal("missing correct answer")
			}
			if seen[r.Current().EN] {
				t.Fatal("repeated word")
			}
			seen[r.Current().EN] = true
			if ok, e := r.Answer(pick(r, true)); !ok || e != nil {
				t.Fatal(e)
			}
			r.Advance()
		}
		if r.Score != 40 || r.State != Won || r.Correct != 30 || r.Wrong != 0 {
			t.Fatalf("bad perfect game: %+v", r)
		}
	}
}
func TestLostAfterFiveErrors(t *testing.T) {
	r := round(t)
	for i := 0; i < 5; i++ {
		ok, e := r.Answer(pick(r, false))
		if ok || e != nil {
			t.Fatal(e)
		}
		r.Advance()
	}
	if r.Score != 0 || r.State != Lost || r.Wrong != 5 {
		t.Fatalf("%+v", r)
	}
}
func TestDoubleAnswerAndAdvanceGuard(t *testing.T) {
	r := round(t)
	r.Advance()
	if r.Index != 0 {
		t.Fatal("advanced unanswered question")
	}
	i := pick(r, true)
	r.Answer(i)
	if _, e := r.Answer(i); e == nil {
		t.Fatal("double answer accepted")
	}
	if r.Score != 11 || r.Correct != 1 {
		t.Fatal("double score")
	}
	r.Advance()
	r.Advance()
	if r.Index != 1 {
		t.Fatal("double advance")
	}
}
func TestInvalidAnswer(t *testing.T) {
	r := round(t)
	for _, i := range []int{-1, 5, 99} {
		if _, e := r.Answer(i); e == nil {
			t.Fatal("invalid accepted")
		}
	}
	if r.Score != 10 {
		t.Fatal("score changed")
	}
}
func TestVictoryScoreExample(t *testing.T) {
	r := round(t)
	for n := 0; n < 30; n++ {
		r.Answer(pick(r, n != 2 && n != 8))
		r.Advance()
	}
	if r.State != Won || r.Score != 34 || r.Correct != 28 || r.Wrong != 2 {
		t.Fatalf("%+v", r)
	}
}
func TestScoreClampedAtZero(t *testing.T) {
	r := round(t)
	r.Score = 1
	r.Answer(pick(r, false))
	if r.State != Lost || r.Score != 0 {
		t.Fatal("negative score")
	}
}
func TestFinishedRoundCannotChange(t *testing.T) {
	r := round(t)
	r.Score = 2
	r.Answer(pick(r, false))
	r.Advance()
	if _, e := r.Answer(0); e == nil {
		t.Fatal("answer after finish")
	}
	if r.Index != 0 || r.Score != 0 {
		t.Fatal("finished state mutated")
	}
}
func TestShuffleDoesNotModifyTopic(t *testing.T) {
	ts := topics(t)
	first := ts[0].Words[0]
	New(ts[0], rand.New(rand.NewPCG(1, 2)))
	if ts[0].Words[0] != first {
		t.Fatal("topic mutated")
	}
}
