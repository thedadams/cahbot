package main

import (
	"math/rand"
	"strconv"
	"time"
)

// Get the scores for a game.
func (g CAHGame) Scores() string {
	var str string = ""
	for _, value := range g.Players {
		str += value.Player.String() + " - " + strconv.Itoa(value.Points) + "\n"
	}
	return str
}

// Get the settings for a game.
func (gs GameSettings) String() string {
	var onOff string
	if gs.MysteryPlayer {
		onOff = "On"
	} else {
		onOff = "Off"
	}
	tmp := "Mystery Player - " + onOff + "\n"
	if gs.TradeInTwoCardsEveryRound {
		onOff = "On"
	} else {
		onOff = "Off"
	}
	tmp += "Trade in 2 cards every round - " + onOff + "\n"
	if gs.PickWorstToo {
		onOff = "On"
	} else {
		onOff = "Off"
	}
	tmp += "Pick the worst answer also " + onOff + "\n"
	tmp += "Each player has " + strconv.Itoa(gs.NumCardsInHand) + " cards in their hand."
	return tmp + "\n\nUse command '/changesettings' to change these settings."
}

// Shuffle an array of strings.
func shuffle(arr []string) {
	t := time.Now()
	rand.Seed(int64(t.Nanosecond()))

	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
}

// This method checks to see if we have answer cards from all the players.
func DoWeHaveAllAnswers(Players map[int]PlayerGameInfo) bool {
	for _, value := range Players {
		if !value.IsCardTzar && value.WaitingForCard {
			return false
		}
	}
	return true
}
