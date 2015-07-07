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

// This function checks to see if we have answer cards from all the players.
func DoWeHaveAllAnswers(Players map[int]PlayerGameInfo) bool {
	for _, value := range Players {
		if !value.IsCardTzar && value.CardBeingPlayed == "" {
			return false
		}
	}
	return true
}

// This function will deal a player's hand or add cards to the player's hand
func DealPlayersHand(Game CAHGame, Hand []string) {
	for len(Hand) < Game.Settings.NumCardsInHand {
		append(Hand, Game.ShuffledAnswerCards[Game.NumACardsLeft])
		Game.NumACardsLeft -= 1
		if Game.NumACardsLeft == -1 {
			ReshuffleACards(Game)
		}
	}
}

//This function reshuffles the answer cards of a game.
func ReshuffleACards(Game CAHGame) {
	shuffle(Game.ShuffledAnswerCards)
	Game.NumACardsLeft = len(Game.ShuffledAnswerCards) - 1
}

//This function reshuffles the question cards of a game.
func ReshuffleQCards(Game CAHGame) {
	shuffle(Game.ShuffledQuestionCards)
	Game.NumQCardsLeft = len(Game.ShuffledQuestionCards) - 1
}
