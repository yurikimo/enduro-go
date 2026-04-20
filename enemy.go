package main

import (
	"image/color"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	enemyWidth     = 16
	enemyHeight    = 24
	enemyBaseSpeed = 2.0
)

type Enemy struct {
	x     float64
	y     float64
	speed float64
}

func NewEnemy(road Road) Enemy {
	startX := randomEnemyX(road)

	return Enemy{
		x:     startX,
		y:     -enemyHeight,
		speed: enemyBaseSpeed,
	}
}

func randomEnemyX(road Road) float64 {
	maxX := road.Right() - enemyWidth
	minX := road.Left()

	return minX + rand.Float64()*(maxX-minX)
}

func (e *Enemy) Update() {
	e.y += e.speed
}

func (e *Enemy) Reset(road Road) {
	e.x = randomEnemyX(road)
	e.y = -enemyHeight
}

func (e *Enemy) SetSpeed(speed float64) {
	e.speed = speed
}

func (e Enemy) IsOffScreen() bool {
	return e.y > screenHeight
}

func (e Enemy) Draw(screen *ebiten.Image) {
	enemyColor := color.RGBA{30, 144, 255, 255}

	ebitenutil.DrawRect(screen, e.x, e.y, enemyWidth, enemyHeight, enemyColor)
}

func (e Enemy) Rect() Rect {
	return Rect{
		X: e.x,
		Y: e.y,
		W: enemyWidth,
		H: enemyHeight,
	}
}
