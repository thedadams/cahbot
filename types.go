package main

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/lib/pq"
	"github.com/thedadams/telegram-bot-api"
)

// CAHBot inherits from tgbotapi.
type CAHBot struct {
	*tgbotapi.BotAPI
	DBConn           *sql.DB
	AllQuestionCards []QuestionCard `json:"all_question_cards"`
	AllAnswerCards   []AnswerCard   `json:"all_answer_cards"`
}

// NewCAHBot creates a new CAHBot.
func NewCAHBot(token string) (*CAHBot, error) {
	GenericBot, err := tgbotapi.NewBotAPI(token)
	// Need to get the card data
	var AllQuestionCards []QuestionCard
	err = json.Unmarshal(AllQuestions, &AllQuestionCards)
	if err != nil {
		log.Printf("%v", err)
	}
	var AllAnswerCards []AnswerCard
	err = json.Unmarshal(AllAnswers, &AllAnswerCards)
	if err != nil {
		log.Printf("%v", err)
	}
	db, err := sql.Open("postgres", "sslmode=disable user=cahbot dbname=cahgames password="+DBPass)
	if err != nil {
		log.Fatal(err)
	}
	return &CAHBot{GenericBot, db, AllQuestionCards, AllAnswerCards}, err
}

// QuestionCard represents a white card in CAH.
type QuestionCard struct {
	ID         int    `json:"id"`
	Text       string `json:"text"`
	NumAnswers int    `json:"numAnswers"`
	Expansion  string `json:"expansion"`
}

// AnswerCard represents a black card in CAH.
type AnswerCard struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	Expansion string `json:"expansion"`
}
