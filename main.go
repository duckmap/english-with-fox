package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"math/rand/v2"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"example.com/word-game/internal/quiz"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets
var assets embed.FS

const W, H = 1280, 800

var navy = color.RGBA{30, 35, 67, 255}
var blue = color.RGBA{68, 156, 218, 255}
var muted = color.RGBA{78, 91, 109, 255}
var cream = color.RGBA{255, 249, 233, 255}

type Profile struct {
	Name    string        `json:"name"`
	Effects bool          `json:"effects"`
	Music   bool          `json:"music"`
	Motion  bool          `json:"motion"`
	Sound   bool          `json:"sound"`
	Results []quiz.Result `json:"results"`
}
type Button struct {
	ID, Label, Sprite, Icon string
	X, Y, W, H              float64
	Disabled                bool
}
type Game struct {
	solidPixel                                                                                                   *ebiten.Image
	topics                                                                                                       []quiz.Topic
	sprites                                                                                                      map[string]*ebiten.Image
	panelCache                                                                                                   map[string]*ebiten.Image
	startedAt, screenStarted, questionStarted, feedbackStarted, pauseStarted, lastEffectsUpdate, lastMusicUpdate time.Time
	font                                                                                                         *text.GoTextFaceSource
	audio                                                                                                        *audio.Context
	player                                                                                                       *audio.Player
	audioIndex                                                                                                   map[string]string
	music                                                                                                        *audio.Player
	sfx                                                                                                          *audio.Player
	resultSoundCount                                                                                             int
	renderScale                                                                                                  float64
	renderW, renderH                                                                                             int
	musicVolume                                                                                                  float64
	hover                                                                                                        map[string]float64
	keyboardFocus                                                                                                bool
	lastScreen                                                                                                   string
	screenAge, questionAge, lastQuestion                                                                         int
	displayProgress                                                                                              float64
	profile                                                                                                      Profile
	screen                                                                                                       string
	round                                                                                                        *quiz.Round
	topic                                                                                                        int
	buttons                                                                                                      []Button
	focus                                                                                                        int
	editName                                                                                                     bool
	nameDraft                                                                                                    string
	feedback                                                                                                     int
	selected                                                                                                     int
	correct                                                                                                      bool
	paused                                                                                                       bool
	tick                                                                                                         int
	rankPage                                                                                                     int
	status, audioError                                                                                           string
}

