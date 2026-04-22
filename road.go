package main

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	roadBottomWidth = (screenWidth / 5) * 4
	roadTopWidth    = 2.0
	roadHorizonY    = 60.0
	edgeLineWidth   = 2.0
	roadScrollScale = 0.02

	curveMaxOffset    = 60.0
	curveSmoothness   = 0.02
	curveMinFrames    = 180
	curveMaxFrames    = 360
	curveStraightBias = 0.30
)

// Road stores the perspective road and its scrolling/curvature state.
//
// The road is responsible for two important teaching ideas in this project:
//   1. Perspective: the road gets narrower toward the horizon.
//   2. Curvature: the road center shifts sideways over time.
//
// Other systems query the road to convert abstract lane positions into concrete screen positions.
type Road struct {
	x           float64
	bottomWidth float64
	topWidth    float64
	horizonY    float64
	lineOffsetY float64
	speed       float64

	curveOffset float64
	curveTarget float64
	curveFrames int
}

// NewRoad creates the default road centered on screen.
func NewRoad() Road {
	centerX := float64(screenWidth) / 2

	road := Road{
		x:           centerX,
		bottomWidth: roadBottomWidth,
		topWidth:    roadTopWidth,
		horizonY:    roadHorizonY,
		lineOffsetY: 0,
		speed:       0,
		curveOffset: 0,
		curveTarget: 0,
		curveFrames: randomCurveFrames(),
	}

	return road
}

// randomCurveFrames picks how long the next road curve should last.
func randomCurveFrames() int {
	return curveMinFrames + rand.Intn(curveMaxFrames-curveMinFrames+1)
}

// randomCurveTarget chooses the next horizontal curve offset.
func randomCurveTarget() float64 {
	if rand.Float64() < curveStraightBias {
		return 0
	}

	return (rand.Float64()*2 - 1) * curveMaxOffset
}

// Left returns the road's left boundary at the bottom of the screen.
func (r *Road) Left() float64 {
	left, _ := r.BoundsAt(float64(screenHeight))
	return left
}

// Right returns the road's right boundary at the bottom of the screen.
func (r *Road) Right() float64 {
	_, right := r.BoundsAt(float64(screenHeight))
	return right
}

// PlayerRightLimit returns the maximum player X value that still keeps the car on the road.
func (r *Road) PlayerRightLimit(playerWidth float64) float64 {
	return r.Right() - playerWidth
}

// SetSpeed updates the road scroll speed.
func (r *Road) SetSpeed(speed float64) {
	r.speed = speed
}

// CenterXAt returns the road center at a given Y coordinate.
func (r *Road) CenterXAt(y float64) float64 {
	left, right := r.roadEdgesAt(y)
	return (left + right) / 2
}

// BoundsAt returns the left and right road edges for a given screen Y.
//
// This is one of the most important perspective helpers in the project.
// Near the horizon the road is narrow, and near the player it is wide.
// Enemies and the player use this to stay visually attached to the road.
func (r *Road) BoundsAt(y float64) (float64, float64) {
	return r.roadEdgesAt(y)
}

// Update advances the lane markers and smoothly steers the road toward its current curve target.
func (r *Road) Update() {
	r.lineOffsetY += roadScrollScale * r.speed

	for r.lineOffsetY >= 1.0 {
		r.lineOffsetY -= 1.0
	}
	for r.lineOffsetY < 0 {
		r.lineOffsetY += 1.0
	}

	r.curveOffset = lerp(r.curveOffset, r.curveTarget, curveSmoothness)

	r.curveFrames--
	if r.curveFrames <= 0 {
		r.curveTarget = randomCurveTarget()
		r.curveFrames = randomCurveFrames()
	}
}
// CurveOffset returns the current horizontal bend amount.
func (r *Road) CurveOffset() float64 {
	return r.curveOffset
}

// curveShiftAt returns how much the road center is shifted sideways at a given Y.
//
// The shift is stronger near the horizon and weaker near the camera to mimic classic pseudo-3D road motion.
func (r *Road) curveShiftAt(y float64) float64 {
	if y < r.horizonY {
		y = r.horizonY
	}
	if y > float64(screenHeight) {
		y = float64(screenHeight)
	}

	progress := (y - r.horizonY) / (float64(screenHeight) - r.horizonY)
	horizonWeight := 1.0 - progress

	return r.curveOffset * horizonWeight * horizonWeight
}

