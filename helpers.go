package main

import (
	"math/rand"
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
	tmp += "Each player has " + strconv.Itoa(gs.NumCardsInHand) + "in their hand."
	return tmp + "\n\nUse command '/changesettings' to change these settings."
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
