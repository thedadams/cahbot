package main

import (
	"tgbotapi"
)

type CAHGame struct {
	Bot           tgbotapi.BotAPI
	ShuffledCards []int
	Players       map[string]PlayerGameInfo
}

type PlayerGameInfo struct {
	Player     tgbotapi.User
	Points     int
	Cards      []int
	IsCardTzar bool
}