// roadEdgesAt is the core perspective calculation for the road.
//
// Given a screen Y, it computes:
//   - the road width at that depth
//   - the curved center X at that depth
//   - the final left and right edges
//
// Many other systems build on this function, so understanding it is key to understanding the project.
func (r *Road) roadEdgesAt(y float64) (float64, float64) {
	if y < r.horizonY {
		y = r.horizonY
	}
	if y > float64(screenHeight) {
		y = float64(screenHeight)
	}

	progress := (y - r.horizonY) / (float64(screenHeight) - r.horizonY)
	width := r.topWidth + (r.bottomWidth-r.topWidth)*progress

	centerX := r.x + r.curveShiftAt(y)

	left := centerX - width/2
	right := centerX + width/2

	return left, right
}

// scaleColor darkens or brightens a color while preserving alpha.
func scaleColor(base color.RGBA, light float64) color.RGBA {
	if light < 0 {
		light = 0
	}
	if light > 1 {
		light = 1
	}

	minBrightness := 0.35
	factor := minBrightness + light*(1.0-minBrightness)

	return color.RGBA{
		R: uint8(float64(base.R) * factor),
		G: uint8(float64(base.G) * factor),
		B: uint8(float64(base.B) * factor),
		A: base.A,
	}
}

// tintEnvironmentColor adds simple day/dusk/night tinting to a base color.
func tintEnvironmentColor(base color.RGBA, sceneLight float64) color.RGBA {
	if sceneLight < 0 {
		sceneLight = 0
	}
	if sceneLight > 1 {
		sceneLight = 1
	}

	duskStrength := 1.0 - math.Abs(sceneLight-0.5)*2.0
	if duskStrength < 0 {
		duskStrength = 0
	}

	nightStrength := 1.0 - sceneLight

	warmed := lerpColor(base, color.RGBA{166, 94, 58, 255}, duskStrength*0.18)
	return lerpColor(warmed, color.RGBA{34, 48, 82, 255}, nightStrength*0.22)
}

// horizonColor returns the current dune color.
func horizonColor(sceneLight float64) color.RGBA {
	return scaleColor(color.RGBA{200, 190, 90, 255}, sceneLight)
}

// duneHighlightColor returns the dune highlight color.
func duneHighlightColor(sceneLight float64) color.RGBA {
	return scaleColor(color.RGBA{224, 206, 120, 255}, sceneLight)
}

// propColor returns the current background prop color.
func propColor(sceneLight float64) color.RGBA {
	return scaleColor(color.RGBA{92, 76, 38, 255}, sceneLight)
}

// drawDune draws a simple layered dune shape at the horizon.
func drawDune(screen *ebiten.Image, x, horizonY, w, h float64, baseColor, highlightColor color.RGBA) {
	for i := 0.0; i < h; i++ {
		progress := i / h
		currentWidth := w * (0.25 + 0.75*progress)
		left := x + (w-currentWidth)/2
		y := horizonY - h + i

		rowColor := baseColor
		if progress < 0.35 {
			rowColor = highlightColor
		}

		ebitenutil.DrawRect(screen, left, y, currentWidth, 1, rowColor)
	}
}

// drawRockSpire draws a simple background rock shape.
func drawRockSpire(screen *ebiten.Image, x, horizonY, w, h float64, rockColor color.RGBA) {
	for i := 0.0; i < h; i++ {
		progress := i / h
		currentWidth := w * (0.35 + 0.65*progress)
		left := x + (w-currentWidth)/2
		y := horizonY - h + i

		ebitenutil.DrawRect(screen, left, y, currentWidth, 1, rockColor)
	}
}

// drawRoadsidePost draws a small roadside post near the horizon.
func drawRoadsidePost(screen *ebiten.Image, x, horizonY, h float64, postColor color.RGBA) {
	postWidth := 2.0
	ebitenutil.DrawRect(screen, x, horizonY-h, postWidth, h, postColor)
	ebitenutil.DrawRect(screen, x-2, horizonY-h, 6, 2, postColor)
}

