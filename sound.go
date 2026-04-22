package main

import (
	"bytes"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	audioSampleRate = 44100
	maxPCM16        = 32767
)

// SoundManager owns the small synthesized sound effects used by the game.
//
// The project generates audio procedurally instead of loading asset files so learners can see
// that sound in Ebiten is ultimately just PCM byte data played through an audio player.
//
// Some fields are pointers or slices because the underlying Ebiten audio types are mutable runtime objects.
// The manager keeps and reuses them instead of recreating them every frame.
type SoundManager struct {
	context       *audio.Context
	enginePlayers []*audio.Player
	currentEngine int
	startBeep     *audio.Player
	pauseBlip     *audio.Player
	crashBurst    *audio.Player
	gameStart     *audio.Player
}

// NewSoundManager builds all audio players once at startup.
//
// Returning SoundManager by value is fine here because the struct mainly stores pointers and slices.
// The expensive runtime objects are still shared through those references.
func NewSoundManager() SoundManager {
	context := audio.NewContext(audioSampleRate)

	engineLoops := [][]byte{
		synthEngineLoop(42, 0.20),
		synthEngineLoop(58, 0.22),
		synthEngineLoop(74, 0.24),
		synthEngineLoop(96, 0.26),
	}

	enginePlayers := make([]*audio.Player, 0, len(engineLoops))
	for _, loopData := range engineLoops {
		loop := audio.NewInfiniteLoop(bytes.NewReader(loopData), int64(len(loopData)))
		player, err := context.NewPlayer(loop)
		if err != nil {
			panic(err)
		}
		player.SetVolume(0)
		enginePlayers = append(enginePlayers, player)
	}

	return SoundManager{
		context:       context,
		enginePlayers: enginePlayers,
		currentEngine: -1,
		startBeep: prepareEffectPlayer(context, synthSquareTone(880, 0.08, 0.18)),
		pauseBlip: prepareEffectPlayer(context, synthSquareTone(520, 0.06, 0.14)),
		crashBurst: prepareEffectPlayer(context, synthCrashBurst(0.28, 0.45)),
		gameStart: prepareEffectPlayer(context, synthSequence(
			synthSquareTone(660, 0.06, 0.16),
			silencePCM(0.02),
			synthSquareTone(880, 0.06, 0.16),
			silencePCM(0.02),
			synthSquareTone(1175, 0.10, 0.18),
		)),
	}
}

// OnGameStart plays the short ascending jingle used when the game begins.
func (s *SoundManager) OnGameStart() {
	s.playEffect(s.gameStart)
}

// OnPauseChanged reacts to pause toggles.
//
// When the game is paused, the engine loop is stopped so the game feels frozen instead of still moving.
func (s *SoundManager) OnPauseChanged(paused bool) {
	if paused {
		s.pauseEngine()
		s.playEffect(s.pauseBlip)
		return
	}

	s.playEffect(s.pauseBlip)
}

// UpdateEngine selects and plays a looping engine sound that matches the current speed.
//
// This is called every frame, so it avoids rebuilding players or sample data.
// Only the active loop and volume are updated.
func (s *SoundManager) UpdateEngine(playerSpeed float64, paused bool, gameOver bool) {
	if paused || gameOver || playerSpeed <= 0 {
		s.pauseEngine()
		return
	}

	engineIndex := selectEngineLoop(playerSpeed)
	if engineIndex != s.currentEngine {
		s.pauseEngine()
		s.currentEngine = engineIndex
	}

	player := s.enginePlayers[s.currentEngine]
	player.SetVolume(engineVolume(playerSpeed))
	if !player.IsPlaying() {
		_ = player.Rewind()
		player.Play()
	}
}

// OnCrash stops the engine loop and plays the crash effect.
func (s *SoundManager) OnCrash() {
	s.pauseEngine()
	s.playEffect(s.crashBurst)
}

// OnReset restores the silent state used at the beginning of a new run.
func (s *SoundManager) OnReset() {
	s.pauseEngine()
	s.playEffect(s.startBeep)
}

// pauseEngine pauses every engine loop and forgets which loop was active.
//
// Resetting currentEngine to -1 forces UpdateEngine to re-evaluate the correct loop next time.
func (s *SoundManager) pauseEngine() {
	for _, player := range s.enginePlayers {
		player.Pause()
	}
	s.currentEngine = -1
}

// prepareEffectPlayer creates a reusable player for a short one-shot effect.
//
// The sample bytes are prepared once, then rewound and replayed whenever needed.
func prepareEffectPlayer(context *audio.Context, sample []byte) *audio.Player {
	player := context.NewPlayerFromBytes(sample)
	return player
}

// playEffect rewinds a one-shot effect and plays it from the start.
func (s *SoundManager) playEffect(player *audio.Player) {
	player.Pause()
	_ = player.Rewind()
	player.Play()
}

