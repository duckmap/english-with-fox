//go:build e2e && js && wasm

package main

import (
	"encoding/json"
	"syscall/js"
)

var debugCallback js.Func
var skipCallback js.Func

func debugInit(g *Game) {
	debugCallback = js.FuncOf(func(this js.Value, args []js.Value) any {
		m := map[string]any{"effects": g.profile.Effects, "sfxPlaying": g.sfx != nil && g.sfx.IsPlaying(), "sfxCount": g.resultSoundCount, "renderScale": g.renderScale, "renderWidth": g.renderW, "renderHeight": g.renderH, "age": g.screenAge, "music": g.profile.Music, "motion": g.profile.Motion, "musicVolume": g.musicVolume, "musicPlaying": g.music != nil && g.music.IsPlaying(), "voicePlaying": g.player != nil && g.player.IsPlaying(), "hover": g.hover["topic:0"], "screen": g.screen, "feedback": g.feedback, "paused": g.paused, "name": g.profile.Name, "results": len(g.profile.Results), "sound": g.profile.Sound, "audioError": g.audioError}
		if g.round != nil {
			m["score"] = g.round.Score
			m["index"] = g.round.Index
			m["correct"] = g.round.Current().RU
			m["word"] = g.round.Current().EN
			m["options"] = g.round.Options
			m["state"] = g.round.State
			m["right"] = g.round.Correct
			m["wrong"] = g.round.Wrong
		}
		b, _ := json.Marshal(m)
		return string(b)
	})
	js.Global().Set("__wordGameState", debugCallback)
	skipCallback = js.FuncOf(func(this js.Value, args []js.Value) any {
		if g.feedback > 0 {
			g.feedback = 1
		}
		return nil
	})
	js.Global().Set("__wordGameSkipFeedback", skipCallback)
}
