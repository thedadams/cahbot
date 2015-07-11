package main

import (
	"html"
	"log"
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
	if gs.TradeInCardsEveryRound {
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
func shuffle(arr []int) {
	rand.Seed(time.Now().UnixNano())

	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
}

// This function shuffles the answers so they don't come out in the same order every time.
func ShuffleAnswers(arr [][]string) {
	rand.Seed(time.Now().UnixNano())

	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
}

// This function checks to see if we have answer cards from all the players.
func DoWeHaveAllAnswers(Players map[string]PlayerGameInfo) bool {
	for _, value := range Players {
		if !value.IsCardTzar && value.AnswerBeingPlayed == "" {
			return false
		}
	}
	return true
}

// This function will deal a player's hand or add cards to the player's hand
func DealPlayerHand(Game CAHGame, Hand []int) []int {
	for len(Hand) < Game.Settings.NumCardsInHand {
		log.Printf("Dealing card %v to user.", Game.NumACardsLeft)
		Hand = append(Hand, Game.ShuffledAnswerCards[Game.NumACardsLeft])
		Game.NumACardsLeft -= 1
		if Game.NumACardsLeft == -1 {
			ReshuffleACards(Game)
		}
	}
	return Hand
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

// This function goes through a game and determines if someone has won.
func DidSomeoneWin(Game CAHGame) (PlayerGameInfo, bool) {
	for _, value := range Game.Players {
		if value.Points == Game.Settings.NumPointsToWin {
			return value, true
		}
	}
	return PlayerGameInfo{}, false
}

// This function builds the list of submitted answers to be available for the players.
func BuildAnswerList(Game CAHGame) [][]string {
	answers := make([][]string, len(Game.Players))
	for i := range answers {
		answers[i] = make([]string, 1)
	}
	i := 0
	for _, value := range Game.Players {
		answers[i][0] = html.UnescapeString(value.AnswerBeingPlayed)
		i++
	}
	ShuffleAnswers(answers)
	return answers
}
