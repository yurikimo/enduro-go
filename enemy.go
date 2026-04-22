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

// enemyLanes stores lane offsets in normalized road space.
//
// -0.5 is left, 0 is center, and 0.5 is right.
// The actual screen X position is computed later from the road width at a specific depth.
var enemyLanes = []float64{-0.5, 0.0, 0.5}

// Enemy represents one AI car on the road.
//
// This type uses a mix of gameplay-space and rendering-space values:
//   - laneOffset is stored in normalized lane space.
//   - y is stored in screen-space and moves toward or away from the player.
//   - speed is the enemy's own forward speed used to compute relative motion.
//
// Methods that mutate an enemy use a pointer receiver because they change the stored state.
type Enemy struct {
	laneOffset  float64
	y           float64
	speed       float64
	framesAlive int
	colorIndex  int
}

// enemyGeometry is a temporary render package derived from an enemy and the road.
//
// It keeps perspective math in one place so drawing and collision can reuse the same result.
type enemyGeometry struct {
	progress float64
	width    float64
	height   float64
	x        float64
	contactY float64
}

// NewEnemy creates an enemy near the horizon.
//
// It returns a value instead of a pointer because callers usually store enemies inside a slice.
// A small value type is convenient there, and individual methods can still mutate it through pointer receivers.
func NewEnemy(road Road) Enemy {
	return Enemy{
		laneOffset:  randomLaneOffset(),
		y:           road.horizonY + enemyStartYGap,
		speed:       randomEnemySpeed(),
		framesAlive: 0,
		colorIndex:  randomEnemyColorIndex(),
	}
}

// randomEnemySpeed picks a speed within the allowed enemy range.
func randomEnemySpeed() float64 {
	return enemyMinSpeed + rand.Float64()*(enemyMaxSpeed-enemyMinSpeed)
}

// randomEnemyColorIndex chooses one of the available enemy sprite colors.
func randomEnemyColorIndex() int {
	return rand.Intn(3)
}

// LaneOffset returns the enemy's normalized lane position.
func (e *Enemy) LaneOffset() float64 {
	return e.laneOffset
}

// laneBlocked reports whether a lane is already reserved by another enemy during respawn.
func laneBlocked(lane float64, blocked []float64) bool {
	for _, b := range blocked {
		if math.Abs(lane-b) < 0.001 {
			return true
		}
	}
	return false
}

// randomLaneOffset chooses any lane with equal probability.
func randomLaneOffset() float64 {
	return enemyLanes[rand.Intn(len(enemyLanes))]
}

// randomLaneOffsetAvoiding chooses a lane that is not currently blocked when possible.
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

// Update moves the enemy using relative motion.
//
// The player is the reference frame of the game:
// if the player is faster than the enemy, the enemy appears to move downward toward the player.
// if the player is slower, the enemy appears to drift upward toward the horizon.
func (e *Enemy) Update(playerSpeed float64) {
	relativeSpeed := playerSpeed - e.speed
	e.y += relativeSpeed
	e.framesAlive++
}

// spawnFromTop reinitializes an enemy near the horizon.
func (e *Enemy) spawnFromTop(road Road, blocked []float64) {
	e.laneOffset = randomLaneOffsetAvoiding(blocked)
	e.y = road.horizonY + enemyStartYGap
	e.speed = randomEnemySpeed()
	e.framesAlive = 0
	e.colorIndex = randomEnemyColorIndex()
}

// spawnFromBottom reinitializes an enemy near the player.
//
// This is used when the player is going very slowly so new traffic can still appear on screen.
func (e *Enemy) spawnFromBottom(road Road, blocked []float64) {
	e.laneOffset = randomLaneOffsetAvoiding(blocked)
	_, height := e.size(road)
	e.y = float64(screenHeight) - height - 4
	e.speed = randomEnemySpeed()
	e.framesAlive = 0
	e.colorIndex = randomEnemyColorIndex()
}

// applySpawnSpacingFromTop keeps newly spawned enemies visually separated.
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

// applySpawnSpacingFromBottom keeps bottom spawns visually separated.
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

