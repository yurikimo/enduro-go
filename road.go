package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	roadBottomWidth = (screenWidth / 5) * 4
	roadTopWidth    = 2.0
	roadHorizonY    = 60.0
	edgeLineWidth   = 2.0
	roadScrollSpeed = 2.0
)

type Road struct {
	x           float64
	bottomWidth float64
	topWidth    float64
	horizonY    float64
	lineOffsetY float64
	speed       float64
}

func NewRoad() Road {
	centerX := float64(screenWidth) / 2

	return Road{
		x:           centerX,
		bottomWidth: roadBottomWidth,
		topWidth:    roadTopWidth,
		horizonY:    roadHorizonY,
		lineOffsetY: 0,
		speed:       roadScrollSpeed,
	}
}

func (r *Road) Left() float64 {
	return r.x - r.bottomWidth/2
}

func (r *Road) Right() float64 {
	return r.x + r.bottomWidth/2
}

func (r *Road) PlayerRightLimit(playerWidth float64) float64 {
	return r.Right() - playerWidth
}

func (r *Road) SetSpeed(speed float64) {
	r.speed = speed
}

func (r *Road) CenterXAt(y float64) float64 {
	left, right := r.roadEdgesAt(y)
	return (left + right) / 2
}

func (r *Road) BoundsAt(y float64) (float64, float64) {
	return r.roadEdgesAt(y)
}

func (r *Road) Update() {
	r.lineOffsetY += 0.04 * r.speed
	if r.lineOffsetY >= 1.0 {
		r.lineOffsetY -= 1.0
	}
}

func (r *Road) roadEdgesAt(y float64) (float64, float64) {
	if y < r.horizonY {
		y = r.horizonY
	}
	if y > float64(screenHeight) {
		y = float64(screenHeight)
	}

	progress := (y - r.horizonY) / (float64(screenHeight) - r.horizonY)
	width := r.topWidth + (r.bottomWidth-r.topWidth)*progress

	left := r.x - width/2
	right := r.x + width/2

	return left, right
}

func scaleColor(base color.RGBA, light float64) color.RGBA {
	if light < 0 {
		light = 0
	}
	if light > 1 {
		light = 1
	}

	minBrightness := 0.35
	factor := minBrightness + light*(1.0-minBrightness)

	return color.RGBA{
		R: uint8(float64(base.R) * factor),
		G: uint8(float64(base.G) * factor),
		B: uint8(float64(base.B) * factor),
		A: base.A,
	}
}
func horizonColor(sceneLight float64) color.RGBA {
	base := color.RGBA{200, 190, 90, 255}
	return scaleColor(base, sceneLight)
}
func drawHill(screen *ebiten.Image, x, y, w, h float64, hillColor color.RGBA) {
	for i := 0.0; i < h; i++ {
		progress := i / h

		// Narrow near the top, wide near the bottom.
		currentWidth := w * (0.3 + 0.7*progress)
		left := x + (w-currentWidth)/2

		ebitenutil.DrawRect(screen, left, y+i, currentWidth, 1, hillColor)
	}
}
func applyVisibility(base color.RGBA, visibility float64, distanceFactor float64) color.RGBA {
	if visibility < 0 {
		visibility = 0
	}
	if visibility > 1 {
		visibility = 1
	}
	if distanceFactor < 0 {
		distanceFactor = 0
	}
	if distanceFactor > 1 {
		distanceFactor = 1
	}

	// Farther objects lose more brightness at night.
	effective := visibility + (1.0-visibility)*(distanceFactor*0.5)

	return color.RGBA{
		R: uint8(float64(base.R) * effective),
		G: uint8(float64(base.G) * effective),
		B: uint8(float64(base.B) * effective),
		A: base.A,
	}
}
func (r *Road) Draw(screen *ebiten.Image, skyColor color.RGBA, sceneLight float64, visibility float64) {
	groundColor := scaleColor(color.RGBA{34, 139, 34, 255}, sceneLight)
	roadColor := scaleColor(color.RGBA{70, 70, 70, 255}, sceneLight)
	lineColor := scaleColor(color.RGBA{240, 240, 240, 255}, sceneLight)

	screen.Fill(skyColor)
	ebitenutil.DrawRect(screen, 0, r.horizonY, float64(screenWidth), float64(screenHeight)-r.horizonY, groundColor)

	hillColor := horizonColor(sceneLight)

	drawHill(screen, 45, r.horizonY, 55, 12, hillColor)
	drawHill(screen, 220, r.horizonY, 50, 10, hillColor)

	for y := int(r.horizonY); y < screenHeight; y++ {
		left, right := r.roadEdgesAt(float64(y))
		ebitenutil.DrawRect(screen, left, float64(y), right-left, 1, roadColor)
	}

	for y := int(r.horizonY); y < screenHeight; y++ {
		left, right := r.roadEdgesAt(float64(y))
		ebitenutil.DrawRect(screen, left, float64(y), edgeLineWidth, 1, lineColor)
		ebitenutil.DrawRect(screen, right-edgeLineWidth, float64(y), edgeLineWidth, 1, lineColor)
	}

	markerCount := 8.0
	for i := 0.0; i < markerCount; i++ {
		progress := (i + r.lineOffsetY) / markerCount
		if progress > 1 {
			progress -= 1
		}

		yProgress := progress * progress
		y := r.horizonY + yProgress*(float64(screenHeight)-r.horizonY)

		left, right := r.roadEdgesAt(y)
		centerX := (left + right) / 2

		lineWidth := 1.0 + progress*3.0
		lineHeight := 3.0 + progress*16.0

		distanceFactor := progress
		markerColor := applyVisibility(lineColor, visibility, distanceFactor)
		ebitenutil.DrawRect(screen, centerX-lineWidth/2, y, lineWidth, lineHeight, markerColor)
	}
}
