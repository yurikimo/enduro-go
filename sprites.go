package main

import (
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	carSpriteWidth  = 24
	carSpriteHeight = 16
)

var (
	spriteOnce   sync.Once
	playerSprite *ebiten.Image
	enemySprites []*ebiten.Image
)

func initSprites() {
	spriteOnce.Do(func() {
		playerSprite = makeCarSprite(
			color.RGBA{220, 30, 30, 255},
			color.RGBA{255, 232, 92, 255},
			color.RGBA{190, 28, 28, 255},
		)
		enemySprites = []*ebiten.Image{
			makeCarSprite(
				color.RGBA{30, 144, 255, 255},
				color.RGBA{198, 238, 255, 255},
				color.RGBA{24, 110, 210, 255},
			),
			makeCarSprite(
				color.RGBA{44, 170, 82, 255},
				color.RGBA{210, 246, 214, 255},
				color.RGBA{34, 132, 64, 255},
			),
			makeCarSprite(
				color.RGBA{232, 210, 72, 255},
				color.RGBA{255, 248, 180, 255},
				color.RGBA{200, 168, 46, 255},
			),
		}
	})
}

func playerCarSprite() *ebiten.Image {
	initSprites()
	return playerSprite
}

func enemyCarSprite(index int) *ebiten.Image {
	initSprites()
	if len(enemySprites) == 0 {
		return nil
	}
	if index < 0 || index >= len(enemySprites) {
		index = 0
	}
	return enemySprites[index]
}

func makeCarSprite(bodyColor, glassColor, accentColor color.RGBA) *ebiten.Image {
	sprite := image.NewRGBA(image.Rect(0, 0, carSpriteWidth, carSpriteHeight))

	darkShade := accentColor
	brightAccent := lerpColor(bodyColor, colorRGBA(255, 255, 255), 0.25)
	windowDark := lerpColor(glassColor, colorRGBA(40, 40, 40), 0.65)
	windowMid := glassColor
	windowLight := lerpColor(glassColor, colorRGBA(255, 255, 255), 0.35)
	tailRed := color.RGBA{220, 0, 0, 255}
	tailOrange := color.RGBA{255, 120, 0, 255}
	wheel := color.RGBA{24, 24, 24, 255}

	// Clear background is implicit.

	// --- Roof / rear window frame ---
	paintRect(sprite, 6, 0, 12, 1, bodyColor)
	paintRect(sprite, 5, 1, 1, 1, bodyColor)
	paintRect(sprite, 18, 1, 1, 1, bodyColor)

	paintRect(sprite, 4, 2, 2, 1, bodyColor)
	paintRect(sprite, 18, 2, 2, 1, bodyColor)

	paintRect(sprite, 3, 3, 2, 1, bodyColor)
	paintRect(sprite, 19, 3, 2, 1, bodyColor)

	paintRect(sprite, 2, 4, 2, 1, bodyColor)
	paintRect(sprite, 20, 4, 2, 1, bodyColor)

	// --- Rear window ---
	paintRect(sprite, 6, 2, 12, 1, windowDark)
	paintRect(sprite, 5, 3, 14, 1, windowDark)
	paintRect(sprite, 4, 4, 16, 1, windowMid)
	paintRect(sprite, 5, 5, 10, 1, windowLight)
	paintRect(sprite, 15, 5, 3, 1, windowDark)

	// --- Upper body / shoulders ---
	paintRect(sprite, 2, 5, 2, 1, bodyColor)
	paintRect(sprite, 20, 5, 2, 1, bodyColor)

	paintRect(sprite, 1, 6, 22, 1, bodyColor)
	paintRect(sprite, 1, 7, 22, 1, bodyColor)

	// Dark under-roof band
	paintRect(sprite, 2, 8, 20, 1, darkShade)

	// Small left details visible in the reference
	paintRect(sprite, 5, 8, 1, 1, darkShade)
	paintRect(sprite, 7, 8, 1, 1, brightAccent)
	paintRect(sprite, 5, 9, 1, 1, brightAccent)
	paintRect(sprite, 7, 9, 1, 1, brightAccent)

	// --- Main rear body ---
	paintRect(sprite, 1, 9, 22, 1, bodyColor)
	paintRect(sprite, 0, 10, 24, 1, darkShade)
	paintRect(sprite, 0, 11, 24, 3, darkShade)

	// Center blue panel / rear grille
	paintRect(sprite, 9, 10, 6, 1, brightAccent)
	paintRect(sprite, 9, 11, 6, 1, brightAccent)
	paintRect(sprite, 8, 12, 8, 1, brightAccent)

	// Lower center recess
	paintRect(sprite, 8, 13, 8, 1, wheel)

	// --- Tail lights ---
	paintRect(sprite, 3, 11, 4, 2, tailRed)
	paintRect(sprite, 17, 11, 4, 2, tailRed)

	paintRect(sprite, 4, 11, 1, 1, tailOrange)
	paintRect(sprite, 6, 11, 1, 1, tailOrange)
	paintRect(sprite, 18, 11, 1, 1, tailOrange)
	paintRect(sprite, 20, 11, 1, 1, tailOrange)

	// Inner vertical light bars
	paintRect(sprite, 7, 11, 1, 2, tailRed)
	paintRect(sprite, 16, 11, 1, 2, tailRed)

	// --- Lower body corners ---
	paintRect(sprite, 0, 14, 8, 1, darkShade)
	paintRect(sprite, 16, 14, 8, 1, darkShade)

	// --- Wheels ---
	paintRect(sprite, 2, 14, 4, 2, wheel)
	paintRect(sprite, 18, 14, 4, 2, wheel)

	return ebiten.NewImageFromImage(sprite)
}

func paintRect(img *image.RGBA, x, y, width, height int, fill color.Color) {
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			img.Set(px, py, fill)
		}
	}
}
