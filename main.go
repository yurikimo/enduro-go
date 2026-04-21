package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 320
	screenHeight = 240
	windowScale  = 2
)

type Game struct {
	road     Road
	player   Player
	enemies  []Enemy
	score    int
	gameOver bool
}

func NewGame() *Game {
	road := NewRoad()
	player := NewPlayer(road)
	enemies := []Enemy{
		NewEnemy(road),
		NewEnemy(road),
	}

	for i := 1; i < len(enemies); i++ {
		enemies[1].Reset(road)
		enemies[1].y = road.horizonY + float64(45*i)
	}

	return &Game{
		road:    road,
		player:  player,
		enemies: enemies,
	}
}

func (g *Game) Reset() {
	g.road = NewRoad()

	g.player = NewPlayer(g.road)

	g.enemies = []Enemy{
		NewEnemy(g.road),
		NewEnemy(g.road),
		NewEnemy(g.road),
	}

	for i := 1; i < len(g.enemies); i++ {
		g.enemies[1].Reset(g.road)
		g.enemies[1].y = g.road.horizonY + float64(45*i)
	}

	g.score = 0
	g.gameOver = false
}

func (g *Game) Speedup() float64 {
	return enemyBaseSpeed + float64(g.score)*0.06
}

func (g *Game) Update() error {
	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.Reset()
		}

		return nil
	}

	g.road.Update()
	g.player.Update(g.road)

	for i := range g.enemies {
		g.enemies[i].Update()

		if g.enemies[i].IsOffScreen() {
			g.score++
			g.enemies[i].SetSpeed(g.Speedup())
			g.road.SetSpeed(g.Speedup())

			g.enemies[i].Reset(g.road)
		}

		if g.player.IsColliding(g.enemies[i].Rect(g.road)) {
			g.gameOver = true
		}
	}

	return nil
}

func (g *Game) hudText() string {
	if g.gameOver {
		return fmt.Sprintf(
			"GAME OVER\nScore: %d\nPress R to restart",
			g.score,
		)
	}

	return fmt.Sprintf(
		"Score: %d\nMove: Left / Right",
		g.score,
	)
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.road.Draw(screen)
	g.player.Draw(screen)

	for _, enemy := range g.enemies {
		enemy.Draw(screen, g.road)
	}

	ebitenutil.DebugPrintAt(screen, g.hudText(), 8, 8)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth*windowScale, screenHeight*windowScale)
	ebiten.SetWindowTitle("Enduro GO - Milestone 1")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
