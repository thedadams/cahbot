package main

import (
	"crypto/sha512"
	"encoding/base64"
	"html"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/thedadams/telegram-bot-api"
)

// HandleUpdate is the starting point for handling an update from chat.
func (bot *CAHBot) HandleUpdate(User *tgbotapi.User, Message *tgbotapi.Message, Callback *tgbotapi.CallbackQuery, messageType string) {
	bot.AddUserToDatabase(User, int64(User.ID))
	GameID, err := GetGameID(User.ID, int64(User.ID), bot.DBConn)
	log.Printf("Message from %s of type %s with game ID %s", User.String(), messageType, GameID)
	if messageType == "command" {
		bot.ProccessCommand(Message, GameID)
	} else if messageType == "callback" {
		log.Printf("We received a callback from %v.", User.ID)
		callbackType := strings.Split(Callback.Data, "::")
		switch callbackType[0] {
		case "ChangeSetting":
			// Handle the change of a setting here.
			HandlePlayerResponse(bot, GameID, Message, SettingIsValid(bot, Message.Text), Callback.Data, bot.ChangeGameSettings)
		case "Answer":
			// Handle the receipt of an answer here.
			answer := AnswerIsValid(bot, int64(User.ID), Message.Text)
			HandlePlayerResponse(bot, GameID, Message, answer, strconv.Itoa(answer), bot.ReceivedAnswerFromPlayer)
		case "TradeInCard":
			// Handle the trading in of a card here.
			answer := AnswerIsValid(bot, int64(User.ID), Message.Text)
			HandlePlayerResponse(bot, GameID, Message, answer, strconv.Itoa(answer), bot.TradeInCard)
		case "CzarBest":
			// Handle the receipt of a czar picking best answer here.
			HandleCzarResponse(bot, GameID, Message, callbackType[0], CzarChoiceIsValid(bot, GameID, Message.Text))
		case "CzarWorst":
			// Handle the receipt of a czar picking the worst answer here.
			HandleCzarResponse(bot, GameID, Message, callbackType[0], CzarChoiceIsValid(bot, GameID, Message.Text))
		}
	} else if messageType == "message" || messageType == "photo" || messageType == "video" || messageType == "audio" || messageType == "contact" || messageType == "document" || messageType == "location" || messageType == "sticker" {
		if err != nil {
			bot.Send(tgbotapi.NewMessage(int64(User.ID), "It seems that you are not involved in any game so your message fell on deaf ears."))
		} else {
			bot.ForwardMessageToGame(Message, GameID)
		}
	}
}

// SendToGame sends a message from a player to the rest of the group.
func (bot *CAHBot) SendToGame(GameID, message string) {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	rows, err := bot.DBConn.Query("SELECT get_user_ids_for_game($1)", GameID)
	defer rows.Close()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	var ID int64
	for rows.Next() {
		if err := rows.Scan(&ID); err != nil {
			log.Printf("ERROR: %v", err)
		} else {
			bot.Send(tgbotapi.NewMessage(ID, message))
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: %v", err)
	}
}

// ForwardMessageToGame forwards a message from a player to the rest of the group.
func (bot *CAHBot) ForwardMessageToGame(m *tgbotapi.Message, GameID string) {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(m.Chat.ID)
		return
	}
	rows, err := bot.DBConn.Query("SELECT get_chat_ids_for_game($1)", GameID)
	defer rows.Close()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(m.Chat.ID)
		return
	}
	var ID int64
	for rows.Next() {
		if err := rows.Scan(&ID); err != nil {
			log.Printf("ERROR: %v", err)
		} else if ID != m.Chat.ID {
			bot.Send(tgbotapi.NewForward(ID, m.Chat.ID, m.MessageID))
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: %v", err)
	}
}

// SendNoGameMessage sends a 'There is no game' message.
func (bot *CAHBot) SendNoGameMessage(ChatID int64) {
	bot.Send(tgbotapi.NewMessage(ChatID, "You are currently not in a game.  Use command /create to create a new one or /join <id> to join a game with an id."))
}

