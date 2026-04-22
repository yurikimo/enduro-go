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

func insetRect(rect Rect, scale float64) Rect {
	if scale <= 0 {
		return Rect{}
	}
	if scale >= 1 {
		return rect
	}

	newWidth := rect.W * scale
	newHeight := rect.H * scale

	return Rect{
		X: rect.X + (rect.W-newWidth)/2,
		Y: rect.Y + (rect.H-newHeight)/2,
		W: newWidth,
		H: newHeight,
	}
}
