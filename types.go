package main

import (
	"tgbotapi"
)

type CAHBot struct {
    *tgbotapi.BotAPI
}

type CAHGame struct {
	ShuffledCards []int
	Players       map[string]PlayerGameInfo
	Settings      GameSettings
}

type PlayerGameInfo struct {
	Player          tgbotapi.User
	Points          int
	Cards           []int
	IsCardTzar      bool
	IsMysteryPlayer bool
}

type GameSettings struct {
	MysteryPlayer             bool
	TradeInTwoCardsEveryRound bool
	PickWorstToo              bool
	NumCardsInHand            int
}
