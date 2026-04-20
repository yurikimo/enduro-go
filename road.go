package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	roadWidth       = 180
	lineWidth       = 4
	lineHeight      = 20
	lineGap         = 20
	roadScrollSpeed = 2
)

type Road struct {
	x           float64
	width       float64
	lineOffSetY float64
	speed       float64
}

func NewRoad() Road {
	x := float64((screenWidth - roadWidth) / 2)

	return Road{
		x:           x,
		width:       roadWidth,
		lineOffSetY: 0,
		speed:       roadScrollSpeed,
	}
}

func (r *Road) Left() float64 {
	return r.x
}

func (r *Road) Right() float64 {
	return r.x + r.width
}

func (r *Road) PlayerRightLimit(playerWidth float64) float64 {
	return r.Right() - playerWidth
}

func (r *Road) Update() {
	r.lineOffSetY += r.speed + roadScrollSpeed

	segmentSize := float64(lineHeight + lineGap)

	if r.lineOffSetY >= segmentSize {
		r.lineOffSetY -= segmentSize
	}
}

func (r *Road) SetSpeed(speed float64) {
	r.speed = speed
}

func (r *Road) Reset(road Road) {
	r.speed = roadScrollSpeed
	r.lineOffSetY = 0
}

func (r *Road) Draw(screen *ebiten.Image) {
	roadColor := color.RGBA{70, 70, 70, 255}
	lineColor := color.RGBA{240, 240, 240, 255}

	ebitenutil.DrawRect(screen, r.x, 0, r.width, float64(screenHeight), roadColor)

	centerX := r.x + r.width/2

	segmentSize := float64(lineHeight + lineGap)

	for y := -segmentSize; y < float64(screenHeight); y += segmentSize {
		drawY := y + r.lineOffSetY

		ebitenutil.DrawRect(screen, centerX-2, drawY, 4, lineHeight, lineColor)
	}
}
