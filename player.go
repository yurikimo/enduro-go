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
	startX := road.Left() + (road.width-float64(playerWidth))/2

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

	if p.x < road.Left() {
		p.x = road.Left()
	}
	if p.x > road.PlayerRightLimit(float64(playerWidth)) {
		p.x = road.PlayerRightLimit(float64(playerWidth))
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
	return p.Rect().X < b.X+b.W &&
		p.Rect().X+p.Rect().W > b.X &&
		p.Rect().Y < b.Y+b.H &&
		p.Rect().Y+p.Rect().H > b.Y
}
