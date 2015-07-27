package main

import (
	"cahbot/secrets"
	"cahbot/tgbotapi"
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"log"
)

// A wrapper for tgbotapi. We need this wrapper to add new methods.
type CAHBot struct {
	*tgbotapi.BotAPI
	db_conn          *sql.DB
	AllQuestionCards []QuestionCard `json:"all_question_cards"`
	AllAnswerCards   []AnswerCard   `json:"all_answer_cards"`
}

// This creates a new CAHBot, which is basically a wrapper for tgbotapi.BotAPI.
// We need this wrapper to add the desired methods.
func NewCAHBot(token string) (*CAHBot, error) {
	GenericBot, err := tgbotapi.NewBotAPI(token)
	// Need to get the card data
	var AllQuestionCards []QuestionCard
	err = json.Unmarshal(secrets.AllQuestions, &AllQuestionCards)
	if err != nil {
		log.Printf("%v", err)
	}
	var AllAnswerCards []AnswerCard
	err = json.Unmarshal(secrets.AllAnswers, &AllAnswerCards)
	if err != nil {
		log.Printf("%v", err)
	}
	db, err := sql.Open("postgres", "sslmode=disable user=cahbot dbname=cahgames password="+secrets.DBPass)
	if err != nil {
		log.Fatal(err)
	}
	return &CAHBot{GenericBot, db, AllQuestionCards, AllAnswerCards}, err
}

// Question card
type QuestionCard struct {
	ID         int    `json:"id""`
	Text       string `json:"text"`
	NumAnswers int    `json:"numAnswers"`
	Expansion  string `json:"expansion"`
}

// Question card
type AnswerCard struct {
	ID        int    `json:"id""`
	Text      string `json:"text"`
	Expansion string `json:"expansion"`
}
