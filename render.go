package main

import (
	"example.com/word-game/internal/view"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
	"math"
	"strings"
)

func (g *Game) Layout(outW, outH int) (int, int) {
	oldScale := g.renderScale
	g.renderScale, g.renderW, g.renderH = view.Fit(outW, outH, ebiten.Monitor().DeviceScaleFactor())
	if oldScale != g.renderScale {
		g.panelCache = map[string]*ebiten.Image{}
	}
	return g.renderW, g.renderH
}
func (g *Game) cursor() (float64, float64) {
	x, y := ebiten.CursorPosition()
	return float64(x) / g.renderScale, float64(y) / g.renderScale
}
func (g *Game) fillRect(s *ebiten.Image, x, y, w, h float32, c color.Color, aa bool) {
	k := float32(g.renderScale)
	vector.DrawFilledRect(s, x*k, y*k, w*k, h*k, c, aa)
}
func (g *Game) fillCircle(s *ebiten.Image, x, y, r float32, c color.Color, aa bool) {
	k := float32(g.renderScale)
	vector.DrawFilledCircle(s, x*k, y*k, r*k, c, aa)
}
func (g *Game) roundRect(s *ebiten.Image, x, y, w, h, r float64, c color.Color) {
	k := g.renderScale
	x *= k
	y *= k
	w *= k
	h *= k
	r *= k
	r = math.Min(r, math.Min(w, h)/2)
	var p vector.Path
	p.MoveTo(float32(x+r), float32(y))
	p.LineTo(float32(x+w-r), float32(y))
	p.QuadTo(float32(x+w), float32(y), float32(x+w), float32(y+r))
	p.LineTo(float32(x+w), float32(y+h-r))
	p.QuadTo(float32(x+w), float32(y+h), float32(x+w-r), float32(y+h))
	p.LineTo(float32(x+r), float32(y+h))
	p.QuadTo(float32(x), float32(y+h), float32(x), float32(y+h-r))
	p.LineTo(float32(x), float32(y+r))
	p.QuadTo(float32(x), float32(y), float32(x+r), float32(y))
	p.Close()
	if g.solidPixel == nil {
		g.solidPixel = ebiten.NewImage(3, 3)
		g.solidPixel.Fill(color.White)
	}
	vs, is := p.AppendVerticesAndIndicesForFilling(nil, nil)
	r16, g16, b16, a16 := c.RGBA()
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = float32(r16) / 65535
		vs[i].ColorG = float32(g16) / 65535
		vs[i].ColorB = float32(b16) / 65535
		vs[i].ColorA = float32(a16) / 65535
	}
	s.DrawTriangles(vs, is, g.solidPixel, &ebiten.DrawTrianglesOptions{AntiAlias: true, ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha})
}
func (g *Game) rawPanel(s *ebiten.Image, name string, x, y, w, h float64) bool {
	if !(strings.HasPrefix(name, "ui/button_") || name == "ui/panel" || name == "ui/topic_card") {
		return false
	}
	fill := color.RGBA{229, 244, 248, 255}
	switch name {
	case "ui/button_primary", "ui/button_round_primary":
		fill = color.RGBA{100, 184, 243, 255}
	case "ui/button_pressed":
		fill = color.RGBA{57, 151, 221, 255}
	case "ui/button_correct":
		fill = color.RGBA{188, 233, 190, 255}
	case "ui/button_incorrect":
		fill = color.RGBA{255, 208, 192, 255}
	case "ui/button_disabled":
		fill = color.RGBA{218, 223, 226, 255}
	case "ui/panel":
		fill = cream
	case "ui/topic_card":
		fill = color.RGBA{217, 240, 251, 255}
	}
	shadow := color.NRGBA{120, 120, 100, 70}
	if strings.Contains(name, "round") {
		r := float32(math.Min(w, h)/2 - 5)
		cx, cy := float32(x+w/2), float32(y+h/2-2)
		g.fillCircle(s, cx, cy+5, r, shadow, true)
		g.fillCircle(s, cx, cy, r, navy, true)
		g.fillCircle(s, cx, cy, r-3, fill, true)
		return true
	}
	r := math.Min(24, h*.27)
	g.roundRect(s, x+3, y+8, w-6, h-10, r, shadow)
	g.roundRect(s, x+3, y+2, w-6, h-12, r, navy)
	g.roundRect(s, x+6, y+5, w-12, h-18, r-3, fill)
	return true
}

// Linear final filtering is a safeguard only; text/vector geometry already render at native pixel density.
func (g *Game) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	screen.Fill(cream)
	op := &ebiten.DrawImageOptions{GeoM: geoM, Filter: ebiten.FilterLinear}
	screen.DrawImage(offscreen, op)
}

// Cache vector UI at CURRENT physical pixel density, never at a fixed 1280x800.
// This keeps crisp native edges while avoiding re-tessellating every button on every frame.
func (g *Game) panel(dst *ebiten.Image, name string, x, y, w, h float64) bool {
	if !(strings.HasPrefix(name, "ui/button_") || name == "ui/panel" || name == "ui/topic_card") {
		return false
	}
	rw, rh := int(math.Ceil(w*g.renderScale)), int(math.Ceil(h*g.renderScale))
	key := fmt.Sprintf("%s:%d:%d", name, rw, rh)
	im := g.panelCache[key]
	if im == nil {
		im = ebiten.NewImage(rw, rh)
		g.rawPanel(im, name, 0, 0, float64(rw)/g.renderScale, float64(rh)/g.renderScale)
		g.panelCache[key] = im
	}
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(math.Round(x*g.renderScale), math.Round(y*g.renderScale))
	dst.DrawImage(im, op)
	return true
}
