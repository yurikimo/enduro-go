package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	playerWidth        = 16
	playerHeight       = 24
	playerY            = screenHeight - 40
	playerSteerSpeed   = 3.5
	playerMinSpeed     = 0.5
	playerMaxSpeed     = 6.0
	playerStartSpeed   = 3.0
	playerAcceleration = 0.06
	playerBrakeSpeed   = 0.10
	playerCurveDrift   = 0.015
)

type Player struct {
	x     float64
	speed float64
}

func NewPlayer(road Road) Player {
	left, right := road.BoundsAt(float64(playerY))
	startX := left + (right-left-float64(playerWidth))/2

	return Player{
		x:     startX,
		speed: playerStartSpeed,
	}
}

func (p *Player) Update(road Road) {
	if ebiten.IsKeyPressed(ebiten.KeyLeft) {
		p.x -= playerSteerSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) {
		p.x += playerSteerSpeed
	}

	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		p.speed += playerAcceleration
	} else if ebiten.IsKeyPressed(ebiten.KeyDown) {
		p.speed -= playerBrakeSpeed
	}

	if p.speed < playerMinSpeed {
		p.speed = playerMinSpeed
	}
	if p.speed > playerMaxSpeed {
		p.speed = playerMaxSpeed
	}

	// Road curvature pushes the player sideways.
	p.x -= road.CurveOffset() * playerCurveDrift

	left, right := road.BoundsAt(float64(playerY))
	if p.x < left {
		p.x = left
	}
	if p.x > right-float64(playerWidth) {
		p.x = right - float64(playerWidth)
	}
}

func (p Player) Speed() float64 {
	return p.speed
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
