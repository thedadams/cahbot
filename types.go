package main

import (
	"cahbot/secrets"
	"cahbot/tgbotapi"
	"encoding/json"
)

// A wrapper for tgbotapi. We need this wrapper to add new methods.
type CAHBot struct {
	*tgbotapi.BotAPI
	CurrentGames     map[int]CAHGame
	AllQuestionCards map[string]interface{}
	AllAnswerCards   map[string]interface{}
}

// This creates a new CAHBot, which is basically a wrapper for tgbotapi.BotAPI.
// We need this wrapper to add the desired methods.
func NewCAHBot(token string) (*CAHBot, error) {
	GenericBot, err := tgbotapi.NewBotAPI(token)
	// Need to get the card data
	var AllQuestionCards map[string]interface{}
	_ = json.Unmarshal(secrets.AllQuestions, &AllQuestionCards)
	var AllAnswerCards map[string]interface{}
	_ = json.Unmarshal(secrets.AllAnswers, &AllAnswerCards)
	return &CAHBot{GenericBot, make(map[int]CAHGame), AllQuestionCards, AllAnswerCards}, err
}

// Struct that represents an instance of a game.
type CAHGame struct {
	ShuffledQuestionCards []string
	ShuffledAnswerCards   []string
	NumQCardsLeft         int
	NumACardsLeft         int
	Players               map[int]PlayerGameInfo
	CardTzarOrder         []int
	CardTzarIndex         int
	QuestionCard          int
	Settings              GameSettings
	HasStarted            bool
}

// Struct that represents a player in a game.
type PlayerGameInfo struct {
	Player          tgbotapi.User
	Points          int
	Cards           []string
	IsCardTzar      bool
	CardBeingPlayed string
}

// Settings for game.
type GameSettings struct {
	MysteryPlayer             bool
	TradeInTwoCardsEveryRound bool
	PickWorstToo              bool
	NumCardsInHand            int
}
