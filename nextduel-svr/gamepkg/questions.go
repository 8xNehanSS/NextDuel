package gamepkg

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

type OpenTDBResponse struct {
	Results []OpenTDBQuestion `json:"results"`
}

type OpenTDBQuestion struct {
	Category         string   `json:"category"`
	Type             string   `json:"type"`
	Difficulty       string   `json:"difficulty"`
	Question         string   `json:"question"`
	CorrectAnswer    string   `json:"correct_answer"`
	IncorrectAnswers []string `json:"incorrect_answers"`
}

type Question struct {
	Round      int      `json:"round"`
	Question   string   `json:"question"`
	Answer     string   `json:"answer"`
	Category   string   `json:"category"`
	Difficulty string   `json:"difficulty"`
	AllAnswers []string `json:"all_answers"`
}

type QuestionManager struct {
	ID        string `json:"id"`
	questions []Question
}

func (qm *QuestionManager) GetQuestion(round int) *Question {
	for _, q := range qm.questions {
		if q.Round == round {
			return &q
		}
	}
	return nil
}

func NewQuestionManager(rounds int) *QuestionManager {
	questions, err := getNewQuestions(rounds)
	if err != nil {
		panic("Failed to fetch questions: " + err.Error())
	}
	return &QuestionManager{
		ID:        uuid.NewString(),
		questions: questions,
	}
}

func getNewQuestions(rounds int) ([]Question, error) {
	url := "https://opentdb.com/api.php?amount=" + fmt.Sprintf("%d", rounds) + "&difficulty=easy&type=multiple"

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp OpenTDBResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	questions := make([]Question, 0, len(apiResp.Results))

	round := 1
	for _, q := range apiResp.Results {

		// randomize the order of answers
		allAnswers := append([]string{q.CorrectAnswer}, q.IncorrectAnswers...)
		for i := range allAnswers {
			j := i + (rand.Intn(len(allAnswers) - i))
			allAnswers[i], allAnswers[j] = allAnswers[j], allAnswers[i]
		}

		questions = append(questions, Question{
			Round:      round,
			Question:   q.Question,
			Answer:     q.CorrectAnswer,
			Category:   q.Category,
			Difficulty: q.Difficulty,
			AllAnswers: allAnswers,
		})
		round++
	}

	return questions, nil
}
