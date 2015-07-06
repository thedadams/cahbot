package main

import (
	"math/rand"
	"time"
)

// Get the scores for a game.
func (g CAHGame) Scores() string {
	var str string = ""
	for key, value := range g.Players {
		str += value.Player.String() + " - " + string(value.Points)
	}
	return str
}

// Get the settings for a game.
func (gs GameSettings) String() string {
	var onOff string
	if gs.MysteryPlayer {
		onOff = "on"
	} else {
		onOff = "off"
	}
	tmp := "Mystery Player - " + onOff + "\n"
	if gs.TradeInTwoCardsEveryRound {
		onOff = "on"
	} else {
		onOff = "off"
	}
	tmp += "Trade in 2 cards every round " + onOff + "\n"
	if gs.PickWorstToo {
		onOff = "on"
	} else {
		onOff = "off"
	}
	tmp += "Pick the worst answer also " + onOff + "\n"
	tmp += "Each player has " + string(gs.NumCardsInHand) + "in their hand."
	return tmp
}

// Shuffle an array of ints.
func shuffle(arr []int) {
	t := time.Now()
	rand.Seed(int64(t.Nanosecond()))

	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
}
