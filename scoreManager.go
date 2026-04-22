package main

import (
	"log"
	"os"
	"strconv"
	"strings"
)

const bestScoreFile = "best_score.txt"

type ScoreManager struct {
	score     int
	bestScore int
}

func NewScoreManager() ScoreManager {
	return ScoreManager{
		score: 0,
	}
}

func (s *ScoreManager) Score() int {
	return s.score
}

func (s *ScoreManager) BestScore() int {
	return s.bestScore
}

func (s *ScoreManager) ResetScore() {
	s.score = 0
}

func (s *ScoreManager) UpdateScore() {
	s.score++

	if s.score > s.bestScore {
		s.bestScore = s.score

		s.SaveBestScore()
	}
}

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

func (s *ScoreManager) SaveBestScore() {
	content := strconv.Itoa(s.score)
	err := os.WriteFile(bestScoreFile, []byte(content), 0644)
	if err != nil {
		log.Println("could not save best score:", err)
	}
}
