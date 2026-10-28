package main

import (
	"example.com/word-game/internal/quiz"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
	"time"
)

// A deformable grid moves the tail independently; the body also steps, sways and celebrates.
func (g *Game) fox(dst *ebiten.Image, pose string, x, y, w, h float64) {
	if !g.profile.Motion || g.paused {
		g.image(dst, "fox/"+pose, x, y, w, h)
		return
	}
	t := g.animationTime()
	bob := -math.Abs(math.Sin(t*2.4)) * 7
	stretch := 1 + math.Sin(t*2.4)*.018
	angle := math.Sin(t*1.8) * .035
	x += math.Sin(t*.85) * 12
	if g.screen == "result" {
		age := g.screenTime()
		if g.round.State == quiz.Won {
			bob = -math.Abs(math.Sin(age*4)) * 17
			angle = math.Sin(age*3) * .065
			x += math.Sin(age*2.5) * 9
		} else {
			bob = -math.Abs(math.Sin(age*2)) * 6
			angle = math.Sin(age*4) * .08 * math.Exp(-age*.8)
		}
	} else if g.feedback > 0 {
		age := time.Since(g.feedbackStarted).Seconds()
		if g.correct {
			bob -= math.Max(0, math.Sin(age*math.Pi*2.2)) * 14 * math.Exp(-age*1.3)
		} else {
			angle += math.Sin(age*7) * .05 * math.Exp(-age*2)
		}
	}
	im := g.sprite("fox/" + pose)
	iw, ih := float64(im.Bounds().Dx()), float64(im.Bounds().Dy())
	op := ebiten.GeoM{}
	op.Translate(-iw/2, -ih)
	op.Scale(w/iw, h*stretch/ih)
	op.Rotate(angle)
	op.Translate(x+w/2, y+h+bob)
	op.Scale(g.renderScale, g.renderScale)
	const n = 12
	vs := make([]ebiten.Vertex, (n+1)*(n+1))
	indices := make([]uint16, 0, n*n*6)
	for row := 0; row <= n; row++ {
		for col := 0; col <= n; col++ {
			u, v := float64(col)/n, float64(row)/n
			tail := math.Max(0, (u-.62)/.38)
			if pose == "celebrate" {
				tail = math.Max(0, (.36-u)/.36)
			}
			tail *= math.Max(0, math.Min(1, (v-.38)*5)) * math.Max(0, math.Min(1, (.96-v)*7))
			dx := math.Sin(t*4) * tail * iw * .028
			dy := math.Cos(t*3) * tail * ih * .015
			px, py := op.Apply(u*iw+dx, v*ih+dy)
			vs[row*(n+1)+col] = ebiten.Vertex{DstX: float32(px), DstY: float32(py), SrcX: float32(u * iw), SrcY: float32(v * ih), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1}
			if row < n && col < n {
				a := uint16(row*(n+1) + col)
				b := a + 1
				c := a + n + 1
				indices = append(indices, a, b, c, b, c+1, c)
			}
		}
	}
	dst.DrawTriangles(vs, indices, im, &ebiten.DrawTrianglesOptions{Filter: ebiten.FilterLinear})
}
