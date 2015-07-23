package main

import (
	"database/sql"
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

// This builds the score list from a return sql.Rows.
func BuildScoreList(rows *sql.Rows) string {
	var str string = ""
	for rows.Next() {
		var response string
		if err := rows.Scan(&response); err == nil {
			arrResponse := strings.Split(response[1:len(response)-1], ",")
			str += arrResponse[0] + " hand " + arrResponse[1] + " Awesome Points\n"
		} else {
			log.Printf("ERROR: %v", err)
			return "ERROR"
		}
	}
	return str
}

// Get the scores for a game.
func GameScores(GameID string, db *sql.DB) string {
	rows, err := db.Query("SELECT get_player_scores($1)", GameID)
	defer rows.Close()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return "ERROR"
	}
	return BuildScoreList(rows)
}

// This function gets the GameID for a player.
func GetGameID(UserID int, db *sql.DB) (string, string, error) {
	var GameID string
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return "", "", err
	}
	err = tx.QueryRow("SELECT get_gameid($1)", UserID).Scan(&GameID)
	GameID = GameID[1 : len(GameID)-1]
	return strings.Split(GameID, ",")[0], strings.Split(GameID, ",")[1], err

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

// This function shuffles the answers so they don't come out in the same order every time.
func ShuffleAnswers(arr []string) []string {
	rand.Seed(time.Now().UnixNano())

	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}
