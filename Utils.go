package main

import "image/color"

type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}
func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(lerp(float64(a.R), float64(b.R), t)),
		G: uint8(lerp(float64(a.G), float64(b.G), t)),
		B: uint8(lerp(float64(a.B), float64(b.B), t)),
		A: 255,
	}
}

func colorRGBA(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 255}
}
