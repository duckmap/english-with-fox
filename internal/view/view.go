package view

import "math"

// Fit renders at physical display density, capped above 4K for a bounded GPU budget.
func Fit(outW, outH int, dpr float64) (float64, int, int) {
	if outW < 1 {
		outW = 1280
	}
	if outH < 1 {
		outH = 800
	}
	if dpr <= 0 {
		dpr = 1
	}
	s := math.Min(float64(outW)/1280, float64(outH)/800) * dpr
	s = math.Max(.25, math.Min(3, s))
	return s, int(math.Round(1280 * s)), int(math.Round(800 * s))
}
