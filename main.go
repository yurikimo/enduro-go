package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	screenWidth  = 320
	screenHeight = 240
	windowScale  = 2
)

// Game owns the full runtime state of the application.
//
// Ebiten repeatedly calls Game.Update and Game.Draw.
// That makes this type the central coordinator for the project.
//
// Teaching note:
// Keeping game state in one struct makes the game loop easier to follow for learners.
type Game struct {
	road         Road
	player       Player
	enemies      []Enemy
	scoreManager ScoreManager
	soundManager SoundManager
	blockedLanes []float64
	occupiedY    []float64
	paused       bool
	gameOver     bool
	timeOfDay    float64
	started      bool
	pKeyDown     bool
	hudCacheKey  hudCacheKey
	hudCacheText string
}

// hudCacheKey captures the fields that affect the generated HUD text.
//
// The project caches the formatted string so it does not rebuild it every frame unnecessarily.
type hudCacheKey struct {
	started    bool
	paused     bool
	gameOver   bool
	score      int
	bestScore  int
	speedKPH   int
	newBest    bool
}

// NewGame creates the initial game state and loads persistent data.
//
// A pointer is returned because Game is the long-lived mutable owner of the whole application state.
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
		enemies[i].y = road.horizonY + float64(45*i)
	}

	scoreManager.LoadBestScore()

	return &Game{
		road:         road,
		player:       player,
		enemies:      enemies,
		scoreManager: scoreManager,
		soundManager: NewSoundManager(),
		blockedLanes: make([]float64, 0, len(enemies)-1),
		occupiedY:    make([]float64, 0, len(enemies)-1),
		started:      false,
	}
}

// Reset starts a new run while keeping persistent best-score data and cached systems alive.
func (g *Game) Reset() {
	g.road = NewRoad()
	g.player = NewPlayer(g.road)

	g.enemies = []Enemy{
		NewEnemy(g.road),
		NewEnemy(g.road),
		NewEnemy(g.road),
	}

	for i := 1; i < len(g.enemies); i++ {
		g.enemies[i].y = g.road.horizonY + float64(45*i)
	}

	g.scoreManager.ResetScore()
	g.soundManager.OnReset()
	g.paused = false
	g.gameOver = false
	g.timeOfDay = 0
	g.hudCacheText = ""
}

// handlePauseToggle performs edge-triggered pause input.
//
// The game checks for the transition from "not pressed" to "pressed" so holding P does not toggle repeatedly.
func (g *Game) handlePauseToggle() {
	pPressed := ebiten.IsKeyPressed(ebiten.KeyP)
	if pPressed && !g.pKeyDown && g.started && !g.gameOver {
		g.paused = !g.paused
		g.soundManager.OnPauseChanged(g.paused)
	}
	g.pKeyDown = pPressed
}

// Update advances the game simulation by one frame.
//
// This is the heart of the game loop.
// A useful way to read it is as a state machine:
//   1. Handle global input like pause.
//   2. If the title screen is active, only wait for the start input.
//   3. If the game is over, only wait for reset input.
//   4. If paused, skip simulation.
//   5. Otherwise update player, road, enemies, scoring, sound, and collisions.
func (g *Game) Update() error {
	g.handlePauseToggle()

	if !g.started {
		g.timeOfDay += 1.0 / 60.0
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			g.started = true
			g.soundManager.OnGameStart()
		}
		return nil
	}

	if g.gameOver {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			g.Reset()
		}
		return nil
	}

	if g.paused {
		return nil
	}

	g.player.Update(g.road)
	g.soundManager.UpdateEngine(g.player.Speed(), g.paused, g.gameOver)
	g.road.SetSpeed(g.player.Speed())
	g.road.Update()

	g.timeOfDay += 1.0 / 60.0

	for i := range g.enemies {
		g.enemies[i].Update(g.player.Speed())

		if g.enemies[i].IsBelowScreen() || (g.enemies[i].IsAboveHorizon(g.road) && g.player.Speed() <= playerMinSpeed) || g.enemies[i].HasExpired() {
			if g.enemies[i].IsBelowScreen() {
				g.scoreManager.UpdateScore()
			}

			g.blockedLanes = g.blockedLanes[:0]
			g.occupiedY = g.occupiedY[:0]
			for j := range g.enemies {
				if j == i {
					continue
				}
				g.blockedLanes = append(g.blockedLanes, g.enemies[j].LaneOffset())
				g.occupiedY = append(g.occupiedY, g.enemies[j].y)
			}

			g.enemies[i].Respawn(g.road, g.blockedLanes, g.player.Speed(), g.occupiedY)
		}

		if g.player.IsColliding(g.enemies[i].Rect(g.road)) {
			g.soundManager.OnCrash()
			g.gameOver = true
		}
	}

	return nil
}

