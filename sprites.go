package main

import (
	"image"
	"image/color"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	carSpriteWidth  = 16
	carSpriteHeight = 24
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

	paintRect(sprite, 5, 0, 6, 2, glassColor)
	paintRect(sprite, 4, 2, 8, 4, bodyColor)
	paintRect(sprite, 3, 6, 10, 14, bodyColor)
	paintRect(sprite, 4, 8, 8, 4, glassColor)
	paintRect(sprite, 4, 14, 8, 4, accentColor)
	paintRect(sprite, 5, 18, 6, 3, bodyColor)
	paintRect(sprite, 4, 21, 8, 2, accentColor)

	paintRect(sprite, 2, 4, 2, 5, color.Black)
	paintRect(sprite, 12, 4, 2, 5, color.Black)
	paintRect(sprite, 2, 14, 2, 5, color.Black)
	paintRect(sprite, 12, 14, 2, 5, color.Black)
	paintRect(sprite, 5, 22, 2, 1, color.RGBA{255, 228, 126, 255})
	paintRect(sprite, 9, 22, 2, 1, color.RGBA{255, 228, 126, 255})

	return ebiten.NewImageFromImage(sprite)
}

func paintRect(img *image.RGBA, x, y, width, height int, fill color.Color) {
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			img.Set(px, py, fill)
		}
	}
}
