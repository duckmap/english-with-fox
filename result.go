package main

import (
	"bytes"
	"example.com/word-game/internal/quiz"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"image/color"
	"math"
)

func (g *Game) stopSFX() {
	if g.sfx != nil {
		g.sfx.Close()
		g.sfx = nil
	}
}
func (g *Game) playResultSound(won bool) {
	g.stopSFX()
	if !g.profile.Effects {
		return
	}
	name := "try_again"
	if won {
		name = "win"
	}
	b, e := assets.ReadFile("assets/sfx/" + name + ".wav")
	if e != nil {
		g.status = "Звук результата недоступен"
		return
	}
	stream, e := wav.DecodeWithSampleRate(44100, bytes.NewReader(b))
	if e != nil {
		return
	}
	g.sfx, e = g.audio.NewPlayer(stream)
	if e != nil {
		return
	}
	g.sfx.SetVolume(.55)
	g.duckMusic()
	g.sfx.Play()
	g.resultSoundCount++
}
func (g *Game) drawResult(s *ebiten.Image) {
	r := g.round
	won := r.State == quiz.Won
	title, sub, pose := "Отличная работа!", "Все 30 слов пройдены!", "celebrate"
	if !won {
		title = "Попробуем ещё раз?"
		sub = "Ошибаться — это часть учёбы!"
		pose = "wave"
	}
	age := g.screenTime()
	dy, titleSize, shownScore := 0., 48., r.Score
	if g.profile.Motion {
		dy = 18 * (1 - easeOut(age/.55))
		titleSize += 4 * math.Sin(math.Min(1, age/.7)*math.Pi)
		shownScore = int(math.Round(float64(r.Score) * easeOut(age/.9)))
		g.resultParticles(s, won, age)
	}
	g.center(s, title, 160, 55, 960, titleSize, navy)
	g.center(s, sub, 160, 122, 960, 25, muted)
	g.fox(s, pose, 530, 165, 220, 220)
	if won {
		ty := 210.
		if g.profile.Motion {
			ty -= 7 * math.Abs(math.Sin(age*2.3))
		}
		g.image(s, "ui/trophy", 385, ty, 150, 150)
	}
	g.image(s, "ui/panel", 300, 382+dy, 680, 232)
	g.center(s, "Тема: "+r.Topic.Name, 310, 402+dy, 660, 23, muted)
	g.center(s, fmt.Sprintf("Баллы: %d", shownScore), 310, 443+dy, 660, 50, navy)
	g.center(s, fmt.Sprintf("Верно: %d   •   Ошибок: %d", r.Correct, r.Wrong), 310, 512+dy, 660, 28, navy)
	if won {
		g.center(s, "Результат сохранён в рейтинге", 310, 559+dy, 660, 23, muted)
	} else {
		g.center(s, "Новая попытка — 10 баллов", 310, 559+dy, 660, 23, muted)
	}
}
func (g *Game) resultParticles(s *ebiten.Image, won bool, age float64) {
	if !won {
		for i := 0; i < 5; i++ {
			phase := math.Mod(age*.45+float64(i)*.2, 1)
			x := 405 + float64(i)*110
			y := 345 - phase*155
			g.image(s, "ui/star", x, y, 12+6*math.Sin(phase*math.Pi), 12+6*math.Sin(phase*math.Pi))
		}
		return
	}
	colors := []color.RGBA{{82, 167, 230, 255}, {240, 190, 67, 255}, {98, 183, 130, 255}, {232, 139, 110, 255}}
	for i := 0; i < 44; i++ {
		t := age - float64(i%4)*.12
		if t < 0 || t > 3.8 {
			continue
		}
		side := -1.
		x0 := 980.
		if i%2 == 0 {
			side = 1
			x0 = 300
		}
		vx := side * (20 + float64(i%9)*12)
		vy := -85 - float64(i%7)*11
		x := x0 + vx*t
		y := 330 + vy*t + 58*t*t
		if y < 170 || y > 355 || x < 160 || x > 1120 {
			continue
		}
		c := colors[i%4]
		c.A = uint8(255 * math.Min(1, (3.8-t)/.6))
		g.fillRect(s, float32(x), float32(y), float32(5+i%3), float32(9+i%4), color.NRGBA{c.R, c.G, c.B, c.A}, true)
	}
	if age > 3 {
		for i := 0; i < 8; i++ {
			g.image(s, "ui/star", 230+float64(i)*115, 205+float64(i%3)*40+8*math.Sin(age+float64(i)), 14, 14)
		}
	}
}