// hudText returns the HUD string shown during play or special states.
//
// It caches the generated text because the visible text often stays the same for many frames.
func (g *Game) hudText() string {
	key := hudCacheKey{
		started:    g.started,
		paused:     g.paused,
		gameOver:   g.gameOver,
		score:      g.scoreManager.Score(),
		bestScore:  g.scoreManager.BestScore(),
		speedKPH:   g.displaySpeedKPH(),
		newBest:    g.scoreManager.HasNewBest(),
	}
	if key == g.hudCacheKey {
		return g.hudCacheText
	}

	g.hudCacheKey = key

	if !g.started {
		g.hudCacheText = ""
		return g.hudCacheText
	}

	if g.gameOver {
		var builder strings.Builder
		fmt.Fprintf(&builder, "GAME OVER\nScore: %d\nBest: %d", key.score, key.bestScore)
		if key.newBest {
			builder.WriteString("\nNEW BEST!")
		}
		builder.WriteString("\nPress R to restart")
		g.hudCacheText = builder.String()
		return g.hudCacheText
	}

	if g.paused {
		g.hudCacheText = fmt.Sprintf(
			"PAUSED\nScore: %d\nBest: %d\nPress P to resume",
			key.score,
			key.bestScore,
		)
		return g.hudCacheText
	}

	g.hudCacheText = fmt.Sprintf(
		"Score: %d\nBest: %d\nSpeed: %d km/h\nMove: Left/Right\nAccel: Up  Brake: Down\nPause: P",
		key.score,
		key.bestScore,
		key.speedKPH,
	)
	return g.hudCacheText
}

// drawTitleScreen renders the title and instructions shown before gameplay starts.
func (g *Game) drawTitleScreen(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen, "ENDURO GO", 8, 8)
	ebitenutil.DebugPrintAt(screen, "Old time road racing in Go", 58, 62)
	ebitenutil.DebugPrintAt(screen, titleBestScoreText(g.scoreManager.BestScore()), 96, 90)
	ebitenutil.DebugPrintAt(screen, "Arrow keys steer", 103, 178)
	ebitenutil.DebugPrintAt(screen, "Up/Down controls speed", 86, 192)

	if math.Mod(g.timeOfDay, 1.0) < 0.5 {
		ebitenutil.DebugPrintAt(screen, "Press SPACE to start", 96, 122)
	}
}

// sceneLight returns a normalized daylight amount used by the day/night cycle.
func (g *Game) sceneLight() float64 {
	dayNightCycleSeconds := 40.0

	phase := math.Mod(g.timeOfDay/dayNightCycleSeconds, 2.0)
	if phase > 1 {
		phase = 2 - phase
	}

	return 1.0 - phase
}

// skyColor picks the sky color for the current time of day.
func (g *Game) skyColor() color.RGBA {
	dayNightCycleSeconds := 40.0
	phase := math.Mod(g.timeOfDay/dayNightCycleSeconds, 2.0)

	day := color.RGBA{40, 80, 220, 255}
	dusk := color.RGBA{220, 120, 60, 255}
	night := color.RGBA{10, 10, 40, 255}

	if phase < 0.5 {
		return lerpColor(day, dusk, phase/0.5)
	}
	if phase < 1.0 {
		return lerpColor(dusk, night, (phase-0.5)/0.5)
	}
	if phase < 1.5 {
		return lerpColor(night, dusk, (phase-1.0)/0.5)
	}
	return lerpColor(dusk, day, (phase-1.5)/0.5)
}

// visibility converts scene light into a general scene visibility factor.
func (g *Game) visibility() float64 {
	return 0.20 + g.sceneLight()*0.80
}

// displaySpeedKPH maps gameplay speed into a player-facing km/h value for the HUD.
func (g *Game) displaySpeedKPH() int {
	ratio := (g.player.Speed() - playerMinSpeed) / (playerMaxSpeed - playerMinSpeed)
	displaySpeed := int(lerp(80, 200, ratio))

	if displaySpeed < 80 {
		return 80
	}
	if displaySpeed > 200 {
		return 200
	}

	return displaySpeed
}

// visibilityFromLight is a small helper that maps light to visibility.
func visibilityFromLight(sceneLight float64) float64 {
	return 0.20 + sceneLight*0.80
}

// Draw renders one complete frame.
//
// Important game-loop note:
// Update changes the state.
// Draw only reads the current state and turns it into pixels.
// Keeping those responsibilities separate makes the program much easier to reason about.
func (g *Game) Draw(screen *ebiten.Image) {
	sceneLight := g.sceneLight()
	visibility := visibilityFromLight(sceneLight)

	g.road.Draw(screen, g.skyColor(), sceneLight, visibility)
	g.player.Draw(screen, g.road)

	for _, enemy := range g.enemies {
		enemy.Draw(screen, g.road, visibility)
	}

	if !g.started {
		g.drawTitleScreen(screen)
		return
	}

	ebitenutil.DebugPrintAt(screen, g.hudText(), 8, 8)
}

// Layout tells Ebiten the internal logical resolution used by the game.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// main configures the window and starts the Ebiten game loop.
func main() {
	ebiten.SetWindowSize(screenWidth*windowScale, screenHeight*windowScale)
	ebiten.SetWindowTitle("Enduro GO - Milestone 1")

	game := NewGame()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
