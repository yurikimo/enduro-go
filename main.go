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
	enemy    Enemy
	score    int
	gameOver bool
}

func NewGame() *Game {
	road := NewRoad()
	player := NewPlayer(road)
	enemy := NewEnemy(road)

	return &Game{
		road:   road,
		player: player,
		enemy:  enemy,
	}
}

func (g *Game) Reset() {
	g.road = NewRoad()

	g.player = NewPlayer(g.road)
	g.enemy = NewEnemy(g.road)

	g.score = 0
	g.gameOver = false
}

func (g *Game) Speedup() float64 {
	return enemyBaseSpeed + float64(g.score)*0.2
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
	g.enemy.Update()

	if g.enemy.IsOffScreen() {
		g.score++

		g.enemy.SetSpeed(g.Speedup())
		g.road.SetSpeed(g.Speedup())

		g.enemy.Reset(g.road)
	}

	if g.player.IsColliding(g.enemy.Rect(g.road)) {
		g.gameOver = true
	}

	return nil
}

func (g *Game) hudText() string {
	if g.gameOver {
		return fmt.Sprintf(
			"GAME OVER\nScore: %d\nSpeed: %.1f\nPress R to restart",
			g.score,
			g.enemy.speed,
		)
	}

	return fmt.Sprintf(
		"Score: %d\nSpeed: %.1f\nMove: Left / Right",
		g.score,
		g.enemy.speed,
	)
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.road.Draw(screen)
	g.player.Draw(screen)
	g.enemy.Draw(screen, g.road)

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