// applyVisibility fades colors based on scene visibility and distance.
func applyVisibility(base color.RGBA, visibility float64, distanceFactor float64) color.RGBA {
	if visibility < 0 {
		visibility = 0
	}
	if visibility > 1 {
		visibility = 1
	}
	if distanceFactor < 0 {
		distanceFactor = 0
	}
	if distanceFactor > 1 {
		distanceFactor = 1
	}

	effective := visibility + (1.0-visibility)*(distanceFactor*0.5)

	return color.RGBA{
		R: uint8(float64(base.R) * effective),
		G: uint8(float64(base.G) * effective),
		B: uint8(float64(base.B) * effective),
		A: base.A,
	}
}

// drawCurvedMarker draws one lane marker segment following the road's curve at each scanline.
func drawCurvedMarker(screen *ebiten.Image, road *Road, y, width, height float64, markerColor color.RGBA) {
	for i := 0.0; i < height; i++ {
		sliceY := y + i

		left, right := road.roadEdgesAt(sliceY)
		centerX := (left + right) / 2

		ebitenutil.DrawRect(screen, centerX-width/2, sliceY, width, 1, markerColor)
	}
}

// Draw renders the entire road scene from back to front.
//
// Teaching note:
// The order matters.
// First the background is drawn, then the ground, then the road body, then the road edges and markers.
// This "paint from farthest to nearest" approach is a simple way to manage layering in 2D games.
func (r *Road) Draw(screen *ebiten.Image, skyColor color.RGBA, sceneLight float64, visibility float64) {
	groundColor := tintEnvironmentColor(scaleColor(color.RGBA{34, 139, 34, 255}, sceneLight), sceneLight)
	roadColor := tintEnvironmentColor(scaleColor(color.RGBA{70, 70, 70, 255}, sceneLight), sceneLight)
	lineColor := scaleColor(color.RGBA{240, 240, 240, 255}, sceneLight)

	screen.Fill(skyColor)
	ebitenutil.DrawRect(screen, 0, r.horizonY, float64(screenWidth), float64(screenHeight)-r.horizonY, groundColor)

	horizonShift := r.curveOffset * 1.15
	duneColor := horizonColor(sceneLight)
	duneHighlight := duneHighlightColor(sceneLight)
	backgroundProp := propColor(sceneLight)

	drawDune(screen, 26+horizonShift*0.55, r.horizonY, 88, 16, duneColor, duneHighlight)
	drawDune(screen, 204+horizonShift*0.55, r.horizonY, 76, 13, duneColor, duneHighlight)
	drawRockSpire(screen, 124+horizonShift*0.8, r.horizonY, 10, 18, backgroundProp)
	drawRoadsidePost(screen, 252+horizonShift*0.95, r.horizonY, 12, backgroundProp)

	for y := int(r.horizonY); y < screenHeight; y++ {
		left, right := r.roadEdgesAt(float64(y))
		ebitenutil.DrawRect(screen, left, float64(y), right-left, 1, roadColor)
	}

	for y := int(r.horizonY); y < screenHeight; y++ {
		left, right := r.roadEdgesAt(float64(y))
		ebitenutil.DrawRect(screen, left, float64(y), edgeLineWidth, 1, lineColor)
		ebitenutil.DrawRect(screen, right-edgeLineWidth, float64(y), edgeLineWidth, 1, lineColor)
	}

	markerCount := 8.0
	for i := 0.0; i < markerCount; i++ {
		progress := (i + r.lineOffsetY) / markerCount
		if progress > 1 {
			progress -= 1
		}

		yProgress := progress * progress
		y := r.horizonY + yProgress*(float64(screenHeight)-r.horizonY)

		lineWidth := 1.0 + progress*3.0
		lineHeight := 3.0 + progress*16.0
		markerColor := applyVisibility(lineColor, visibility, progress)

		drawCurvedMarker(screen, r, y, lineWidth, lineHeight, markerColor)
	}
}
