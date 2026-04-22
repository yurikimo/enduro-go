package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

const bestScoreFile = "best_score.txt"

// ScoreManager owns the current score and the persistent best score.
//
// This logic is isolated from Game so the scoring rules stay easy to read and test.
type ScoreManager struct {
	score     int
	bestScore int
	newBest   bool
}

// NewScoreManager creates a score manager with a zeroed current score.
func NewScoreManager() ScoreManager {
	return ScoreManager{
		score: 0,
	}
}

// Score returns the current run score.
func (s *ScoreManager) Score() int {
	return s.score
}

// BestScore returns the saved best score.
func (s *ScoreManager) BestScore() int {
	return s.bestScore
}

// ResetScore clears the current score for a new run.
func (s *ScoreManager) ResetScore() {
	s.score = 0
	s.newBest = false
}

// UpdateScore increases the score and updates the best score if needed.
func (s *ScoreManager) UpdateScore() {
	s.score++

	if s.score > s.bestScore {
		s.bestScore = s.score
		s.newBest = true

		s.SaveBestScore()
	}
}

// HasNewBest reports whether the current run has set a new best score.
func (s *ScoreManager) HasNewBest() bool {
	return s.newBest
}

// LoadBestScore loads the best score from disk if the file exists.
//
// Errors are intentionally treated gently because failing to load a score should not stop the game.
func (s *ScoreManager) LoadBestScore() {
	data, err := os.ReadFile(bestScoreFile)
	if err != nil {
		return
	}

	scoreText := strings.TrimSpace(string(data))
	s.bestScore, err = strconv.Atoi(scoreText)
	if err != nil {
		return
	}

	if s.bestScore < 0 {
		s.bestScore = 0
	}
}

// SaveBestScore writes the best score to disk.
func (s *ScoreManager) SaveBestScore() {
	content := strconv.Itoa(s.bestScore)
	err := os.WriteFile(bestScoreFile, []byte(content), 0644)
	if err != nil {
		log.Println("could not save best score:", err)
	}
}