// Respawn places the enemy back into traffic after it leaves the active play area.
func (e *Enemy) Respawn(road Road, blocked []float64, playerSpeed float64, occupiedY []float64) {
	if playerSpeed <= playerMinSpeed {
		e.spawnFromBottom(road, blocked)
		e.applySpawnSpacingFromBottom(occupiedY)
		return
	}

	e.spawnFromTop(road, blocked)
	e.applySpawnSpacingFromTop(occupiedY)
}

// Reset restores an enemy to a fresh near-horizon state.
func (e *Enemy) Reset(road Road) {
	e.laneOffset = randomLaneOffset()
	e.y = road.horizonY + enemyStartYGap
	e.speed = randomEnemySpeed()
	e.framesAlive = 0
	e.colorIndex = randomEnemyColorIndex()
}

// IsBelowScreen reports whether the enemy has moved past the bottom of the screen.
func (e *Enemy) IsBelowScreen() bool {
	return e.y > float64(screenHeight)
}

// IsAboveHorizon reports whether the enemy has moved far enough above the horizon to recycle.
func (e *Enemy) IsAboveHorizon(road Road) bool {
	return e.y < road.horizonY+enemyStartYGap-20
}

// HasExpired is a safety valve so an enemy cannot live forever due to unusual movement patterns.
func (e *Enemy) HasExpired() bool {
	return e.framesAlive >= enemyMaxLifetimeFrames
}

// perspectiveProgress returns how far the enemy is from the horizon toward the bottom of the screen.
//
// 0 means "near the horizon" and 1 means "near the camera".
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

// contactY estimates the Y position where the enemy touches the road.
//
// Using the contact point instead of the top of the sprite helps the perspective math feel more grounded.
func (e *Enemy) contactY(road Road) float64 {
	_, height := e.sizeFromProgress(e.perspectiveProgressGuess(road))
	return e.y + height
}

// perspectiveProgressGuess provides a cheaper first approximation of perspective from the enemy's top Y.
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

// sizeFromProgress converts perspective progress into a screen-space size.
//
// This is the bridge between abstract depth and the actual width/height used for drawing.
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

// size returns the current screen-space enemy size for the road perspective.
func (e *Enemy) size(road Road) (float64, float64) {
	return e.sizeFromProgress(e.perspectiveProgressGuess(road))
}

// screenX converts the enemy's normalized lane offset into a concrete screen X position.
//
// This is one of the key perspective steps in the project.
// The road is wider near the camera and narrower near the horizon, so the same laneOffset maps
// to different pixel positions depending on contactY.
func (e *Enemy) screenX(road Road, width, contactY float64) float64 {
	left, right := road.BoundsAt(contactY)
	centerX := (left + right) / 2
	roadWidthAtY := right - left

	return centerX + e.laneOffset*(roadWidthAtY*0.5) - width/2
}

// contactHeight returns the current enemy height.
func (e *Enemy) contactHeight(road Road) float64 {
	_, height := e.size(road)
	return height
}

// geometry computes the render and collision data for the enemy at the current frame.
//
// Teaching note:
// This function is where the project turns "game logic state" into "things that can be drawn".
// It combines perspective progress, scaled size, contact point, and lane-based X placement.
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

// Draw renders the enemy sprite using the geometry derived from road perspective.
//
// Important teaching note:
// The sprite itself is always authored at carSpriteWidth x carSpriteHeight.
// GeoM.Scale converts that fixed sprite-space size into the world-space size needed for this frame.
func (e *Enemy) Draw(screen *ebiten.Image, road Road, visibility float64) {
	geometry := e.geometry(road)
	tint := applyVisibility(colorRGBA(255, 255, 255), visibility, geometry.progress)

	options := &ebiten.DrawImageOptions{}
	options.GeoM.Scale(geometry.width/carSpriteWidth, geometry.height/carSpriteHeight)
	options.GeoM.Translate(geometry.x, e.y)
	options.ColorScale.ScaleWithColor(tint)

	screen.DrawImage(enemyCarSprite(e.colorIndex), options)
}

// Rect returns the enemy collision rectangle in screen space.
//
// It reuses the same perspective-derived geometry as drawing so collisions match what the player sees.
func (e *Enemy) Rect(road Road) Rect {
	geometry := e.geometry(road)

	return insetRect(Rect{
		X: geometry.x,
		Y: e.y,
		W: geometry.width,
		H: geometry.height,
	}, enemyCollisionScale)
}