// WrongCommand sends a "wrong command" message.
func (bot *CAHBot) WrongCommand(ChatID int64) {
	bot.Send(tgbotapi.NewMessage(ChatID, "Sorry, I don't know that command."))
}

// SendActionFailedMessage sends a generic sorry message.
func (bot *CAHBot) SendActionFailedMessage(ChatID int64) {
	bot.Send(tgbotapi.NewMessage(ChatID, "I'm sorry, but it seems I have have difficulties right now.  You can try again later or contact my developer @thedadams."))
}

// DetectKindMessageReceived detects the kind of message we received from the user.
func (bot *CAHBot) DetectKindMessageReceived(u tgbotapi.Update) string {
	log.Printf("Detecting the type of message received")
	if u.CallbackQuery != nil {
		return "callback"
	}
	if u.Message.Text != "" {
		if u.Message.IsCommand() {
			return "command"
		}
		return "message"
	}
	if len(*u.Message.Photo) != 0 {
		return "photo"
	}
	if u.Message.Audio.FileID != "" {
		return "audio"
	}
	if u.Message.Video.FileID != "" {
		return "video"
	}
	if u.Message.Document.FileID != "" {
		return "document"
	}
	if u.Message.Sticker.FileID != "" {
		return "sticker"
	}
	if len(*u.Message.NewChatMembers) != 0 {
		return "newParticipant"
	}
	if u.Message.LeftChatMember.ID != 0 {
		return "byeParticipant"
	}
	if u.Message.NewChatTitle != "" {
		return "newChatTitle"
	}
	if len(*u.Message.NewChatPhoto) != 0 {
		return "newChatPhoto"
	}
	if u.Message.DeleteChatPhoto {
		return "deleteChatPhoto"
	}
	if u.Message.GroupChatCreated {
		return "newGroupChat"
	}
	if u.Message.Contact.UserID != 0 || u.Message.Contact.FirstName != "" || u.Message.Contact.LastName != "" {
		return "contact"
	}
	if u.Message.Location.Longitude != 0 && u.Message.Location.Latitude != 0 {
		return "location"
	}
	return "undetermined"
}

