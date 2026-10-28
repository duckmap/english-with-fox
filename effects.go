package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"

	"image/color"
	"math"
	"strconv"
	"strings"
	"time"
)

func easeOut(t float64) float64 { t = math.Max(0, math.Min(1, t)); return 1 - math.Pow(1-t, 3) }
func (g *Game) updateEffects() {
	now := time.Now()
	dt := 1. / 60
	if !g.lastEffectsUpdate.IsZero() {
		dt = now.Sub(g.lastEffectsUpdate).Seconds()
	}
	g.lastEffectsUpdate = now
	if g.lastScreen != g.screen {
		g.lastScreen = g.screen
		g.screenAge = 0
		g.screenStarted = now
		g.hover = map[string]float64{}
	} else {
		g.screenAge++
	}
	if g.round != nil {
		if g.lastQuestion != g.round.Index {
			g.lastQuestion = g.round.Index
			g.questionAge = 0
			g.questionStarted = now
		} else if !g.paused {
			g.questionAge++
		}
	}
	if !g.profile.Motion {
		g.hover = map[string]float64{}
		g.displayProgress = 0
		if g.round != nil {
			g.displayProgress = float64(g.round.Index+1) / 30
		}
		return
	}
	x, y := g.cursor()
	for _, b := range g.buttons {
		target := 0.
		if !b.Disabled && float64(x) >= b.X && float64(x) < b.X+b.W && float64(y) >= b.Y && float64(y) < b.Y+b.H {
			target = 1
		}
		g.hover[b.ID] += (target - g.hover[b.ID]) * .22
	}
	if g.round != nil {
		target := float64(g.round.Index+1) / 30
		g.displayProgress += (target - g.displayProgress) * (1 - math.Exp(-dt*10))
	}
}
func (g *Game) animatedButton(b Button) Button {
	if !g.profile.Motion {
		return b
	}
	h := g.hover[b.ID]
	scale := 1 + .012*h
	dy := -3 * h
	x, y := g.cursor()
	pressed := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) && !b.Disabled && float64(x) >= b.X && float64(x) < b.X+b.W && float64(y) >= b.Y && float64(y) < b.Y+b.H
	if pressed {
		scale = .985
		dy = 2
	}
	// Topic cards enter in a short, row-major stagger; hit boxes remain stable.
	delay := 0.
	if strings.HasPrefix(b.ID, "topic:") {
		i, _ := strconv.Atoi(strings.TrimPrefix(b.ID, "topic:"))
		delay = float64(i) * .025
	}
	if !g.paused {
		dy += 12 * (1 - easeOut((g.screenTime()-delay)/.36))
	}
	b.X -= (b.W*scale - b.W) / 2
	b.Y += dy - (b.H*scale-b.H)/2
	b.W *= scale
	b.H *= scale
	return b
}
func (g *Game) focusMarker(dst *ebiten.Image, b Button, i int) {
	// A quiet dot is only shown after Tab, never a hover rectangle.
	if g.keyboardFocus && g.focus == i {
		g.fillCircle(dst, float32(b.X+b.W/2), float32(b.Y+b.H-13), 3, color.RGBA{30, 35, 67, 255}, true)
	}
}
func onOff(label string, on bool) string {
	state := "выкл."
	if on {
		state = "вкл."
	}
	return fmt.Sprintf("%s: %s", label, state)
}

func (g *Game) animationTime() float64 { return time.Since(g.startedAt).Seconds() }
func (g *Game) screenTime() float64 {
	if g.screenStarted.IsZero() {
		return 0
	}
	return time.Since(g.screenStarted).Seconds()
}
func (g *Game) questionTime() float64 {
	if g.questionStarted.IsZero() {
		return 0
	}
	return time.Since(g.questionStarted).Seconds()
}
