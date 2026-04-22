package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 320
	screenHeight = 240
	windowScale  = 2
)

type Game struct {
	road         Road
	player       Player
	enemies      []Enemy
	scoreManager ScoreManager
	gameOver     bool
	timeOfDay    float64
	started      bool
}

func NewGame() *Game {
	road := NewRoad()
	player := NewPlayer(road)
	scoreManager := NewScoreManager()
	enemies := []Enemy{
		NewEnemy(road),
		NewEnemy(road),
		NewEnemy(road),
	}

	for i := 1; i < len(enemies); i++ {
		enemies[i].Reset(road)
		enemies[i].y = road.horizonY + float64(45*i)
	}

	scoreManager.LoadBestScore()

	return &Game{
		road:         road,
		player:       player,
		enemies:      enemies,
		scoreManager: scoreManager,
		started:      false,
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
		g.enemies[i].Reset(g.road)
		g.enemies[i].y = g.road.horizonY + float64(45*i)
	}

	g.scoreManager.ResetScore()
	g.gameOver = false
	g.timeOfDay = 0
}

func (g *Game) Speedup() float64 {
	return enemyBaseSpeed + float64(g.scoreManager.Score())*0.03
}

func (g *Game) Update() error {
	if !g.started {
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.started = true
		}
		return nil
	}

	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.Reset()
		}

		return nil
	}

	g.road.Update()
	g.player.Update(g.road)

	g.timeOfDay += 1.0 / 60.0

	for i := range g.enemies {
		g.enemies[i].Update()

		if g.enemies[i].IsOffScreen() {
			g.scoreManager.UpdateScore()
			g.enemies[i].SetSpeed(g.Speedup())
			g.road.SetSpeed(g.Speedup())

			blockedLanes := make([]float64, 0, len(g.enemies)-1)

			for j := range g.enemies {
				if j == i {
					continue
				}
				blockedLanes = append(blockedLanes, g.enemies[j].LaneOffset())
			}

			g.enemies[i].ResetAvoidingLanes(g.road, blockedLanes)
		}

		if g.player.IsColliding(g.enemies[i].Rect(g.road)) {
			g.gameOver = true
		}
	}

	return nil
}

func (g *Game) hudText() string {
	if !g.started {
		return fmt.Sprintf("ENDURO GO\n\nBest: %d\nPress SPACE to start\nArrow keys to move", g.scoreManager.BestScore())
	}

	if g.gameOver {
		return fmt.Sprintf(
			"GAME OVER\nScore: %d\nBest: %d\nPress R to restart",
			g.scoreManager.Score(),
			g.scoreManager.BestScore(),
		)
	}

	return fmt.Sprintf(
		"Score: %d\nBest: %d\nMove: Left / Right",
		g.scoreManager.Score(),
		g.scoreManager.BestScore(),
	)
}
func (g *Game) sceneLight() float64 {
	dayNightCycleSeconds := 40.0

	phase := math.Mod(g.timeOfDay/dayNightCycleSeconds, 2.0)
	if phase > 1 {
		phase = 2 - phase
	}

	return 1.0 - phase
}
func (g *Game) skyColor() color.RGBA {
	dayNightCycleSeconds := 40.0
	phase := math.Mod(g.timeOfDay/dayNightCycleSeconds, 2.0)

	day := color.RGBA{40, 80, 220, 255}
	dusk := color.RGBA{220, 120, 60, 255}
	night := color.RGBA{10, 10, 40, 255}

	if phase < 0.5 {
		t := phase / 0.5
		return lerpColor(day, dusk, t)
	}

	if phase < 1.0 {
		t := (phase - 0.5) / 0.5
		return lerpColor(dusk, night, t)
	}

	if phase < 1.5 {
		t := (phase - 1.0) / 0.5
		return lerpColor(night, dusk, t)
	}

	t := (phase - 1.5) / 0.5
	return lerpColor(dusk, day, t)
}
func (g *Game) visibility() float64 {
	// At full day: 1.0
	// At full night: 0.45
	return 0.20 + g.sceneLight()*0.80
}
func (g *Game) Draw(screen *ebiten.Image) {
	visibility := g.visibility()

	g.road.Draw(screen, g.skyColor(), g.sceneLight(), visibility)

	g.player.Draw(screen)

	for _, enemy := range g.enemies {
		enemy.Draw(screen, g.road, visibility)
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
