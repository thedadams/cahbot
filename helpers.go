package cahbot

import (
	"math/rand"
	"time"
)

// Get the scores for a game.
func (g *CAHGame) Scores() string {
	var str string = ""
	for key, value := range g.Players {
		str += value[Player].String() + " - " + string(value[Points])
	}
	return str
}

// Get the settings for a game.
func (gs GameSettings) String() string {
	tmp := "Mystery Player - " + If(gs.MysteryPlayer, "on", "off") + "\n"
	tmp += "Trade in 2 cards every round " + If(gs.TradeInTwoCardsEveryRound, "on", "off") + "\n"
	tmp += "Pick the worst answer also " + If(gs.PickWorstToo, "on", "off") + "\n"
	tmp += "Each player has " + gs.NumCardsInHand + "in their hand."
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