// ProccessCommand figures out which command the user invoked,
// and calls the appropriate method.
func (bot *CAHBot) ProccessCommand(m *tgbotapi.Message, GameID string) {
	log.Printf("Processing command....")
	// Get the command.
	switch strings.Replace(strings.Fields(m.Text)[0], "/", "", 1) {
	case "start":
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, "Welcome to Cards Against Humanity for Telegram.  To create a new game, use the command /create.  If you create a game, you will be given a 5 character id you can share with friends so they can join you.  You can also join a game using the /join <id> command where the <id> is replaced with a game id created by someone else.  To see all available commands, use /help."))
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, "While you are in a game, any (non-command) message you send to me will be automatically forwarded to everyone else in the game so you're all in the loop."))
	case "help":
		// TODO: use helpers to build a help message.
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, "A help message should go here."))
	case "create":
		if GameID != "" {
			bot.Send(tgbotapi.NewMessage(m.Chat.ID, "You are already part of a game with id "+GameID+" and cannot create another game.  You can leave your current game with the command /leave."))
		} else {
			ID := bot.CreateNewGame(m.Chat.ID, m.From)
			if ID != "" {
				bot.Send(tgbotapi.NewMessage(m.Chat.ID, "The game was created successfully.  Tell your friends to use the command '/join "+ID+"' to join your game.  Remember that your game will be deleted after 2 days of inactivity."))
				bot.AddPlayerToGame(ID, m.From, m.Chat.ID)
			} else {
				bot.Send(tgbotapi.NewMessage(m.Chat.ID, "An error occurred while trying to create the game.  The game was not created."))
			}
		}
	case "remove":
		tx, err := bot.DBConn.Begin()
		defer tx.Rollback()
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendActionFailedMessage(m.Chat.ID)
			return
		}
		// If the user is in a game, we remove them.
		if GameID != "" {
			bot.RemovePlayerFromGame(GameID, m.From, m.Chat.ID)
		}
		log.Printf("Removing user from the database.")
		_, err = tx.Exec("SELECT remove_user($1)", m.From.ID)
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendActionFailedMessage(m.Chat.ID)
			return
		}
		tx.Commit()
		bot.Send(tgbotapi.NewMessage(m.Chat.ID, "You have been removed from our records. If you ever want to come back, send the command /start.  Thank you for playing."))

	case "begin":
		if GameID != "" {
			bot.BeginGame(GameID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "end":
		if GameID != "" {
			bot.EndGame(GameID, m.From.String(), false)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "join":
		if len(strings.Fields(m.Text)) > 1 {
			if GameID != "" {
				bot.Send(tgbotapi.NewMessage(m.Chat.ID, "You are already part of a game with id "+GameID+" and cannot join another game.  You can leave your current game with the command /leave."))
			} else {
				// If the user is not part of another game, we check to see if they id they
				// game is valid.
				tx, err := bot.DBConn.Begin()
				defer tx.Rollback()
				if err != nil {
					log.Printf("ERROR: %v", err)
					bot.SendActionFailedMessage(m.Chat.ID)
					return
				}
				var exists bool
				row := tx.QueryRow("SELECT check_game_exists($1)", strings.Fields(m.Text)[1])
				if _ = row.Scan(&exists); exists {
					// If the game is in the middle of a round, we don't add them yet.
					var InRound bool
					err = tx.QueryRow("SELECT is_game_in_round($1)", GameID).Scan(&InRound)
					if err != nil || InRound {
						bot.Send(tgbotapi.NewMessage(m.Chat.ID, "The game you are trying to join is in the middle of a round.  Please wait until they are finished to join."))
					} else {
						// The id is valid and we add them and the game is not in-round.
						bot.AddPlayerToGame(strings.Fields(m.Text)[1], m.From, m.Chat.ID)
					}
				} else {
					bot.Send(tgbotapi.NewMessage(m.Chat.ID, "There is no game with id "+strings.Fields(m.Text)[1]+".  Please try again with a new id or use /create to create a game."))
				}
			}
		} else {
			bot.Send(tgbotapi.NewMessage(m.Chat.ID, "You did not enter a game id.  Try again with the format /join <id>."))
		}
	case "gameid":
		if GameID != "" {
			bot.Send(tgbotapi.NewMessage(m.Chat.ID, "The game you are currently playing has id "+GameID+".  Others can join your game by using the command '/join "+GameID+"'."))
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "leave":
		if GameID != "" {
			bot.RemovePlayerFromGame(GameID, m.From, m.Chat.ID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "next":
		if GameID != "" {
			bot.StartRound(GameID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "cards":
		if GameID != "" {
			bot.ListCardsForUserWithMessage(GameID, m.Chat.ID, "Your cards are listed in the keyboard area.")
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "scores":
		if GameID != "" {
			bot.Send(tgbotapi.NewMessage(m.Chat.ID, "Here are the current scores:\n"+GameScores(GameID, bot.DBConn)))
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "settings":
		if GameID != "" {
			bot.SendGameSettings(GameID, m.Chat.ID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "changesettings":
		if GameID != "" {
			tx, err := bot.DBConn.Begin()
			defer tx.Rollback()
			if err != nil {
				log.Printf("ERROR: %v", err)
				bot.SendActionFailedMessage(m.Chat.ID)
				return
			}
			var InRound bool
			err = tx.QueryRow("SELECT is_game_in_round($1)", GameID).Scan(&InRound)
			if err != nil || InRound {
				log.Printf("User attempting to change the settings for game with id %v in the middle of a round.", GameID)
				bot.Send(tgbotapi.NewMessage(m.Chat.ID, "You cannot change settings while the game is in the middle of a round.  Please wait until the round is finished and try again."))
				return
			}
			bot.SendGameSettings(GameID, m.Chat.ID)
			tx.Commit()
			message := tgbotapi.NewMessage(m.Chat.ID, "Which setting would you like to change?")
			message.ReplyMarkup = SetupInlineKeyboard(bot.Settings, 1)
			bot.Send(message)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "czar":
		if GameID != "" {
			var czar string
			tx, err := bot.DBConn.Begin()
			defer tx.Rollback()
			if err != nil {
				log.Printf("ERROR: %v", err)
				bot.SendActionFailedMessage(m.Chat.ID)
				return
			}
			err = tx.QueryRow("SELECT who_is_czar($1)", GameID).Scan(&czar)
			if err != nil {
				log.Printf("ERROR: %v", err)
				bot.SendActionFailedMessage(m.Chat.ID)
				return
			}
			bot.Send(tgbotapi.NewMessage(m.Chat.ID, "The current Card czar is "+czar+"."))
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "logging":
		if len(strings.Fields(m.Text)) > 1 {
			hasher := sha512.New()
			if strings.EqualFold(base64.URLEncoding.EncodeToString(hasher.Sum([]byte(strings.Fields(m.Text)[1]))), os.Getenv("APPPASS")) {
				bot.Debug = !bot.Debug
				log.Printf("Debugging/verbose logging has been turned to %v.", bot.Debug)
			}
		} else {
			bot.WrongCommand(m.Chat.ID)
		}
	default:
		bot.WrongCommand(m.Chat.ID)
	}
}

// AddPlayerToGame adds a player to a game if the player is not playing.
func (bot *CAHBot) AddPlayerToGame(GameID string, User *tgbotapi.User, ChatID int64) {
	// This is supposed to check that there are not more than 10 players in a game.
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	var numPlayersInGame int
	err = tx.QueryRow("SELECT num_players_in_game($1)", GameID).Scan(&numPlayersInGame)
	if numPlayersInGame > 10 {
		bot.Send(tgbotapi.NewMessage(ChatID, "Player limit of 10 reached, we can not add any more players."))
	} else {
		var tmp bool
		row := tx.QueryRow("SELECT is_player_in_game($2,$1)", GameID, ChatID)
		if _ = row.Scan(&tmp); tmp {
			bot.Send(tgbotapi.NewMessage(ChatID, "You are already playing in this game.  Use command /leave to remove yourself."))
		} else {
			log.Printf("Adding %v to the game %v...", User, GameID)
			_, err = tx.Exec("SELECT add_player_to_game($1, $2)", GameID, ChatID)
			if err != nil {
				log.Printf("ERROR: %v", err)
				bot.SendActionFailedMessage(ChatID)
				return
			}
			if tx.Commit() != nil {
				log.Printf("ERROR %T: %v", err, err)
				bot.SendActionFailedMessage(ChatID)
				return
			}
			bot.SendToGame(GameID, User.String()+" has joined the game!")
			bot.Send(tgbotapi.NewMessage(ChatID, "Welcome to the game!  Here are the currect game settings for your review."))
			bot.SendGameSettings(GameID, ChatID)
		}
	}
}

// AddUserToDatabase adds a user to the database. It does not link them to a game.
func (bot *CAHBot) AddUserToDatabase(User *tgbotapi.User, ChatID int64) bool {
	// Check to see if the user is already in the database.
	var exists bool
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("Cannot connect to the database.")
		return false
	}
	_ = tx.QueryRow("SELECT does_user_exist($1::int)", User.ID).Scan(&exists)
	if !exists {
		log.Printf("Adding user with ID %v to the database.", User.ID)
		_, err = tx.Exec("SELECT add_user($1, $2, $3, $4, $5, $6)", User.ID, ChatID, User.FirstName, User.LastName, User.UserName, User.String())
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendActionFailedMessage(ChatID)
			return false
		}
		tx.Commit()
		return true
	}
	log.Printf("User with id %v is already in the database.", User.ID)
	return true
}

// BeginGame begins an already created game.
func (bot *CAHBot) BeginGame(GameID string) {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendToGame(GameID, "We could not start the game because of an internal error.")
		return
	}
	// Check to see if there are more than 2 players.
	var tmp int
	err = tx.QueryRow("SELECT num_players_in_game($1)", GameID).Scan(&tmp)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendToGame(GameID, "We could not start the game because of an internal error.")
		return
	}
	if tmp < 2 {
		log.Printf("There aren't enough players in game with id " + GameID + " to start it.")
		tx.Rollback()
		bot.SendToGame(GameID, "You really need at least 3 players to make it interesting.  Right now, you have "+strconv.Itoa(tmp)+".  Tell others to use the command '/join "+GameID+"' to join your game.")
		return
	}
	log.Printf("Trying to start game with id %v.", GameID)
	tx.Commit()
	bot.SendToGame(GameID, "Get ready, we are starting the game!")
	bot.StartRound(GameID)
}

// ChangeGameSettings changes a setting for the given game.
// TODO: implement.
func (bot *CAHBot) ChangeGameSettings(ChatID int64, GameID string, Setting string) {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("GameID: %v - ERROR: %v", GameID, err)
		bot.SendToGame(GameID, "We could not start the game because of an internal error.")
		return
	}
}

// CreateNewGame creates a new game.
func (bot *CAHBot) CreateNewGame(ChatID int64, User *tgbotapi.User) string {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return ""
	}
	var GameID string
	for {
		GameID = GetRandomID()
		var tmp bool
		_ = tx.QueryRow("SELECT check_game_exists($1)", GameID).Scan(&tmp)
		if !tmp {
			break
		}
	}
	log.Printf("Creating a new game with ID %v.", GameID)
	// Get the keys for the All Cards map.SE
	ShuffledQuestionCards := make([]int, len(bot.AllQuestionCards))
	for i := 0; i < len(ShuffledQuestionCards); i++ {
		ShuffledQuestionCards[i] = i
	}
	ShuffledAnswerCards := make([]int, len(bot.AllAnswerCards))
	for i := 0; i < len(ShuffledAnswerCards); i++ {
		ShuffledAnswerCards[i] = i
	}
	if err != nil {
		log.Printf("Error creating game: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return ""
	}
	tx.Exec("SELECT add_game($1,$2,$3,$4)", GameID, ArrayTransformForPostgres(ShuffledQuestionCards), ArrayTransformForPostgres(ShuffledAnswerCards), User.ID)
	err = tx.Commit()
	if err != nil {
		log.Printf("Game could not be created. ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return ""
	}
	log.Printf("Game with id %v created successfully!", GameID)
	return GameID
}

// CzarChoseAnswer handles the czar choosing an answer.
func (bot *CAHBot) CzarChoseAnswer(ChatID int64, GameID string, Answer string, BestAnswer bool) {
	log.Printf("The Card Czar for game with id %v chose a valid answer.", GameID)
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	var response string
	err = tx.QueryRow("SELECT czar_chose_answer($1,$2)", GameID, Answer).Scan(&response)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	response = response[1 : len(response)-1]
	winner := strings.Split(response, ",")[0]
	gameOver := strings.ToUpper(strings.Split(response, ",")[1]) == "TRUE"
	if BestAnswer {
		bot.SendToGame(GameID, "The czar chose the best answer: "+Answer+"\n\nThis was "+winner+"'s answer.  You get one Awesome Point!")
	} else {
		bot.SendToGame(GameID, "The czar chose the worst answer: "+Answer+"\n\nThis was "+winner+"'s answer.  You lose one Awesome Point.")
	}
	if gameOver {
		tx.Commit()
		bot.EndGame(GameID, "", true)
	} else {
		tx.Exec("SELECT end_round($1)", GameID)
		err = tx.QueryRow("SELECT who_is_czar($1)", GameID).Scan(&winner)
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendActionFailedMessage(ChatID)
			return
		}
		var czarChatID int64
		err = tx.QueryRow("SELECT czar_chat_id($1, $2)", GameID, "").Scan(&czarChatID)
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendActionFailedMessage(ChatID)
			return
		}
		tx.Commit()
		bot.SendToGame(GameID, "The new Card Czar is "+winner+".  They will start the new round soon.")
		bot.Send(tgbotapi.NewMessage(czarChatID, "You are the Card Czar for the next round.  Use the command /next to start the next round."))
	}
}

// DisplayQuestionCard sends a message show the players the question card.
func (bot *CAHBot) DisplayQuestionCard(GameID string, AddCardsToPlayersHands bool) {
	log.Printf("Getting question card index for game with id %v", GameID)
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	var index int
	err = tx.QueryRow("SELECT get_question_card($1)", GameID).Scan(&index)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	log.Printf("The current question cards for game with id %v has index %v.", GameID, index)
	if AddCardsToPlayersHands && bot.AllQuestionCards[index].NumAnswers > 1 {
		_, err = tx.Exec("SELECT add_cards_to_all_in_game($1, $2)", GameID, bot.AllQuestionCards[index].NumAnswers-1)
		if err != nil {
			log.Printf("ERROR: %v", err)
			return
		}
	}
	tx.Commit()
	log.Printf("Sending question card to game with ID %v...", GameID)
	message := "Here is the question card:\n\n"
	message += strings.Replace(html.UnescapeString(bot.AllQuestionCards[index].Text), "\\\"", "", -1)
	bot.SendToGame(GameID, html.UnescapeString(message))
}

// EndGame stops and ends an already created game.
func (bot *CAHBot) EndGame(GameID string, UserThatStoppedGame string, SomeoneWon bool) {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendToGame(GameID, "There was an error when I tried to end the game.  You can try again or contact my developer @thedadams.")
		return
	}
	log.Printf("Deleting a game with id %v...", GameID)
	rows, err := tx.Query("SELECT end_game($1)", GameID)
	defer rows.Close()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendToGame(GameID, "There was an error when I tried to end the game.  You can try again or contact my developer @thedadams.")
		return
	}
	if !SomeoneWon {
		// Someone ended the game.
		bot.SendToGame(GameID, "The game has been stopped by "+UserThatStoppedGame+".\nHere are the scores:\n"+BuildScoreList(rows)+"Thanks for playing!")
	} else {
		// The game ended because someone won.
		bot.SendToGame(GameID, "The game has ended.  Here are the scores:\n"+BuildScoreList(rows)+"Thanks for playing!")
	}
	tx.Commit()
}

// ListAnswers lists the answers for everyone and allows the czar to choose one.
func (bot *CAHBot) ListAnswers(GameID string) {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	var cards string
	err = tx.QueryRow("SELECT get_answers($1)", GameID).Scan(&cards)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	text := "Here are the submitted answers:\n\n"
	cardsKeyboard := make([][]tgbotapi.KeyboardButton, 1)
	for i, val := range ShuffleAnswers(strings.Split(cards[1:len(cards)-1], "+=+\",")) {
		text += strings.Replace(html.UnescapeString(strings.Replace(val[1:len(val)-1], "+=+", "", -1)), "\\\"", "", -1) + "\n"
		cardsKeyboard[i] = make([]tgbotapi.KeyboardButton, 1)
		cardsKeyboard[i][0] = tgbotapi.KeyboardButton{Text: strings.Replace(html.UnescapeString(strings.Replace(val[1:len(val)-1], "+=+", "", -1)), "\\\"", "", -1), RequestContact: false, RequestLocation: false}
	}
	log.Printf("Showing everyone the answers submitted for game %v.", GameID)
	bot.SendToGame(GameID, text)
	var czarChatID int64
	err = tx.QueryRow("SELECT czar_chat_id($1, $2)", GameID, "czarbest").Scan(&czarChatID)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	tx.Commit()
	log.Printf("Asking the czar, %v, to pick an answer for game with id %v.", czarChatID, GameID)
	message := tgbotapi.NewMessage(czarChatID, "Czar, please choose the best answer.")
	message.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{Keyboard: cardsKeyboard, ResizeKeyboard: true, OneTimeKeyboard: true, Selective: false}
	bot.Send(message)
}

// ListCardsForUserWithMessage lists a user's cards using a custom keyboard in the Telegram API.  If we need them to respond to a question, this is handled.
func (bot *CAHBot) ListCardsForUserWithMessage(GameID string, ChatID int64, text string) {
	log.Printf("Showing the user %v their cards.", ChatID)
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	var response string
	err = tx.QueryRow("SELECT get_user_cards($1)", ChatID).Scan(&response)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	response = response[1 : len(response)-1]
	message := tgbotapi.NewMessage(ChatID, text)
	cards := make([][]tgbotapi.InlineKeyboardButton, len(strings.Split(response, ",")))
	for i := range cards {
		cards[i] = make([]tgbotapi.InlineKeyboardButton, 1)
	}
	for i := 0; i < len(strings.Split(response, ",")); i++ {
		tmp, _ := strconv.Atoi(strings.Split(response, ",")[i])
		answerText := html.UnescapeString(bot.AllAnswerCards[tmp].Text)
		callbackData := "answer::" + answerText
		cards[i][0] = tgbotapi.InlineKeyboardButton{Text: answerText, CallbackData: &callbackData}
	}
	message.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{InlineKeyboard: cards}
	bot.Send(message)
}

// ReceivedAnswerFromPlayer handles the receipt of an answer from a player.
func (bot *CAHBot) ReceivedAnswerFromPlayer(ChatID int64, GameID string, Answer string) {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	AnswerIndex, _ := strconv.Atoi(Answer)
	var QuestionIndex int
	var DisplayName string
	var CurrentAnswer string
	err = tx.QueryRow("SELECT get_question_card($1)", GameID).Scan(&QuestionIndex)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	err = tx.QueryRow("SELECT get_current_answer($1)", ChatID).Scan(&CurrentAnswer)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	if CurrentAnswer == "" {
		CurrentAnswer = bot.AllQuestionCards[QuestionIndex].Text
		// If the question is really a question without any blanks, we just list the answer.
		if !strings.Contains(CurrentAnswer, "_") {
			CurrentAnswer = bot.AllAnswerCards[AnswerIndex].Text
		}
	}
	CurrentAnswer = strings.Replace(CurrentAnswer, "_", TrimPunctuation(bot.AllAnswerCards[AnswerIndex].Text), 1)
	_, err = tx.Exec("SELECT received_answer_from_user($1, $2, $3, $4)", ChatID, AnswerIndex, CurrentAnswer, !strings.Contains(CurrentAnswer, "_"))
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	err = tx.QueryRow("SELECT get_display_name($1)", ChatID).Scan(&DisplayName)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	if strings.Contains(CurrentAnswer, "_") {
		log.Printf("We received a valid answer from user with id %v, but we need another answer.", ChatID)
		tx.Commit()
		bot.ListCardsForUserWithMessage(GameID, ChatID, "We received your answer, but this is a multi-answer questions.  Please choose another answer.")
	} else {
		log.Printf("We received a valid, complete answer from user with id %v.", ChatID)
		bot.SendToGame(GameID, "We received "+DisplayName+"'s answer.")
		err = tx.QueryRow("SELECT do_we_have_all_answers($1)", GameID).Scan(&QuestionIndex)
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendActionFailedMessage(ChatID)
			return
		}
		tx.Commit()
		if QuestionIndex == 1 {
			go bot.ListAnswers(GameID)
		}
	}
}

// RemovePlayerFromGame removes a player from a game if the player is playing.
func (bot *CAHBot) RemovePlayerFromGame(GameID string, User *tgbotapi.User, ChatID int64) {
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	log.Printf("Removing %v from the game %v...", User, GameID)
	str := ""
	err = tx.QueryRow("SELECT remove_player_from_game($1)", ChatID).Scan(&str)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return
	}
	bot.Send(tgbotapi.NewMessage(ChatID, "Thanks for playing, "+User.String()+"!  You collected "+strings.Split(str[1:len(str)-1], ",")[1]+" Awesome Points."))
	// Now check to see if there is anyone still in the game.
	var numPlayersInGame int
	err = tx.QueryRow("SELECT num_players_in_game($1)", GameID).Scan(&numPlayersInGame)
	if err != nil {
		log.Printf("ERROR: %v", err)
		tx.Commit()
		return
	}
	tx.Commit()
	if numPlayersInGame == 0 {
		log.Printf("There are no more players in game with id %v.  We shall end it.", GameID)
		bot.EndGame(GameID, User.String(), false)
	} else {
		bot.SendToGame(GameID, User.String()+" has left the game with a score of "+strings.Split(str[1:len(str)-1], ",")[1]+".")
	}
}

// SendGameSettings sends the game settings to the person that requested them.
func (bot *CAHBot) SendGameSettings(GameID string, ChatID int64) {
	tx, err := bot.DBConn.Begin()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		tx.Rollback()
		return
	}
	var settings string
	err = tx.QueryRow("SELECT game_settings($1)", GameID).Scan(&settings)
	tx.Commit()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	text := "Game settings:\n"
	settings = strings.Replace(settings, "false", "No", -1)
	settings = strings.Replace(settings, "true", "Yes", -1)
	for _, val := range strings.Split(settings[1:len(settings)-1], ",") {
		text += val[1:len(val)-1] + "\n"
	}
	log.Printf("Sending game settings for %v.", GameID)
	bot.Send(tgbotapi.NewMessage(ChatID, text))
}

// StartRound handles the starting/resuming of a round.
func (bot *CAHBot) StartRound(GameID string) {
	log.Printf("Attempting to start the next round for game with id %v.", GameID)
	ids := make([]int64, 0)
	tx, err := bot.DBConn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendToGame(GameID, "We ran into an error and cannot start the next round.  You can report the error to my developer, @thedadams, or try again later.")
		return
	}
	// Check to see if the game is running and if we are waiting for answers.
	var waiting bool
	err = tx.QueryRow("SELECT waiting_for_answers($1)", GameID).Scan(&waiting)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendToGame(GameID, "We ran into an error and cannot start the next round.  You can report the error to my developer, @thedadams, or try again later.")
		return
	}
	if waiting {
		bot.SendToGame(GameID, "We are waiting for players to give answers.")
	} else {
		rows, err := tx.Query("SELECT start_round($1)", GameID)
		defer rows.Close()
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendToGame(GameID, "We ran into an error and cannot start the next round.  You can report the error to my developer, @thedadams, or try again later.")
			return
		}
		for rows.Next() {
			var tmp int64
			err = rows.Scan(&tmp)
			if err != nil {
				log.Printf("ERROR: %v", err)
			}
			ids = append(ids, tmp)
		}
		tx.Commit()
		if ids[0] == -1 {
			log.Printf("We cannot start the next round for game with id %v because someone is changing the settings.", GameID)
			var tmp string
			tx, err = bot.DBConn.Begin()
			err = tx.QueryRow("SELECT get_display_name($1)", ids[1]).Scan(&tmp)
			tx.Rollback()
			if err != nil {
				log.Printf("ERROR: %v", err)
				bot.SendToGame(GameID, "We cannot begin the next round because someone is changing the settings of the game.")
				return
			}
			bot.SendToGame(GameID, "We cannot start the next round because "+tmp+" is changing the settings of the game.")
			return
		}
		bot.DisplayQuestionCard(GameID, true)
		for i := range ids {
			log.Printf("Asking %v for an answer card.", ids[i])
			bot.ListCardsForUserWithMessage(GameID, ids[i], "Please pick an answer for the question.")
		}
	}
}

// TradeInCard handles the trading in of a card at the end of the round.
func (bot *CAHBot) TradeInCard(ChatID int64, GameID string, Answer string) {

}
