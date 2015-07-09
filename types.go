package main

import (
	"cahbot/secrets"
	"cahbot/tgbotapi"
	"encoding/json"
	"log"
)

// A wrapper for tgbotapi. We need this wrapper to add new methods.
type CAHBot struct {
	*tgbotapi.BotAPI
	CurrentGames     map[int]CAHGame
	AllQuestionCards []QuestionCard
	AllAnswerCards   []AnswerCard
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
	return &CAHBot{GenericBot, make(map[int]CAHGame), AllQuestionCards, AllAnswerCards}, err
}

// Struct that represents an instance of a game.
type CAHGame struct {
	ChatID                int
	ShuffledQuestionCards []int
	ShuffledAnswerCards   []int
	NumQCardsLeft         int
	NumACardsLeft         int
	Players               map[int]PlayerGameInfo
	CardTzarOrder         []int
	CardTzarIndex         int
	QuestionCard          int
	Settings              GameSettings
	HasBegun              bool
	WaitingForAnswers     bool
}

// Struct that represents a player in a game.
// The ReplyID is to the join message the user sends so we can reply to it if they don't have a username.
type PlayerGameInfo struct {
	Player          tgbotapi.User
	ReplyID         int
	Points          int
	Cards           []int
	IsCardTzar      bool
	CardBeingPlayed int
}

// Settings for game.
type GameSettings struct {
	MysteryPlayer             bool
	TradeInTwoCardsEveryRound bool
	PickWorstToo              bool
	NumCardsInHand            int
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
