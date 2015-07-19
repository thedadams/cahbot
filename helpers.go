package main

import (
	"database/sql"
	"html"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// Transforms an array for input into postges database.
func ArrayTransforForPostgres(theArray []int) string {
	value := "{"
	for item := range theArray {
		value += strconv.Itoa(theArray[item]) + ","
	}
	value = value[0:len(value)-1] + "}"
	return value
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

// This builds the score list from a return sql.Rows.
func BuildScoreList(rows *sql.Rows) string {
	var str string = ""
	for rows.Next() {
		var response string
		if err := rows.Scan(&response); err == nil {
			log.Print(response)
			arrResponse := strings.Split(response[1:len(response)-1], ",")
			str += arrResponse[0] + " - " + arrResponse[1] + "\n"
		} else {
			log.Printf("ERROR: %v", err)
			return "ERROR"
		}
	}
	return str
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

// This function goes through a game and determines if someone has won.
func DidSomeoneWin(Game CAHGame) (PlayerGameInfo, bool) {
	for _, value := range Game.Players {
		if value.Points == Game.Settings.NumPointsToWin {
			return value, true
		}
	}
	return PlayerGameInfo{}, false
}

// Get the scores for a game.
func GameScores(GameID string, db *sql.DB) string {
	// Write a stored procedure for this query
	rows, err := db.Query("SELECT get_player_scores($1)", GameID)
	defer rows.Close()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return "ERROR"
	}
	return BuildScoreList(rows)
}

// This function gets the GameID for a player.
func GetGameID(UserID int, db *sql.DB) (string, error) {
	var GameID string
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return "", err
	}
	err = tx.QueryRow("SELECT get_gameid($1)", UserID).Scan(&GameID)
	return GameID, err

}

// Creates a random string for a Game ID.
func GetRandomID() string {
	var id string = ""
	characters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "!", "#", "$", "@", "?", "-", "&", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	n := len(characters)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 5; i++ {
		id += characters[rand.Intn(n)]
	}
	return id
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
