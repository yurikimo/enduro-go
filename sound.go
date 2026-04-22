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

type SoundManager struct {
	context       *audio.Context
	enginePlayers []*audio.Player
	currentEngine int
	effectPlayers []*audio.Player
	startBeep     []byte
	pauseBlip     []byte
	crashBurst    []byte
	gameStart     []byte
}

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
		startBeep:     synthSquareTone(880, 0.08, 0.18),
		pauseBlip:     synthSquareTone(520, 0.06, 0.14),
		crashBurst:    synthCrashBurst(0.28, 0.45),
		gameStart: synthSequence(
			synthSquareTone(660, 0.06, 0.16),
			silencePCM(0.02),
			synthSquareTone(880, 0.06, 0.16),
			silencePCM(0.02),
			synthSquareTone(1175, 0.10, 0.18),
		),
	}
}

func (s *SoundManager) OnGameStart() {
	s.playEffect(s.gameStart)
}

func (s *SoundManager) OnPauseChanged(paused bool) {
	if paused {
		s.pauseEngine()
		s.playEffect(s.pauseBlip)
		return
	}

	s.playEffect(s.pauseBlip)
}

func (s *SoundManager) UpdateEngine(playerSpeed float64, paused bool, gameOver bool) {
	s.cleanupFinishedEffects()

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

func (s *SoundManager) OnCrash() {
	s.pauseEngine()
	s.playEffect(s.crashBurst)
}

func (s *SoundManager) OnReset() {
	s.pauseEngine()
	s.playEffect(s.startBeep)
}

func (s *SoundManager) pauseEngine() {
	for _, player := range s.enginePlayers {
		player.Pause()
	}
	s.currentEngine = -1
}

func (s *SoundManager) playEffect(sample []byte) {
	player := s.context.NewPlayerFromBytes(sample)
	player.Play()
	s.effectPlayers = append(s.effectPlayers, player)
}

func (s *SoundManager) cleanupFinishedEffects() {
	active := s.effectPlayers[:0]
	for _, player := range s.effectPlayers {
		if player.IsPlaying() {
			active = append(active, player)
			continue
		}

		_ = player.Close()
	}
	s.effectPlayers = active
}

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

func silencePCM(durationSeconds float64) []byte {
	frameCount := int(float64(audioSampleRate) * durationSeconds)
	return make([]byte, frameCount*4)
}

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

func attackReleaseEnvelope(progress float64, attack, release float64) float64 {
	if progress < attack {
		return progress / attack
	}
	if progress > 1.0-release {
		return (1.0 - progress) / release
	}
	return 1.0
}

func pseudoNoise(index int) float64 {
	value := math.Sin(float64(index)*12.9898) * 43758.5453
	return 2*(value-math.Floor(value)) - 1
}

func appendPCM16Stereo(data []byte, left, right int16) []byte {
	return append(
		data,
		byte(left), byte(left>>8),
		byte(right), byte(right>>8),
	)
}
