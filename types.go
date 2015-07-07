package main

import (
	"cahbot/tgbotapi"
)

// A wrapper for tgbotapi. We need this wrapper to add new methods.
type CAHBot struct {
	*tgbotapi.BotAPI
	CurrentGames map[int]CAHGame
	AllCards     map[int]map[string]string
}

// This creates a new CAHBot, which is basically a wrapper for tgbotapi.BotAPI.
// We need this wrapper to add the desired methods.
func NewCAHBot(token string) (*CAHBot, error) {
	GenericBot, err := tgbotapi.NewBotAPI(token)
	// Need to get the card data
	Cards := make(map[int]map[string]string)
	return &CAHBot{GenericBot, make(map[int]CAHGame), Cards}, err
}

// Struct that represents an instance of a game.
type CAHGame struct {
	ShuffledCards []int
	Players       map[int]PlayerGameInfo
	CardTzarOrder []int
	CardTzarID    int
	Settings      GameSettings
	HasStarted    bool
}

// Struct that represents a player in a game.
type PlayerGameInfo struct {
	Player          tgbotapi.User
	Points          int
	Cards           []int
	IsCardTzar      bool
	IsMysteryPlayer bool
	WaitingForCard  bool
}

// Settings for game.
type GameSettings struct {
	MysteryPlayer             bool
	TradeInTwoCardsEveryRound bool
	PickWorstToo              bool
	NumCardsInHand            int
}
