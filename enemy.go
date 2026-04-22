package main

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	enemyBaseWidth         = 1.0
	enemyBaseHeight        = 1.0
	enemyMaxWidth          = 32
	enemyMaxHeight         = 24
	enemyMinSpeed          = 2.0
	enemyMaxSpeed          = 4.0
	enemyStartYGap         = 10.0
	enemySpawnGap          = 34.0
	enemyMaxSizeAt         = 3.0 / 6.0
	enemyMaxLifetimeFrames = 900
	enemyCollisionScale    = 1
)

var enemyLanes = []float64{-0.5, 0.0, 0.5}

type Enemy struct {
	laneOffset  float64
	y           float64
	speed       float64
	framesAlive int
	colorIndex  int
}

type enemyGeometry struct {
	progress float64
	width    float64
	height   float64
	x        float64
	contactY float64
}

func NewEnemy(road Road) Enemy {
	return Enemy{
		laneOffset:  randomLaneOffset(),
		y:           road.horizonY + enemyStartYGap,
		speed:       randomEnemySpeed(),
		framesAlive: 0,
		colorIndex:  randomEnemyColorIndex(),
	}
}

func randomEnemySpeed() float64 {
	return enemyMinSpeed + rand.Float64()*(enemyMaxSpeed-enemyMinSpeed)
}

func randomEnemyColorIndex() int {
	return rand.Intn(3)
}

func (e *Enemy) LaneOffset() float64 {
	return e.laneOffset
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

func (e *Enemy) Update(playerSpeed float64) {
	relativeSpeed := playerSpeed - e.speed
	e.y += relativeSpeed
	e.framesAlive++
}

func (e *Enemy) spawnFromTop(road Road, blocked []float64) {
	e.laneOffset = randomLaneOffsetAvoiding(blocked)
	e.y = road.horizonY + enemyStartYGap
	e.speed = randomEnemySpeed()
	e.framesAlive = 0
	e.colorIndex = randomEnemyColorIndex()
}

func (e *Enemy) spawnFromBottom(road Road, blocked []float64) {
	e.laneOffset = randomLaneOffsetAvoiding(blocked)
	_, height := e.size(road)
	e.y = float64(screenHeight) - height - 4
	e.speed = randomEnemySpeed()
	e.framesAlive = 0
	e.colorIndex = randomEnemyColorIndex()
}

func (e *Enemy) applySpawnSpacingFromTop(occupiedY []float64) {
	topMostY := float64(screenHeight)
	found := false

	for _, y := range occupiedY {
		if y < topMostY {
			topMostY = y
			found = true
		}
	}

	if found {
		spacedY := topMostY - enemySpawnGap
		if spacedY < e.y {
			e.y = spacedY
		}
	}

	minSpawnY := -enemyMaxHeight - enemySpawnGap
	if e.y < minSpawnY {
		e.y = minSpawnY
	}
}

func (e *Enemy) applySpawnSpacingFromBottom(occupiedY []float64) {
	bottomMostY := -enemySpawnGap
	found := false

	for _, y := range occupiedY {
		if y > bottomMostY {
			bottomMostY = y
			found = true
		}
	}

	if found {
		spacedY := bottomMostY + enemySpawnGap
		if spacedY > e.y {
			e.y = spacedY
		}
	}

	_, height := e.sizeFromProgress(1)
	maxSpawnY := float64(screenHeight) - height - 4
	if e.y > maxSpawnY {
		e.y = maxSpawnY
	}
}

func (e *Enemy) Respawn(road Road, blocked []float64, playerSpeed float64, occupiedY []float64) {
	if playerSpeed <= playerMinSpeed {
		e.spawnFromBottom(road, blocked)
		e.applySpawnSpacingFromBottom(occupiedY)
		return
	}

	e.spawnFromTop(road, blocked)
	e.applySpawnSpacingFromTop(occupiedY)
}

func (e *Enemy) Reset(road Road) {
	e.laneOffset = randomLaneOffset()
	e.y = road.horizonY + enemyStartYGap
	e.speed = randomEnemySpeed()
	e.framesAlive = 0
	e.colorIndex = randomEnemyColorIndex()
}

func (e *Enemy) IsBelowScreen() bool {
	return e.y > float64(screenHeight)
}

func (e *Enemy) IsAboveHorizon(road Road) bool {
	return e.y < road.horizonY+enemyStartYGap-20
}

func (e *Enemy) HasExpired() bool {
	return e.framesAlive >= enemyMaxLifetimeFrames
}

func (e *Enemy) perspectiveProgress(road Road) float64 {
	progress := (e.contactY(road) - road.horizonY) / (float64(screenHeight) - road.horizonY)

	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	return progress
}

func (e *Enemy) contactY(road Road) float64 {
	_, height := e.sizeFromProgress(e.perspectiveProgressGuess(road))
	return e.y + height
}

func (e *Enemy) perspectiveProgressGuess(road Road) float64 {
	progress := (e.y - road.horizonY) / (float64(screenHeight) - road.horizonY)

	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}

	return progress
}

func (e *Enemy) sizeFromProgress(progress float64) (float64, float64) {
	if progress < 0 {
		progress = 0
	}

	sizeProgress := progress / enemyMaxSizeAt
	if sizeProgress > 1 {
		sizeProgress = 1
	}

	width := enemyBaseWidth + (enemyMaxWidth-enemyBaseWidth)*sizeProgress
	height := enemyBaseHeight + (enemyMaxHeight-enemyBaseHeight)*sizeProgress

	return width, height
}

func (e *Enemy) size(road Road) (float64, float64) {
	return e.sizeFromProgress(e.perspectiveProgressGuess(road))
}

func (e *Enemy) screenX(road Road, width, contactY float64) float64 {
	left, right := road.BoundsAt(contactY)
	centerX := (left + right) / 2
	roadWidthAtY := right - left

	return centerX + e.laneOffset*(roadWidthAtY*0.5) - width/2
}

func (e *Enemy) contactHeight(road Road) float64 {
	_, height := e.size(road)
	return height
}

func (e *Enemy) geometry(road Road) enemyGeometry {
	progress := e.perspectiveProgressGuess(road)
	width, height := e.sizeFromProgress(progress)
	contactY := e.y + height

	return enemyGeometry{
		progress: progress,
		width:    width,
		height:   height,
		x:        e.screenX(road, width, contactY),
		contactY: contactY,
	}
}

func (e *Enemy) Draw(screen *ebiten.Image, road Road, visibility float64) {
	geometry := e.geometry(road)
	tint := applyVisibility(colorRGBA(255, 255, 255), visibility, geometry.progress)

	options := &ebiten.DrawImageOptions{}
	options.GeoM.Scale(geometry.width/carSpriteWidth, geometry.height/carSpriteHeight)
	options.GeoM.Translate(geometry.x, e.y)
	options.ColorScale.ScaleWithColor(tint)

	screen.DrawImage(enemyCarSprite(e.colorIndex), options)
}

func (e *Enemy) Rect(road Road) Rect {
	geometry := e.geometry(road)

	return insetRect(Rect{
		X: geometry.x,
		Y: e.y,
		W: geometry.width,
		H: geometry.height,
	}, enemyCollisionScale)
}
