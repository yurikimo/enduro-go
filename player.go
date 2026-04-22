package main

import "github.com/hajimehoshi/ebiten/v2"

const (
	playerWidth          = 32
	playerHeight         = 24
	playerY              = screenHeight - 40
	playerSteerSpeed     = 3.5
	playerMinSpeed       = 0.5
	playerMaxSpeed       = 6.0
	playerStartSpeed     = 3.0
	playerAcceleration   = 0.06
	playerBrakeSpeed     = 0.10
	playerCurveDrift     = 0.015
	playerCollisionScale = 0.75
)

// Player stores the controllable car state.
//
// x is the left edge of the player in screen space.
// speed is the forward speed used by the road and enemy relative-motion systems.
//
// Methods that modify the player use pointer receivers because they update this state each frame.
type Player struct {
	x     float64
	speed float64
}

// NewPlayer creates a player centered at the bottom portion of the road.
//
// It returns a value because Player is small and easy to copy when constructing the Game.
func NewPlayer(road Road) Player {
	left, right := road.BoundsAt(float64(playerY))
	startX := left + (right-left-float64(playerWidth))/2

	return Player{
		x:     startX,
		speed: playerStartSpeed,
	}
}

// Update handles steering, acceleration, braking, and road-bound clamping.
//
// The player is kept in screen space because it is the closest object to the camera and does not need
// the same perspective calculations used by enemies.
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

// Speed returns the current forward speed.
func (p Player) Speed() float64 {
	return p.speed
}

// Draw renders the player sprite near the bottom of the road.
//
// As with enemies, the sprite is authored in sprite space and then scaled to gameplay size.
// Here the target world-space size is fixed, so the scale factors are constant.
func (p *Player) Draw(screen *ebiten.Image, road Road) {
	sprite := playerCarSprite()

	options := &ebiten.DrawImageOptions{}
	options.GeoM.Scale(
		float64(playerWidth)/carSpriteWidth,
		float64(playerHeight)/carSpriteHeight,
	)
	options.GeoM.Translate(p.x, float64(playerY))

	screen.DrawImage(sprite, options)
}

// Rect returns the player's collision rectangle.
func (p Player) Rect() Rect {
	return insetRect(Rect{
		X: p.x,
		Y: float64(playerY),
		W: playerWidth,
		H: playerHeight,
	}, playerCollisionScale)
}

// IsColliding performs a simple axis-aligned rectangle overlap test.
func (p Player) IsColliding(b Rect) bool {
	playerRect := p.Rect()

	return playerRect.X < b.X+b.W &&
		playerRect.X+playerRect.W > b.X &&
		playerRect.Y < b.Y+b.H &&
		playerRect.Y+playerRect.H > b.Y
}
