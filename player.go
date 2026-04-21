package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	playerWidth  = 16
	playerHeight = 24
	playerSpeed  = 3.5
	playerY      = screenHeight - 40
)

type Player struct {
	x float64
}

func NewPlayer(road Road) Player {
	left, right := road.BoundsAt(float64(playerY))
	startX := left + (right-left-float64(playerWidth))/2

	return Player{
		x: startX,
	}
}

func (p *Player) Update(road Road) {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.x -= playerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.x += playerSpeed
	}

	left, right := road.BoundsAt(float64(playerY))

	if p.x < left {
		p.x = left
	}
	if p.x > right-float64(playerWidth) {
		p.x = right - float64(playerWidth)
	}
}

func (p *Player) Draw(screen *ebiten.Image) {
	playerColor := color.RGBA{220, 30, 30, 255}

	ebitenutil.DrawRect(screen, p.x, float64(playerY), playerWidth, playerHeight, playerColor)
}

func (p Player) Rect() Rect {
	return Rect{
		X: p.x,
		Y: float64(playerY),
		W: playerWidth,
		H: playerHeight,
	}
}

func (p Player) IsColliding(b Rect) bool {
	playerRect := p.Rect()

	return playerRect.X < b.X+b.W &&
		playerRect.X+playerRect.W > b.X &&
		playerRect.Y < b.Y+b.H &&
		playerRect.Y+playerRect.H > b.Y
}
