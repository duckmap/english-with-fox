package view

import "testing"

func TestPhysicalResolution(t *testing.T) {
	for _, c := range []struct {
		w, h   int
		d, s   float64
		rw, rh int
	}{{1280, 800, 1, 1, 1280, 800}, {1920, 1080, 1, 1.35, 1728, 1080}, {3840, 2160, 1, 2.7, 3456, 2160}, {1920, 1080, 2, 2.7, 3456, 2160}, {1280, 800, 2, 2, 2560, 1600}, {7680, 4320, 1, 3, 3840, 2400}} {
		s, w, h := Fit(c.w, c.h, c.d)
		if s != c.s || w != c.rw || h != c.rh {
			t.Fatalf("%+v: %v %d %d", c, s, w, h)
		}
		x := 150.
		if x*s/s != x {
			t.Fatal("input mapping")
		}
	}
}
