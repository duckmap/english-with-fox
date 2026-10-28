package main

import (
	"bytes"
	"io"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// Music is a local, original 20-second loop; no network or synthesizer at runtime.
func (g *Game) initMusic() error {
	b, err := assets.ReadFile("assets/music/fox_garden.wav")
	if err != nil {
		return err
	}
	stream, err := wav.DecodeWithSampleRate(44100, bytes.NewReader(b))
	if err != nil {
		return err
	}
	pcm, err := io.ReadAll(stream)
	if err != nil {
		return err
	}
	loop := audio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
	g.music, err = g.audio.NewPlayer(loop)
	if err != nil {
		return err
	}
	g.music.SetVolume(0)
	return nil
}
func (g *Game) updateMusic() {
	if g.music == nil {
		return
	}
	now := time.Now()
	dt := 1. / 60
	if !g.lastMusicUpdate.IsZero() {
		dt = now.Sub(g.lastMusicUpdate).Seconds()
	}
	g.lastMusicUpdate = now
	target := .20
	if !g.profile.Music {
		target = 0
	} else if (g.player != nil && g.player.IsPlaying()) || (g.sfx != nil && g.sfx.IsPlaying()) {
		target = .035
	} else if g.paused {
		target = .08
	}
	speed := 1 - math.Exp(-dt*2.8)
	if target < g.musicVolume {
		speed = 1 - math.Exp(-dt*12)
	}
	g.musicVolume += (target - g.musicVolume) * speed
	if g.musicVolume < .001 && target == 0 {
		g.musicVolume = 0
		g.music.Pause()
	} else if !g.music.IsPlaying() {
		g.music.Play()
	}
	g.music.SetVolume(g.musicVolume)
}
func (g *Game) duckMusic() {
	if g.music != nil && g.musicVolume > .035 {
		g.musicVolume = .035
		g.music.SetVolume(g.musicVolume)
	}
}
