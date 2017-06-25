package main

import (
	"database/sql"
	"html"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/thedadams/telegram-bot-api"
)

// AnswerIsValid checks that the answer we received from the user is a valid answer.
func AnswerIsValid(bot *CAHBot, ChatID int64, Answer string) int {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return 0
	}
	var response string
	err = tx.QueryRow("SELECT get_user_cards($1)", ChatID).Scan(&response)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return 0
	}
	response = response[1 : len(response)-1]
	for _, val := range strings.Split(response, ",") {
		var tmp int
		tmp, _ = strconv.Atoi(val)
		if html.UnescapeString(bot.AllAnswerCards[tmp].Text) == Answer {
			return tmp
		}
	}
	return -1
}

// ArrayTransformForPostgres transforms an array for input into postgres database.
func ArrayTransformForPostgres(theArray []int) string {
	value := "{"
	for item := range theArray {
		value += strconv.Itoa(theArray[item]) + ","
	}
	value = value[0:len(value)-1] + "}"
	return value
}

// BuildScoreList builds the score list from a return sql.Rows.
func BuildScoreList(rows *sql.Rows) string {
	str := ""
	for rows.Next() {
		var response string
		if err := rows.Scan(&response); err == nil {
			arrResponse := strings.Split(response[1:len(response)-1], ",")
			str += strings.Replace(arrResponse[0], "\"", "", -1) + " had " + arrResponse[1] + " Awesome Points\n"
		} else {
			log.Printf("ERROR: %v", err)
			return "ERROR"
		}
	}
	return str
}

// CzarChoiceIsValid checks to see if we got a valid answer from the czar.
func CzarChoiceIsValid(bot *CAHBot, GameID, Answer string) int {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return 0
	}
	var answers string
	err = tx.QueryRow("SELECT get_answers($1)", GameID).Scan(&answers)
	tx.Commit()
	for _, val := range ShuffleAnswers(strings.Split(answers[1:len(answers)-1], "+=+\",")) {
		if Answer == strings.Replace(html.UnescapeString(strings.Replace(val[1:len(val)-1], "+=+", "", -1)), "\\\"", "", -1) {
			return 1
		}
	}
	return -1
}

// GameScores gets the scores for a game.
func GameScores(GameID string, db *sql.DB) string {
	rows, err := db.Query("SELECT get_player_scores($1)", GameID)
	defer rows.Close()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return "ERROR"
	}
	return BuildScoreList(rows)
}

// GetGameID gets the GameID for a player.
func GetGameID(UserID int, ChatID int64, db *sql.DB) (string, error) {
	var GameID string
	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return "", err
	}
	err = tx.QueryRow("SELECT get_game_id($1, $2)", UserID, ChatID).Scan(&GameID)
	if err != nil {
		return "", err
	}
	return GameID, err
}

// GetRandomID creates a random string for a Game ID.
func GetRandomID() string {
	id := ""
	characters := []string{"A", "B", "C", "D", "E", "F", "G", "H", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "!", "#", "$", "@", "?", "-", "&", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	n := len(characters)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < 5; i++ {
		id += characters[rand.Intn(n)]
	}
	return id
}

// HandleCzarResponse handles a response from the card czar.
func HandleCzarResponse(bot *CAHBot, GameID string, Message *tgbotapi.Message, Response string, CheckDigit int) {
	if CheckDigit == -1 {
		log.Printf("The text we received was not a valid answer.  We assume it was a message to the game so we are forwarding it.")
		bot.ForwardMessageToGame(Message, GameID)
	} else if CheckDigit == 0 {
		log.Printf("GameID: %v - We encountered an error when trying to validate the Card Czar's choice.  We are reporting that error to the Card Czar.", GameID)
		bot.SendActionFailedMessage(Message.Chat.ID)
		log.Printf("GameID: %v - Asking the Czar to try again...", GameID)
		bot.ListAnswers(GameID)
	} else {
		bot.CzarChoseAnswer(Message.Chat.ID, GameID, Message.Text, strings.Contains(Response, "best"))
	}
}

// HandlePlayerResponse handles a response that is not a command from a player.
func HandlePlayerResponse(bot *CAHBot, GameID string, Message *tgbotapi.Message, CheckDigit int, ThirdArg string, Handler func(int64, string, string)) {
	if CheckDigit == -1 {
		log.Printf("The text we received was not a valid answer.")
		bot.SendActionFailedMessage(Message.Chat.ID)
	} else if CheckDigit == 0 {
		log.Printf("GameID: %v - Asking the player with ID %v to try again...", GameID, Message.From.ID)
		message := tgbotapi.NewEditMessageText(Message.Chat.ID, Message.MessageID, "We encountered an error. Please try picking an answer again.")
		message.ReplyMarkup = SetupInlineKeyboard(bot.Settings, 1)
		bot.Send(message)
	} else {
		Handler(Message.Chat.ID, GameID, ThirdArg)
	}
}

// LastCharactorIsPunctuation checks to see if the last character of a string is punctuation.
func LastCharactorIsPunctuation(TheString string) bool {
	length := len(TheString) - 1
	if string(TheString[length]) == "." || string(TheString[length]) == "!" || string(TheString[length]) == "?" {
		return true
	}
	return false
}

// SetupInlineKeyboard builds the inline keyboard to change a setting.
func SetupInlineKeyboard(ButtonData []Setting, NumCols int) *tgbotapi.InlineKeyboardMarkup {
	// Settings that need to support changing: Mystery Player, Trade In Cards, Number of Cards to Trade In, Number of Cards In Hand, Pick Worst Also, Points To Win.
	settingIndex := 0
	numSettings := len(ButtonData)
	settingsKeyboard := make([][]tgbotapi.InlineKeyboardButton, (len(ButtonData)+NumCols-1)/NumCols)
	for i := 0; i < len(settingsKeyboard) && settingIndex < numSettings; i++ {
		settingsKeyboard[i] = make([]tgbotapi.InlineKeyboardButton, NumCols)
		for j := 0; j < NumCols && settingIndex < numSettings; j++ {
			settingsKeyboard[i][j] = tgbotapi.InlineKeyboardButton{Text: ButtonData[settingIndex].Name, CallbackData: &ButtonData[settingIndex].CData}
			settingIndex++
		}
	}
	return &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: settingsKeyboard}
}

// SettingIsValid checks to see if we received valid setting from the user.
func SettingIsValid(bot *CAHBot, Setting string) int {

	return 0
}

// ShuffleAnswers shuffles the answers so they don't come out in the same order every time.
func ShuffleAnswers(arr []string) []string {
	rand.Seed(time.Now().UnixNano())

	for i := len(arr) - 1; i > 0; i-- {
		j := rand.Intn(i)
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

// TrimPunctuation trims the punctuation on an answer to help the grammar.
func TrimPunctuation(TheString string) string {
	if !LastCharactorIsPunctuation(TheString) {
		return TheString
	}
	return TrimPunctuation(strings.TrimSuffix(strings.TrimSuffix(strings.TrimSuffix(TheString, "!"), "?"), "."))
}
