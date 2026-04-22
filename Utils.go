package main

import (
	"fmt"
	"image/color"
)

// Rect is a simple axis-aligned rectangle used for collision and layout math.
//
// The project uses float64 values so geometry can stay in the same coordinate system as drawing code.
type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

// lerp returns the value between a and b at the normalized position t.
//
// t=0 returns a, t=1 returns b, and values in between blend linearly.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// lerpColor linearly blends two colors channel by channel.
func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(lerp(float64(a.R), float64(b.R), t)),
		G: uint8(lerp(float64(a.G), float64(b.G), t)),
		B: uint8(lerp(float64(a.B), float64(b.B), t)),
		A: 255,
	}
}

// colorRGBA creates a fully opaque color from RGB values.
func colorRGBA(r, g, b uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

// insetRect shrinks a rectangle toward its center.
//
// This is mainly used to make collisions feel fairer than the full drawn sprite size.
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

// titleBestScoreText formats the title-screen best-score label.
func titleBestScoreText(bestScore int) string {
	return fmt.Sprintf("BEST SCORE  %d", bestScore)
}