// selectEngineLoop maps the current speed to one of the prebuilt engine loops.
//
// The thresholds are simple by design so learners can easily tune them.
func selectEngineLoop(playerSpeed float64) int {
	switch {
	case playerSpeed < 1.8:
		return 0
	case playerSpeed < 3.0:
		return 1
	case playerSpeed < 4.5:
		return 2
	default:
		return 3
	}
}

// engineVolume converts gameplay speed into an audio volume in a safe range.
func engineVolume(playerSpeed float64) float64 {
	normalized := (playerSpeed - playerMinSpeed) / (playerMaxSpeed - playerMinSpeed)
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}

	return 0.10 + normalized*0.22
}

// synthSequence concatenates several PCM buffers into a single sample.
func synthSequence(parts ...[]byte) []byte {
	totalLength := 0
	for _, part := range parts {
		totalLength += len(part)
	}

	sequence := make([]byte, 0, totalLength)
	for _, part := range parts {
		sequence = append(sequence, part...)
	}

	return sequence
}

// silencePCM creates silent stereo PCM data for spacing between notes.
func silencePCM(durationSeconds float64) []byte {
	frameCount := int(float64(audioSampleRate) * durationSeconds)
	return make([]byte, frameCount*4)
}

// synthSquareTone creates a simple square-wave tone with a tiny attack/release envelope.
//
// The envelope reduces clicks at the beginning and end of the sound.
func synthSquareTone(frequency float64, durationSeconds float64, amplitude float64) []byte {
	frameCount := int(float64(audioSampleRate) * durationSeconds)
	data := make([]byte, 0, frameCount*4)

	for i := 0; i < frameCount; i++ {
		progress := float64(i) / float64(frameCount)
		envelope := attackReleaseEnvelope(progress, 0.08, 0.24)

		value := -1.0
		if math.Sin(2*math.Pi*frequency*float64(i)/audioSampleRate) >= 0 {
			value = 1.0
		}

		sample := int16(maxPCM16 * amplitude * envelope * value)
		data = appendPCM16Stereo(data, sample, sample)
	}

	return data
}

// synthEngineLoop creates a short repeating engine-like sample.
//
// It mixes a square wave with a harmonic and slight vibrato so the loop feels less flat.
func synthEngineLoop(baseFrequency float64, amplitude float64) []byte {
	const durationSeconds = 0.24

	frameCount := int(float64(audioSampleRate) * durationSeconds)
	data := make([]byte, 0, frameCount*4)

	for i := 0; i < frameCount; i++ {
		t := float64(i) / audioSampleRate
		vibrato := 1.0 + 0.06*math.Sin(2*math.Pi*5*t)
		frequency := baseFrequency * vibrato

		square := -1.0
		if math.Sin(2*math.Pi*frequency*t) >= 0 {
			square = 1.0
		}

		harmonic := math.Sin(2 * math.Pi * frequency * 2 * t)
		value := square*0.78 + harmonic*0.22
		sample := int16(maxPCM16 * amplitude * value)

		data = appendPCM16Stereo(data, sample, sample)
	}

	return data
}

// synthCrashBurst creates a noisy, decaying crash effect.
func synthCrashBurst(durationSeconds float64, amplitude float64) []byte {
	frameCount := int(float64(audioSampleRate) * durationSeconds)
	data := make([]byte, 0, frameCount*4)

	for i := 0; i < frameCount; i++ {
		progress := float64(i) / float64(frameCount)
		noise := pseudoNoise(i)
		crunch := math.Sin(2*math.Pi*80*float64(i)/audioSampleRate) * 0.25
		envelope := math.Pow(1.0-progress, 2.4)

		value := (noise*0.85 + crunch) * amplitude * envelope
		sample := int16(maxPCM16 * value)
		data = appendPCM16Stereo(data, sample, sample)
	}

	return data
}

// attackReleaseEnvelope returns a simple fade-in/fade-out multiplier for a normalized sound position.
func attackReleaseEnvelope(progress float64, attack, release float64) float64 {
	if progress < attack {
		return progress / attack
	}
	if progress > 1.0-release {
		return (1.0 - progress) / release
	}
	return 1.0
}

// pseudoNoise returns a repeatable pseudo-random value in the range -1..1.
//
// It is deterministic, which is enough here because the effect only needs a noisy texture.
func pseudoNoise(index int) float64 {
	value := math.Sin(float64(index)*12.9898) * 43758.5453
	return 2*(value-math.Floor(value)) - 1
}

// appendPCM16Stereo appends one stereo 16-bit PCM frame to the destination buffer.
func appendPCM16Stereo(data []byte, left, right int16) []byte {
	return append(
		data,
		byte(left), byte(left>>8),
		byte(right), byte(right>>8),
	)
}
