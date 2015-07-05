package main

import (
	"tgbotapi"
)

// A wrapper for tgbotapi. We need this wrapper to add new methods.
type CAHBot struct {
    *tgbotapi.BotAPI
}

// Struct that represents an instance of a game.
type CAHGame struct {
	ShuffledCards []int
	Players       map[string]PlayerGameInfo
	Settings      GameSettings
}

// Struct that represents a player in a game.
type PlayerGameInfo struct {
	Player          tgbotapi.User
	Points          int
	Cards           []int
	IsCardTzar      bool
	IsMysteryPlayer bool
}

// Settings for game.
type GameSettings struct {
	MysteryPlayer             bool
	TradeInTwoCardsEveryRound bool
	PickWorstToo              bool
	NumCardsInHand            int
}
