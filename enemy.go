package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	enemyBaseWidth  = 8.0
	enemyBaseHeight = 12.0
	enemyMaxWidth   = 16.0
	enemyMaxHeight  = 24.0
	enemyBaseSpeed  = 2.0
	enemyStartYGap  = 10.0
)

var enemyLanes = []float64{-0.75, 0.0, 0.75}

type Enemy struct {
	laneOffset float64
	y          float64
	speed      float64
}

func NewEnemy(road Road) Enemy {
	return Enemy{
		laneOffset: randomLaneOffset(),
		y:          road.horizonY + enemyStartYGap,
		speed:      enemyBaseSpeed,
	}
}

func (e *Enemy) LaneOffset() float64 {
	return e.laneOffset
}
func (e *Enemy) ResetAvoidingLanes(road Road, blocked []float64) {
	e.laneOffset = randomLaneOffsetAvoiding(blocked)
	e.y = road.horizonY + enemyStartYGap
}

func laneBlocked(lane float64, blocked []float64) bool {
	for _, b := range blocked {
		if math.Abs(lane-b) < 0.001 {
			return true
		}
	}
	return false
}

func randomLaneOffset() float64 {
	return enemyLanes[rand.Intn(len(enemyLanes))]
}

func randomLaneOffsetAvoiding(blocked []float64) float64 {
	choices := make([]float64, 0, len(enemyLanes))

	for _, lane := range enemyLanes {
		if !laneBlocked(lane, blocked) {
			choices = append(choices, lane)
		}
	}

	if len(choices) == 0 {
		return randomLaneOffset()
	}

	return choices[rand.Intn(len(choices))]
}

func (e *Enemy) Update() {
	e.y += e.speed
}

func (e *Enemy) Reset(road Road) {
	e.laneOffset = randomLaneOffset()
	e.y = road.horizonY + enemyStartYGap
}

func (e *Enemy) SetSpeed(speed float64) {
	e.speed = speed
}

func (e *Enemy) IsOffScreen() bool {
	return e.y > float64(screenHeight)
}

func (e *Enemy) perspectiveProgress(road Road) float64 {
	progress := (e.y - road.horizonY) / (float64(screenHeight) - road.horizonY)

	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	return progress
}

func (e *Enemy) size(road Road) (float64, float64) {
	progress := e.perspectiveProgress(road)

	width := enemyBaseWidth + (enemyMaxWidth-enemyBaseWidth)*progress
	height := enemyBaseHeight + (enemyMaxHeight-enemyBaseHeight)*progress

	return width, height
}

func (e *Enemy) screenX(road Road) float64 {
	left, right := road.BoundsAt(e.y)
	centerX := (left + right) / 2
	roadWidthAtY := right - left

	width, _ := e.size(road)

	return centerX + e.laneOffset*(roadWidthAtY*0.5) - width/2
}

func (e *Enemy) Draw(screen *ebiten.Image, road Road) {
	enemyColor := color.RGBA{30, 144, 255, 255}

	width, height := e.size(road)
	x := e.screenX(road)

	ebitenutil.DrawRect(screen, x, e.y, width, height, enemyColor)
}

func (e *Enemy) Rect(road Road) Rect {
	width, height := e.size(road)
	x := e.screenX(road)

	return Rect{
		X: x,
		Y: e.y,
		W: width,
		H: height,
	}
}