func newGame() (*Game, error) {
	g := &Game{sprites: map[string]*ebiten.Image{}, profile: Profile{Name: "Игрок", Sound: true, Music: true, Effects: true, Motion: !reducedMotionRequested()}, screen: "menu", focus: -1, selected: -1, renderScale: 1, renderW: W, renderH: H, audio: audio.NewContext(44100)}
	b, e := assets.ReadFile("assets/data/topics.json")
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(b, &g.topics); e != nil {
		return nil, e
	}
	if e = quiz.Validate(g.topics); e != nil {
		return nil, e
	}
	b, e = assets.ReadFile("assets/fonts/game.ttf")
	if e != nil {
		return nil, e
	}
	g.font, e = text.NewGoTextFaceSource(bytes.NewReader(b))
	if e != nil {
		return nil, e
	}
	b, e = assets.ReadFile("assets/audio/index.json")
	if e != nil {
		return nil, e
	}
	if e = json.Unmarshal(b, &g.audioIndex); e != nil {
		return nil, e
	}
	b, e = loadSave()
	if e != nil {
		g.status = "Не удалось прочитать сохранение. Результаты пока не загружены."
	} else if len(b) > 0 {
		if e = json.Unmarshal(b, &g.profile); e != nil {
			g.profile = Profile{Name: "Игрок", Sound: true, Music: true, Effects: true, Motion: !reducedMotionRequested()}
			g.status = "Сохранение повреждено. Игра начнёт новую таблицу."
		}
	}
	g.startedAt = time.Now()
	g.panelCache = map[string]*ebiten.Image{}
	g.hover = map[string]float64{}
	g.lastQuestion = -1
	if err := g.initMusic(); err != nil {
		g.status = "Не удалось включить фоновую музыку."
	}
	g.profile.Name = cleanName(g.profile.Name)
	g.nameDraft = g.profile.Name
	return g, nil
}
func cleanName(s string) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) > 18 {
		s = string(r[:18])
	}
	if s == "" {
		return "Игрок"
	}
	return s
}
func (g *Game) save() {
	b, e := json.MarshalIndent(g.profile, "", "  ")
	if e == nil {
		e = writeSave(b)
	}
	if e != nil {
		g.status = "Не удалось сохранить результат: проверь доступ к хранилищу."
	}
}
func (g *Game) sprite(name string) *ebiten.Image {
	if im := g.sprites[name]; im != nil {
		return im
	}
	b, e := assets.ReadFile("assets/" + name + ".png")
	if e != nil {
		panic(e)
	}
	im, _, e := image.Decode(bytes.NewReader(b))
	if e != nil {
		panic(e)
	}
	out := ebiten.NewImageFromImage(im)
	g.sprites[name] = out
	return out
}
func (g *Game) image(dst *ebiten.Image, name string, x, y, w, h float64) {
	if g.panel(dst, name, x, y, w, h) {
		return
	}
	im := g.sprite(name)
	x *= g.renderScale
	y *= g.renderScale
	w *= g.renderScale
	h *= g.renderScale
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(w/float64(im.Bounds().Dx()), h/float64(im.Bounds().Dy()))
	op.GeoM.Translate(x, y)
	op.Filter = ebiten.FilterLinear
	dst.DrawImage(im, op)
}
func (g *Game) measure(s string, size float64) (float64, float64) {
	return text.Measure(s, &text.GoTextFace{Source: g.font, Size: size}, size*1.2)
}
func (g *Game) label(dst *ebiten.Image, s string, x, y, size float64, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x*g.renderScale, y*g.renderScale)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, &text.GoTextFace{Source: g.font, Size: size * g.renderScale}, op)
}
func (g *Game) center(dst *ebiten.Image, s string, x, y, w, size float64, c color.Color) {
	tw, _ := g.measure(s, size)
	g.label(dst, s, x+(w-tw)/2, y, size, c)
}
func (g *Game) fit(dst *ebiten.Image, s string, x, y, w, size float64, c color.Color) {
	for size > 15 {
		tw, _ := g.measure(s, size)
		if tw <= w-24 {
			break
		}
		size--
	}
	g.center(dst, s, x, y, w, size, c)
}
func (g *Game) stopAudio() {
	if g.player != nil {
		g.player.Close()
		g.player = nil
	}
}
func (g *Game) speak() {
	g.stopAudio()
	if !g.profile.Sound || g.round == nil {
		return
	}
	name, ok := g.audioIndex[g.round.Current().EN]
	if !ok {
		g.audioError = "Запись этого слова недоступна"
		return
	}
	b, e := assets.ReadFile("assets/audio/" + name)
	if e != nil {
		g.audioError = "Не удалось открыть запись"
		return
	}
	stream, e := wav.DecodeWithSampleRate(44100, bytes.NewReader(b))
	if e != nil {
		g.audioError = "Не удалось прочитать звук"
		return
	}
	g.player, e = g.audio.NewPlayer(stream)
	if e != nil {
		g.audioError = "Не удалось включить звук"
		return
	}
	g.duckMusic()
	g.player.SetVolume(.85)
	g.player.Play()
	g.audioError = ""
}
func (g *Game) start(i int) {
	g.stopSFX()
	g.profile.Name = cleanName(g.nameDraft)
	g.nameDraft = g.profile.Name
	g.save()
	g.topic = i
	g.round = quiz.New(g.topics[i], rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(g.tick))))
	g.screen = "play"
	g.feedback = 0
	g.questionAge = 0
	g.lastQuestion = -1
	g.displayProgress = 0
	g.selected = -1
	g.paused = false
	g.editName = false
	g.focus = -1
	g.speak()
}
func (g *Game) finish() {
	r := g.round
	g.screen = "result"
	g.playResultSound(r.State == quiz.Won)
	g.focus = -1
	g.stopAudio()
	g.profile.Results = append(g.profile.Results, quiz.Result{Name: g.profile.Name, Topic: r.Topic.Name, Score: r.Score, Correct: r.Correct, Wrong: r.Wrong, Won: r.State == quiz.Won, Date: time.Now()})
	sort.SliceStable(g.profile.Results, func(i, j int) bool {
		a, b := g.profile.Results[i], g.profile.Results[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.Correct != b.Correct {
			return a.Correct > b.Correct
		}
		return a.Date.After(b.Date)
	})
	if len(g.profile.Results) > 50 {
		g.profile.Results = g.profile.Results[:50]
	}
	g.save()
}
func (g *Game) answer(i int) {
	if g.screen != "play" || g.paused || g.feedback > 0 {
		return
	}
	ok, e := g.round.Answer(i)
	if e != nil {
		return
	}
	g.correct = ok
	g.selected = i
	g.feedback = 90
	g.feedbackStarted = time.Now()
	g.stopAudio()
}
func (g *Game) add(id, label, sprite, icon string, x, y, w, h float64) {
	g.buttons = append(g.buttons, Button{ID: id, Label: label, Sprite: sprite, Icon: icon, X: x, Y: y, W: w, H: h})
}
func (g *Game) buildButtons() {
	g.buttons = nil
	switch g.screen {
	case "menu":
		g.add("rank", "Рейтинг", "button_normal", "trophy", 1020, 28, 228, 76)
		g.add("motion", onOff("Движение", g.profile.Motion), "button_normal", "", 760, 28, 228, 76)
		g.add("effects", onOff("Эффекты", g.profile.Effects), "button_normal", "", 760, 110, 228, 56)
		g.add("music", onOff("Музыка", g.profile.Music), "button_normal", "music", 1020, 110, 228, 56)
		for i, t := range g.topics {
			g.add(fmt.Sprintf("topic:%d", i), t.Name, "topic_card", "", 32+float64(i%4)*310, 180+float64(i/4)*150, 286, 134)
		}
		g.add("name", "", "button_normal", "", 32, 662, 510, 72)
		s := "Голос: вкл."
		if !g.profile.Sound {
			s = "Голос: выкл."
		}
		g.add("sound", s, "button_normal", "speaker", 1010, 662, 238, 72)
	case "play":
		g.add("pause", "Темы", "button_normal", "back", 24, 24, 184, 72)
		g.add("sound", "", "button_normal", "speaker", 1170, 24, 82, 72)
		g.add("speak", "", "button_primary", "speaker", 876, 235, 110, 110)
		for i, s := range g.round.Options {
			x := float64(100 + (i%3)*370)
			y := float64(454 + (i/3)*120)
			if i >= 3 {
				x = 285 + float64(i-3)*370
			}
			sp := "button_normal"
			if g.feedback > 0 {
				if s == g.round.Current().RU {
					sp = "button_correct"
				} else if i == g.selected {
					sp = "button_incorrect"
				}
			}
			g.add(fmt.Sprintf("answer:%d", i), s, sp, "", x, y, 340, 100)
			g.buttons[len(g.buttons)-1].Disabled = g.feedback > 0
		}
		if g.paused {
			g.buttons = nil
			g.add("resume", "Продолжить", "button_primary", "", 460, 348, 360, 88)
			g.add("leave", "Выйти к темам", "button_normal", "", 460, 454, 360, 88)
			g.add("music", onOff("Музыка", g.profile.Music), "button_normal", "", 420, 550, 210, 56)
			g.add("motion", onOff("Движение", g.profile.Motion), "button_normal", "", 650, 550, 210, 56)
			g.add("effects", onOff("Эффекты", g.profile.Effects), "button_normal", "", 420, 612, 210, 56)
			g.add("sound", onOff("Голос", g.profile.Sound), "button_normal", "", 650, 612, 210, 56)
		}
	case "result":
		g.add("menu", "Выбрать тему", "button_primary", "", 275, 656, 350, 90)
		g.add("again", "Ещё раз", "button_normal", "", 655, 656, 350, 90)
		g.add("rank", "Рейтинг", "button_normal", "trophy", 1020, 24, 228, 72)
	case "rank":
		g.add("menu", "Темы", "button_normal", "back", 24, 24, 184, 72)
		if g.rankPage > 0 {
			g.add("prev", "Назад", "button_normal", "", 340, 694, 260, 72)
		}
		if (g.rankPage+1)*7 < len(g.profile.Results) {
			g.add("next", "Дальше", "button_normal", "", 680, 694, 260, 72)
		}
	}
}
func (g *Game) activate(id string) {
	switch {
	case strings.HasPrefix(id, "topic:"):
		var i int
		fmt.Sscanf(id, "topic:%d", &i)
		if i >= 0 && i < len(g.topics) {
			g.start(i)
		}
	case strings.HasPrefix(id, "answer:"):
		var i int
		fmt.Sscanf(id, "answer:%d", &i)
		g.answer(i)
	case id == "menu" || id == "leave":
		g.stopSFX()
		g.stopAudio()
		g.feedback = 0
		g.screen = "menu"
		g.paused = false
		g.focus = -1
	case id == "rank":
		g.stopSFX()
		g.stopAudio()
		g.screen = "rank"
		g.rankPage = 0
		g.focus = -1
	case id == "again":
		g.start(g.topic)
	case id == "sound":
		g.profile.Sound = !g.profile.Sound
		if !g.profile.Sound {
			g.stopAudio()
		}
		g.save()
	case id == "effects":
		g.profile.Effects = !g.profile.Effects
		if !g.profile.Effects {
			g.stopSFX()
		}
		g.save()
	case id == "music":
		g.profile.Music = !g.profile.Music
		g.save()
	case id == "motion":
		g.profile.Motion = !g.profile.Motion
		g.save()
	case id == "speak":
		g.speak()
	case id == "pause":
		g.pauseStarted = time.Now()
		g.paused = true
		g.stopAudio()
		g.focus = -1
	case id == "resume":
		if g.paused && g.feedback > 0 {
			g.feedbackStarted = g.feedbackStarted.Add(time.Since(g.pauseStarted))
		}
		g.paused = false
		g.focus = -1
		g.speak()
	case id == "name":
		g.editName = true
		g.focus = -1
	case id == "next":
		g.rankPage++
		g.focus = -1
	case id == "prev":
		g.rankPage--
		g.focus = -1
	}
}
func (g *Game) Update() error {
	g.tick++
	if len(g.panelCache) > 192 {
		g.panelCache = map[string]*ebiten.Image{}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	defer g.updateEffects()
	g.updateMusic()
	if g.feedback > 0 && !g.paused {
		elapsed := time.Since(g.feedbackStarted).Seconds()
		if g.feedback == 1 || elapsed >= 1.5 {
			g.feedback = 0
		} else {
			g.feedback = max(2, 90-int(elapsed*60))
		}
		if g.feedback == 0 {
			if g.round.State != quiz.Playing {
				g.finish()
			} else {
				g.round.Advance()
				g.selected = -1
				g.speak()
			}
		}
	}
	g.buildButtons()
	if g.editName {
		for _, r := range ebiten.AppendInputChars(nil) {
			if !unicode.IsControl(r) && utf8.RuneCountInString(g.nameDraft) < 18 {
				g.nameDraft += string(r)
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			r := []rune(g.nameDraft)
			if len(r) > 0 {
				g.nameDraft = string(r[:len(r)-1])
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.editName = false
			g.profile.Name = cleanName(g.nameDraft)
			g.nameDraft = g.profile.Name
			g.save()
		}
	} else {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			if g.screen == "play" {
				if g.paused {
					g.activate("resume")
				} else {
					g.activate("pause")
				}
			} else {
				g.activate("menu")
			}
			g.buildButtons()
			return nil
		}
		if g.screen == "play" && !g.paused {
			for i, k := range []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5} {
				if inpututil.IsKeyJustPressed(k) {
					g.answer(i)
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				g.speak()
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyTab) && len(g.buttons) > 0 {
			g.keyboardFocus = true
			g.focus = (g.focus + 1) % len(g.buttons)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && g.focus >= 0 && g.focus < len(g.buttons) {
			b := g.buttons[g.focus]
			if !b.Disabled {
				g.activate(b.ID)
				g.buildButtons()
				return nil
			}
		}
	}
	x, y := g.cursor()
	clicked := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if ts := inpututil.AppendJustPressedTouchIDs(nil); len(ts) > 0 {
		tx, ty := ebiten.TouchPosition(ts[0])
		x, y = float64(tx)/g.renderScale, float64(ty)/g.renderScale
		clicked = true
	}
	if clicked {
		g.keyboardFocus = false
		for _, b := range g.buttons {
			if float64(x) >= b.X && float64(x) < b.X+b.W && float64(y) >= b.Y && float64(y) < b.Y+b.H {
				if g.editName && b.ID != "name" {
					g.editName = false
					g.profile.Name = cleanName(g.nameDraft)
					g.nameDraft = g.profile.Name
					g.save()
				}
				if !b.Disabled {
					g.activate(b.ID)
				}
				break
			}
		}
	}
	g.buildButtons()
	return nil
}
func (g *Game) background(s *ebiten.Image) {
	s.Fill(cream)
	g.fillCircle(s, 165, 820, 165, color.RGBA{188, 216, 151, 255}, true)
	g.fillCircle(s, 1100, 840, 180, color.RGBA{188, 216, 151, 255}, true)
	for i, p := range [][2]float32{{100, 130}, {1160, 180}} {
		if g.profile.Motion {
			p[0] = float32(math.Mod(float64(p[0])+g.animationTime()*(14+float64(i)*4)+160, 1600) - 160)
		}
		g.fillCircle(s, p[0], p[1], 24, color.RGBA{222, 240, 247, 255}, true)
		g.fillCircle(s, p[0]+26, p[1]+4, 30, color.RGBA{222, 240, 247, 255}, true)
		g.fillCircle(s, p[0]+56, p[1]+9, 21, color.RGBA{222, 240, 247, 255}, true)
	}
}
func (g *Game) drawButton(s *ebiten.Image, b Button, i int) {
	b = g.animatedButton(b)
	sp := b.Sprite
	if b.Label == "" && b.ID != "name" {
		sp = "button_round"
		if b.ID == "speak" {
			sp = "button_round_primary"
		}
	}
	g.image(s, "ui/"+sp, b.X, b.Y, b.W, b.H)
	g.focusMarker(s, b, i)
	if strings.HasPrefix(b.ID, "topic:") {
		var n int
		fmt.Sscanf(b.ID, "topic:%d", &n)
		g.image(s, "topics/"+g.topics[n].ID, b.X+8, b.Y+12, 98, 98)
		g.fit(s, b.Label, b.X+108, b.Y+31, b.W-116, 23, navy)
		g.center(s, "30 слов", b.X+108, b.Y+70, b.W-116, 20, muted)
		return
	}
	if b.ID == "name" {
		v := g.nameDraft
		if g.editName && g.tick%60 < 30 {
			v += "|"
		}
		g.label(s, "Имя: "+v, b.X+24, b.Y+17, 25, navy)
		return
	}
	if b.ID == "sound" && !g.profile.Sound && b.Label == "" {
		g.center(s, "×", b.X+b.W-27, b.Y+3, 24, 27, navy)
	}
	if b.Icon != "" {
		sz := math.Min(b.H-20, 54)
		if b.Label == "" {
			g.image(s, "ui/"+b.Icon, b.X+(b.W-sz)/2, b.Y+(b.H-sz)/2-3, sz, sz)
		} else {
			g.image(s, "ui/"+b.Icon, b.X+14, b.Y+(b.H-sz)/2-3, sz, sz)
			g.fit(s, b.Label, b.X+sz+15, b.Y+(b.H-36)/2-2, b.W-sz-26, 25, navy)
		}
	} else {
		g.fit(s, b.Label, b.X, b.Y+(b.H-40)/2-3, b.W, 29, navy)
	}
}
func (g *Game) Draw(s *ebiten.Image) {
	g.background(s)
	switch g.screen {
	case "menu":
		g.label(s, "Угадай слово", 32, 24, 54, navy)
		g.label(s, "Учим английский играя", 35, 94, 24, muted)
		g.fox(s, "idle", 490, 8, 100, 100)
		g.center(s, "Выбери тему!", 0, 128, W, 30, navy)
		g.center(s, "Старт: 10 баллов   •   Верно: +1   •   Ошибка: −2", 0, 746, W, 22, navy)
		if g.editName {
			g.label(s, "Введи имя и нажми Enter", 560, 686, 21, muted)
		} else {
			g.label(s, "Нажми на имя, чтобы изменить", 560, 686, 20, muted)
		}
	case "play":
		r := g.round
		g.fit(s, "Тема: "+r.Topic.Name, 218, 42, 330, 25, navy)
		g.center(s, fmt.Sprintf("Слово %d из 30", r.Index+1), 545, 30, 270, 24, navy)
		g.fillRect(s, 570, 72, 220, 12, color.RGBA{217, 231, 233, 255}, true)
		g.fillRect(s, 570, 72, float32(220*g.displayProgress), 12, blue, true)
		g.image(s, "ui/coin", 870, 24, 65, 65)
		g.label(s, fmt.Sprintf("Баллы: %d", r.Score), 940, 40, 31, navy)
		g.center(s, "Слушай и выбирай перевод!", 0, 120, W, 36, navy)
		g.image(s, "ui/panel", 260, 184, 760, 228)
		g.center(s, "Как это по-русски?", 290, 204, 550, 23, muted)
		dy := 0.
		if g.profile.Motion {
			dy = 12 * (1 - easeOut(g.questionTime()/.33))
		}
		g.fit(s, r.Current().EN, 285, 257+dy, 570, 68, navy)
		g.center(s, "Слушать", 867, 352, 126, 18, muted)
		pose := "idle"
		if g.feedback > 0 {
			if g.correct {
				pose = "celebrate"
			} else {
				pose = "wave"
			}
		}
		g.fox(s, pose, 18, 563, 180, 180)
		if g.feedback > 0 {
			msg := "Верно! +1 балл"
			c := color.RGBA{38, 107, 63, 255}
			if !g.correct {
				msg = "Правильно: " + r.Current().RU + "   •   −2 балла"
				c = color.RGBA{154, 61, 45, 255}
			}
			g.fit(s, msg, 190, 704, 900, 28, c)
		} else {
			g.center(s, "Выбери ответ • Клавиши 1–5 • Пробел — послушать", 190, 709, 900, 21, muted)
		}
	case "result":
		g.drawResult(s)
	case "rank":
		g.center(s, "Наши рекорды", 220, 32, 840, 46, navy)
		g.center(s, "Локальный рейтинг этого устройства • лучшие 50 попыток", 140, 96, 1000, 21, muted)
		if len(g.profile.Results) == 0 {
			g.image(s, "ui/trophy", 520, 230, 240, 240)
			g.center(s, "Здесь появятся твои результаты!", 100, 500, 1080, 30, navy)
		}
		for _, h := range []struct {
			label string
			x     float64
		}{{"Место", 80}, {"Игрок", 225}, {"Тема", 475}, {"Баллы", 670}, {"Итог", 825}, {"Дата", 1030}} {
			g.label(s, h.label, h.x, 142, 18, muted)
		}
		for i := 0; i < 7; i++ {
			n := g.rankPage*7 + i
			if n >= len(g.profile.Results) {
				break
			}
			r := g.profile.Results[n]
			y := float64(180 + i*72)
			g.fillRect(s, 64, float32(y), 1152, 62, color.RGBA{230, 242, 247, 255}, true)
			g.label(s, fmt.Sprintf("%d", n+1), 84, y+13, 25, navy)
			g.fit(s, r.Name, 133, y+14, 235, 25, navy)
			g.fit(s, r.Topic, 378, y+15, 270, 24, navy)
			g.label(s, fmt.Sprintf("%d б.", r.Score), 670, y+13, 27, navy)
			state := "Попытка"
			if r.Won {
				state = "Пройдено"
			}
			g.label(s, state, 825, y+19, 20, muted)
			g.label(s, r.Date.Local().Format("02.01.2006"), 1030, y+19, 18, muted)
		}
	}
	if g.paused && g.screen == "play" {
		g.fillRect(s, 0, 0, W, H, color.RGBA{28, 34, 56, 185}, true)
		g.image(s, "ui/panel", 360, 220, 560, 500)
		g.center(s, "Сделаем перерыв?", 380, 258, 520, 35, navy)
		g.center(s, "При выходе попытка не сохранится", 380, 306, 520, 19, muted)
	}
	for i, b := range g.buttons {
		g.drawButton(s, b, i)
	}
	if g.status != "" {
		g.fillRect(s, 0, 774, W, 26, color.RGBA{255, 225, 191, 255}, true)
		g.fit(s, g.status, 10, 777, 1260, 15, navy)
	}
	if g.audioError != "" && g.screen == "play" {
		g.center(s, g.audioError, 0, 754, W, 17, muted)
	}
}

func main() {
	g, e := newGame()
	if e != nil {
		log.Fatal(e)
	}
	debugInit(g)
	g.buildButtons()
	ebiten.SetWindowSize(1280, 800)
	ebiten.SetWindowTitle("Угадай слово • English with Fox 1.2")
	ebiten.SetWindowSizeLimits(800, 500, -1, -1)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if e := ebiten.RunGame(g); e != nil {
		log.Fatal(e)
	}
}
