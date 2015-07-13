package main

import (
	"cahbot/secrets"
	"cahbot/tgbotapi"
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"log"
)

// A wrapper for tgbotapi. We need this wrapper to add new methods.
type CAHBot struct {
	*tgbotapi.BotAPI
	db_conn          *sql.DB
	CurrentGames     map[string]CAHGame `json:"current_games"`
	AllQuestionCards []QuestionCard     `json:"all_question_cards"`
	AllAnswerCards   []AnswerCard       `json:"all_answer_cards"`
}

// This creates a new CAHBot, which is basically a wrapper for tgbotapi.BotAPI.
// We need this wrapper to add the desired methods.
func NewCAHBot(token string) (*CAHBot, error) {
	GenericBot, err := tgbotapi.NewBotAPI(token)
	// Need to get the card data
	var AllQuestionCards []QuestionCard
	err = json.Unmarshal(secrets.AllQuestions, &AllQuestionCards)
	if err != nil {
		log.Printf("%v", err)
	}
	var AllAnswerCards []AnswerCard
	err = json.Unmarshal(secrets.AllAnswers, &AllAnswerCards)
	if err != nil {
		log.Printf("%v", err)
	}
	db, err := sql.Open("postgres", "sslmode=disable user=cahbot dbname=cahgames password="+secrets.DBPass)
	if err != nil {
		log.Fatal(err)
	}
	return &CAHBot{GenericBot, db, make(map[string]CAHGame), AllQuestionCards, AllAnswerCards}, err
}

// Struct that represents an instance of a game.
type CAHGame struct {
	ChatID                int                       `json:"chat_id"`
	ShuffledQuestionCards []int                     `json:"shuffled_q_cards"`
	ShuffledAnswerCards   []int                     `json:"shuffled_a_cards"`
	NumQCardsLeft         int                       `json:"q_cards_left"`
	NumACardsLeft         int                       `json:"a_cards_left"`
	Players               map[string]PlayerGameInfo `json:"players"`
	CardTzarOrder         []string                  `json:"tzar_order"`
	CardTzarIndex         int                       `json:"tzar_index"`
	QuestionCard          int                       `json:"current_q_card"`
	Settings              GameSettings              `json:"settings"`
	HasBegun              bool                      `json:"has_begun"`
	WaitingForAnswers     bool                      `json:"waiting_for_answers"`
}

// Struct that represents a player in a game.
// The ReplyID is to the join message the user sends so we can reply to it if they don't have a username.
type PlayerGameInfo struct {
	Player            tgbotapi.User `json:"user"`
	ReplyID           int           `json:"reply_id"`
	Points            int           `json:"points"`
	Cards             []int         `json:"cards"`
	IsCardTzar        bool          `json:"is_tzar"`
	AnswerBeingPlayed string        `json:"answer_played"`
}

// Settings for game.
type GameSettings struct {
	MysteryPlayer          bool `json:"mystery_player"`
	TradeInCardsEveryRound bool `json:"trade_in_cards_every_round"`
	NumCardsToTradeIn      int  `json:"num_cards_to_trade_in"`
	PickWorstToo           bool `json:"pick_worst"`
	NumCardsInHand         int  `json:"num_cards_in_hand"`
	NumPointsToWin         int  `json:"points_to_win"`
}

// Question card
type QuestionCard struct {
	ID         int    `json:"id""`
	Text       string `json:"text"`
	NumAnswers int    `json:"numAnswers"`
	Expansion  string `json:"expansion"`
}

// Question card
type AnswerCard struct {
	ID        int    `json:"id""`
	Text      string `json:"text"`
	Expansion string `json:"expansion"`
}
