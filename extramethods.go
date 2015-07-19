package main

import (
	"cahbot/secrets"
	"cahbot/tgbotapi"
	"crypto/sha512"
	_ "database/sql"
	"encoding/base64"
	"html"
	"log"
	"strconv"
	"strings"
)

// This is the starting point for handling an update from chat.
func (bot *CAHBot) HandleUpdate(update *tgbotapi.Update) {
	bot.AddUserToDatabase(update.Message.From, update.Message.Chat.ID)
	GameID, err := GetGameID(update.Message.From.ID, bot.db_conn)
	messageType := bot.DetectKindMessageRecieved(&update.Message)
	log.Printf("[%s] Message type: %s", update.Message.From.UserName, messageType)
	if messageType == "command" {
		bot.ProccessCommand(&update.Message, GameID)
	} else if messageType == "message" || messageType == "photo" || messageType == "video" || messageType == "audio" || messageType == "contact" || messageType == "document" || messageType == "location" || messageType == "sticker" {
		if err != nil {
			bot.SendMessage(tgbotapi.NewMessage(update.Message.Chat.ID, "It seems that you are not involved in any game so your message fell on deaf ears."))
		} else {
			bot.ForwardMessageToGame(&update.Message, GameID)
		}
	}
}

// This method forwards a message from a player to the rest of the group.
func (bot *CAHBot) SendMessageToGame(GameID, message string) {
	tx, err := bot.db_conn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	rows, err := bot.db_conn.Query("SELECT get_userids_for_game($1)", GameID)
	defer rows.Close()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	var ID int
	for rows.Next() {
		if err := rows.Scan(&ID); err != nil {
			log.Printf("ERROR: %v", err)
		} else {
			bot.SendMessage(tgbotapi.NewMessage(ID, message))
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: %v", err)
	}
}

// This method forwards a message from a player to the rest of the group.
func (bot *CAHBot) ForwardMessageToGame(m *tgbotapi.Message, GameID string) {
	tx, err := bot.db_conn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(m.From.ID)
		return
	}
	rows, err := bot.db_conn.Query("SELECT get_userids_for_game($1)", GameID)
	defer rows.Close()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(m.From.ID)
		return
	}
	var ID int
	for rows.Next() {
		if err := rows.Scan(&ID); err != nil {
			log.Printf("ERROR: %v", err)
		} else if ID != m.From.ID {
			bot.ForwardMessage(tgbotapi.NewForward(ID, m.Chat.ID, m.MessageID))
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("ERROR: %v", err)
	}
}

// Send a 'There is no game' message
func (bot *CAHBot) SendNoGameMessage(ChatID int) {
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "You are currently not in a game.  Use command '/create' to create a new one or '/join <id>' to join a game with an id."))
}

func (bot *CAHBot) WrongCommand(ChatID int) {
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "Sorry, I don't know that command."))
}

// This method sends a generic sorry message.
func (bot *CAHBot) SendActionFailedMessage(ChatID int) {
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "I'm sorry, but it seems I have have difficulties right now.  You can try again later or contact my developer @thedadams."))
}

// Here we detect the kind of message we received from the user.
func (bot *CAHBot) DetectKindMessageRecieved(m *tgbotapi.Message) string {
	log.Printf("Detecting the type of message received")
	if m.Text != "" {
		if strings.HasPrefix(m.Text, "/") {
			return "command"
		} else {
			return "message"
		}
	}
	if len(m.Photo) != 0 {
		return "photo"
	}
	if m.Audio.FileID != "" {
		return "audio"
	}
	if m.Video.FileID != "" {
		return "video"
	}
	if m.Document.FileID != "" {
		return "document"
	}
	if m.Sticker.FileID != "" {
		return "sticker"
	}
	if m.NewChatParticipant.ID != 0 {
		return "newparicipant"
	}
	if m.LeftChatParticipant.ID != 0 {
		return "byeparticipant"
	}
	if m.NewChatTitle != "" {
		return "newchattitle"
	}
	if len(m.NewChatPhoto) != 0 {
		return "newchatphoto"
	}
	if m.DeleteChatPhoto {
		return "deletechatphoto"
	}
	if m.GroupChatCreated {
		return "newgroupchat"
	}
	if m.Contact.UserID != "" || m.Contact.FirstName != "" || m.Contact.LastName != "" {
		return "contact"
	}
	if m.Location.Longitude != 0 && m.Location.Latitude != 0 {
		return "location"
	}

	return "undetermined"
}

// Here, we know we have a command, we figure out which command the user invoked,
// and call the appropriate method.
func (bot *CAHBot) ProccessCommand(m *tgbotapi.Message, GameID string) {
	log.Printf("Processing command....")
	// Get the command.
	switch strings.ToLower(strings.Replace(strings.Fields(m.Text)[0], "/", "", 1)) {
	case "start":
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Welcome to Cards Against Humanity for Telegram.  To create a new game, use the command '/create'.  If you create a game, you will be given a 5 character id you can share with friends so they can join you.  You can also join a game using the '/join <id>' command where the '<id>' is replaced with a game id created by someone else.  To see all available commands, use '/help'."))
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "While you are in a game, any (non-command) message you send to me will be automatically forwarded to everyone else in the game so you're all in the loop."))
	case "help":
		// TODO: use helpers to build a help message.
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "A help message should go here."))
	case "create":
		if GameID != "" {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "You are already part of a game with id "+GameID+" and cannot create another game.  You can leave your current game with the command '/leave'."))
		} else {
			ID := bot.CreateNewGame(m.Chat.ID, m.From)
			if ID != "" {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "The game was created successfully.  Tell your friends to use the command '/join "+ID+"' to join your game.  Remember that your game will be deleted after 2 days of inactivity."))
				bot.AddPlayerToGame(ID, m.From)
			} else {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "An error occurred while trying to create the game.  The game was not created."))
			}
		}
	case "remove":
		tx, err := bot.db_conn.Begin()
		defer tx.Rollback()
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendActionFailedMessage(m.Chat.ID)
			return
		}
		// If the user is in a game, we remove them.
		if GameID != "" {
			bot.RemovePlayerFromGame(GameID, m.From)
		}
		log.Printf("Removing user from the database.")
		_, err = tx.Exec("SELECT remove_user($1)", m.From.ID)
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendActionFailedMessage(m.Chat.ID)
			return
		}
		tx.Commit()

		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "You have been removed from our records. If you ever want to come back, send the command '/start'.  Thank you for playing."))

	case "begin", "resume":
		if GameID != "" {
			bot.BeginGame(GameID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "end":
		if GameID != "" {
			bot.EndGame(GameID, m.From)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "pause":
		if GameID != "" {
			bot.PauseGame(GameID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "join":
		if len(strings.Fields(m.Text)) > 1 {
			if GameID != "" {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "You are already part of a game with id "+GameID+" and cannot join another game.  You can leave your current game with the command '/leave'."))
			} else {
				// If the user is not part of another game, we check to see if they id they
				// game is valid.
				tx, err := bot.db_conn.Begin()
				defer tx.Rollback()
				if err != nil {
					log.Printf("ERROR: %v", err)
					bot.SendActionFailedMessage(m.Chat.ID)
					return
				}
				var exists bool
				row := tx.QueryRow("SELECT check_game_exists($1)", strings.Fields(m.Text)[1])
				if _ = row.Scan(&exists); exists {
					// The id is valid and we add them.
					bot.AddPlayerToGame(strings.Fields(m.Text)[1], m.From)
				} else {
					bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "There is no game with id "+strings.Fields(m.Text)[1]+".  Please try again with a new id or use '/create' to create a game."))
				}
			}
		} else {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "You did not enter a game id.  Try again with the format '/join <id>'."))
		}
	case "gameid":
		if GameID != "" {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "The game you are currently playing has id "+GameID+".  Others can join your game by using the command '/join "+GameID+"'."))
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "leave":
		if GameID != "" {
			bot.RemovePlayerFromGame(GameID, m.From)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "next":
		if GameID != "" {
			bot.StartRound(GameID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "mycards":
		if GameID != "" {
			bot.ListCardsForUserWithMessage(GameID, m.From.ID, "Your cards are listed in the keyboard area.")
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "scores":
		if GameID != "" {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Here are the current scores:\n"+GameScores(GameID, bot.db_conn)))
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
			bot.ChangeGameSettings(GameID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "whoistzar":
		if GameID != "" {
			var Tzar string
			tx, err := bot.db_conn.Begin()
			defer tx.Rollback()
			if err != nil {
				log.Printf("ERROR: %v", err)
				bot.SendActionFailedMessage(m.Chat.ID)
				return
			}
			err = tx.QueryRow("SELECT whoistzar($1)", GameID).Scan(&Tzar)
			if err != nil {
				log.Printf("ERROR: %v", err)
				bot.SendActionFailedMessage(m.Chat.ID)
				return
			}
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "The current Card Tzar is "+Tzar+"."))
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "feedback":
		bot.ReceiveFeedback(m.Chat.ID)
	case "logging":
		if len(strings.Fields(m.Text)) > 1 {
			hasher := sha512.New()
			if strings.EqualFold(base64.URLEncoding.EncodeToString(hasher.Sum([]byte(strings.Fields(m.Text)[1]))), secrets.AppPass) {
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

// Add a player to a game if the player is not playing.
func (bot *CAHBot) AddPlayerToGame(GameID string, User tgbotapi.User) {
	// This is supposed to check that there are not more than 10 players in a game.
	tx, err := bot.db_conn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(User.ID)
		return
	}
	var numPlayersInGame int
	err = tx.QueryRow("SELECT num_players_in_game($1)", GameID).Scan(&numPlayersInGame)
	if numPlayersInGame > 10 {
		bot.SendMessage(tgbotapi.NewMessage(User.ID, "Player limit of 10 reached, we can not add any more players."))
	} else {
		var tmp bool
		row := tx.QueryRow("SELECT is_player_in_game($2,$1)", GameID, User.ID)
		if _ = row.Scan(&tmp); tmp {
			bot.SendMessage(tgbotapi.NewMessage(User.ID, "You are already playing in this game.  Use command '/leave' to remove yourself."))
		} else {
			log.Printf("Adding %v to the game %v...", User, GameID)
			tx.Exec("SELECT add_player_to_game($1, $2)", GameID, User.ID)
			if tx.Commit() != nil {
				log.Printf("ERROR %T: %v", err, err)
				bot.SendActionFailedMessage(User.ID)
				return
			}
			bot.SendMessageToGame(GameID, User.String()+" has joined the game!")
		}
	}
}

// This method adds a user to the database. It does not link them to a game.
func (bot *CAHBot) AddUserToDatabase(User tgbotapi.User, ChatID int) bool {
	// Check to see if the user is already in the database.
	var exists bool
	tx, err := bot.db_conn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("Cannot connect to the database.")
		return false
	}
	_ = tx.QueryRow("SELECT does_user_exist($1)", User.ID).Scan(&exists)
	if !exists {
		log.Printf("Adding user with ID %v to the database.", User.ID)
		_, err = tx.Exec("SELECT add_user($1,$2,$3,$4,$5)", User.ID, User.FirstName, User.LastName, User.UserName, User.String())
		if err != nil {
			log.Printf("ERROR: %v", err)
			bot.SendActionFailedMessage(User.ID)
			return false
		}
		tx.Commit()
		return true
	} else {
		log.Printf("User with id %v is already in the database.", User.ID)
	}
	return true
}

// This method begins an already created game.
func (bot *CAHBot) BeginGame(GameID string) {
	// If there is only one person in the game, the app will crash if we continue.
	if len(bot.CurrentGames[GameID].Players) < 2 {
		log.Printf("We could not start the game because there aren't enough players to do so.  Only %v player. ", len(bot.CurrentGames[GameID].Players))
		bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, "There aren't enough people in the game to start it.  Please have others join using the '/join' command."))
	} else {
		log.Printf("Starting game for Chat ID %v.", GameID)
		// There is a bug in Go that does not allow for things like bot.CurrentGames[ChatID].HasBegun = true.  This is a workaround.
		tmp := bot.CurrentGames[GameID]
		tmp.HasBegun = true
		bot.CurrentGames[GameID] = tmp
		if DoWeHaveAllAnswers(bot.CurrentGames[GameID].Players) {
			log.Printf("Asking the Card Tzar, %v, to pick the best and/or worse answer.", bot.CurrentGames[GameID].Players[bot.CurrentGames[GameID].CardTzarOrder[bot.CurrentGames[GameID].CardTzarIndex]].Player)
			bot.TzarChooseAnswer(GameID)
		} else {
			bot.StartRound(GameID)
		}
	}
}

func (bot *CAHBot) ChangeGameSettings(GameID string) {

}

// This method creates a new game.
func (bot *CAHBot) CreateNewGame(ChatID int, User tgbotapi.User) string {
	tx, err := bot.db_conn.Begin()
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
	shuffle(ShuffledQuestionCards)
	ShuffledAnswerCards := make([]int, len(bot.AllAnswerCards))
	for i := 0; i < len(ShuffledAnswerCards); i++ {
		ShuffledAnswerCards[i] = i
	}
	shuffle(ShuffledAnswerCards)
	if err != nil {
		log.Printf("Error creating game: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return ""
	}
	tx.Exec("SELECT add_game($1,$2,$3,$4)", GameID, ArrayTransforForPostgres(ShuffledQuestionCards), ArrayTransforForPostgres(ShuffledAnswerCards), User.ID)
	err = tx.Commit()
	if err != nil {
		log.Printf("Game could not be created. ERROR: %v", err)
		bot.SendActionFailedMessage(ChatID)
		return ""
	}
	log.Printf("Game with id %v created successfully!", GameID)
	return GameID
}

// Sends a message show the players the question card.
func (bot *CAHBot) DisplayQuestionCard(GameID string) {
	log.Printf("Getting question card index for game with id %v", GamID)
	tx, err := bot.db_conn.Begin()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	var index int
	err = tx.QueryRow("SELECT get_question_card($1)", GameID).Scan(&index)
	tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		return
	}
	log.Printf("Sending question card to game with ID %v...", GameID)
	var message string = "Here is the question card:\n"
	message += bot.AllQuestionCards[index].Text
	bot.SendMessageToGame(GameID, html.UnescapeString(message))
}

// This method stops and ends an already created game.
func (bot *CAHBot) EndGame(GameID string, User tgbotapi.User) {
	tx, err := bot.db_conn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendMessageToGame(GameID, "There was an error when I tried to end the game.  You can try again or contact my developer @thedadams.")
		return
	}
	log.Printf("Deleting a game with id %v...", GameID)
	rows, err := tx.Query("SELECT end_game($1)", GameID)
	defer rows.Close()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendMessageToGame(GameID, "There was an error when I tried to end the game.  You can try again or contact my developer @thedadams.")
		return
	}
	if User.ID != -1 {
		// Someone ended the game.
		bot.SendMessageToGame(GameID, "The game has been stopped "+User.String()+".  Here are the scores:\n"+BuildScoreList(rows)+"Thanks for playing!")
	} else {
		// The game ended because someone won.
		bot.SendMessageToGame(GameID, "The game has ended.  Here are the scores:\n"+BuildScoreList(rows)+"Thanks for playing!")
	}
	tx.Commit()
}

// This method asks the Card Tzar to make a choice.
func (bot *CAHBot) GetQuestionCard(GameID string) {
	Game := bot.CurrentGames[GameID]
	Game.QuestionCard = Game.ShuffledQuestionCards[Game.NumQCardsLeft]
	Game.NumQCardsLeft -= 1
	Game.WaitingForAnswers = true
	log.Printf("The question card is %v: %v", Game.QuestionCard, bot.AllQuestionCards[Game.QuestionCard])
	if Game.NumQCardsLeft == -1 {
		log.Printf("Reshuffling question cards...")
		ReshuffleQCards(Game)
	}
	// This is the dumb Go map bug again.
	bot.CurrentGames[GameID] = Game
}

// This method lists the answers for everyone and allows the Tzar to choose one.
func (bot *CAHBot) ListAnswers(GameID string) {
	Tzar := bot.CurrentGames[GameID].Players[bot.CurrentGames[GameID].CardTzarOrder[bot.CurrentGames[GameID].CardTzarIndex]]
	cards := BuildAnswerList(bot.CurrentGames[GameID])
	text := "Here are the submitted answers:\n\n"
	for i := range cards {
		text += cards[i][0] + "\n"
	}
	log.Printf("Showing everyone the answers submitted for game %v.", GameID)
	message := tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, text)
	bot.SendMessage(message)
	if Tzar.Player.UserName != "" {
		message.Text += "@" + Tzar.Player.UserName
	} else {
		message.ReplyToMessageID = Tzar.ReplyID
	}
	message.Text = "Tzar " + Tzar.Player.String() + ", please choose the best answer."
	message.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{cards, true, true, true}
	bot.SendMessage(message)
}

// This method lists a user's cards using a custom keyboard in the Telegram API.  If we need them to respond to a question, this is handled.
func (bot *CAHBot) ListCardsForUserWithMessage(GameID string, UserID int, text string) {
	log.Printf("Showing the user %v their cards.", UserID)
	message := tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, text)
	cards := make([][]string, len(bot.CurrentGames[GameID].Players[strconv.Itoa(UserID)].Cards))
	for i := range cards {
		cards[i] = make([]string, 1)
	}
	for i := 0; i < len(bot.CurrentGames[GameID].Players[strconv.Itoa(UserID)].Cards); i++ {
		cards[i][0] = html.UnescapeString(bot.AllAnswerCards[bot.CurrentGames[GameID].Players[strconv.Itoa(UserID)].Cards[i]].Text)
	}
	message.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{cards, true, true, true}
	bot.SendMessage(message)
}

func (bot *CAHBot) MoveCardTzar(GameID string) {
	log.Printf("Switch the Card Tzar...")
	Game := bot.CurrentGames[GameID]
	Game.CardTzarIndex = (Game.CardTzarIndex + 1)
	if Game.CardTzarIndex >= len(Game.CardTzarOrder) {
		Game.CardTzarIndex = Game.CardTzarIndex % len(Game.CardTzarOrder)
		if Game.Settings.TradeInCardsEveryRound {
			log.Printf("We are about to start a new round.  Letting the players trade in a card.")
			for _ = range Game.Players {
				// This just to get it to compile for now.
				ChatID := 00000
				bot.ListCardsForUserWithMessage(GameID, ChatID, "Please choose "+strconv.Itoa(Game.Settings.NumCardsToTradeIn)+" card to trade in.")
			}
		}
	}
	log.Printf("The Card Tzar is now %v.", Game.Players[Game.CardTzarOrder[Game.CardTzarIndex]].Player)
	bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, "The Card Tzar is now "+Game.Players[Game.CardTzarOrder[Game.CardTzarIndex]].Player.String()))
	bot.CurrentGames[GameID] = Game
}

// This method pauses a started game.
func (bot *CAHBot) PauseGame(GameID string) {
	log.Printf("Pausing game for Chat %v...", GameID)
	// There is a bug in Go that does not allow for things like bot.CurrentGames[ChatID].HasStarted = false.  This is a workaround.
	tmp := bot.CurrentGames[GameID]
	tmp.HasBegun = false
	bot.CurrentGames[GameID] = tmp
	bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, "The game has been paused.  Use command '/resume' to resume."))
}

func (bot *CAHBot) ReceiveFeedback(ChatID int) {

}

// Remove a player from a game if the player is playing.
func (bot *CAHBot) RemovePlayerFromGame(GameID string, User tgbotapi.User) {
	tx, err := bot.db_conn.Begin()
	defer tx.Rollback()
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(User.ID)
		return
	}
	log.Printf("Removing %v from the game %v...", User, GameID)
	var str string = ""
	err = tx.QueryRow("SELECT remove_player_from_game($1)", User.ID).Scan(&str)
	if err != nil {
		log.Printf("ERROR: %v", err)
		bot.SendActionFailedMessage(User.ID)
		return
	}
	bot.SendMessage(tgbotapi.NewMessage(User.ID, "Thanks for playing, "+User.String()+"!  You collected "+strings.Split(str[1:len(str)-1], ",")[1]+" cards."))
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
		bot.EndGame(GameID, User)
	} else {
		bot.SendMessageToGame(GameID, User.String()+" has left the game with a score of "+strings.Split(str[1:len(str)-1], ",")[1]+".")
	}
}

func (bot *CAHBot) SendGameSettings(GameID string, ChatID int) {
	log.Printf("Sending game settings for %v.", ChatID)
	bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, "Game settings: \n"+bot.CurrentGames[GameID].Settings.String()))
}

// This method handles the starting/resuming of a round.
func (bot *CAHBot) StartRound(GameID string) {
	// Check to see if the game is running and if we are waiting for answers.
	if bot.CurrentGames[GameID].HasBegun {
		if bot.CurrentGames[GameID].WaitingForAnswers {
			bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, "We are waiting for other players to answer the question."))
		} else {
			// Once we get here, we are either starting a game, resuming a game, or going onto another round.
			// Check to see if someone won.
			if winner, ans := DidSomeoneWin(bot.CurrentGames[GameID]); ans {
				// Someone won, so we end the game.
				log.Printf("%v won the game with ID %v.", winner, GameID)
				bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, "We have a winner!  Congratulations to "+winner.Player.String()+" on the victory.  We are now ending the game."))
				bot.EndGame(GameID, tgbotapi.User{ID: -1})
			} else {
				if bot.CurrentGames[GameID].CardTzarIndex == -1 {
					log.Printf("Start a new game for chat ID %v.", GameID)
					bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, "Get ready.  We are starting the game!"))
					bot.MoveCardTzar(GameID)
				}
				if bot.CurrentGames[GameID].QuestionCard == -1 {
					bot.GetQuestionCard(GameID)
					bot.DisplayQuestionCard(GameID)
				}
				for _, value := range bot.CurrentGames[GameID].Players {
					if !value.IsCardTzar && value.AnswerBeingPlayed == "" {
						log.Printf("Asking %v for an answer card.", value)
						bot.ListCardsForUserWithMessage(GameID, value.Player.ID, "Please pick an answer for the question.")
					}
				}
			}
		}
	} else {
		bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[GameID].ChatID, "The game is not currently running.  Use command '/resume' to start it up."))
	}
}

// This method handles the Tzar choosing an answer.
func (bot *CAHBot) TzarChooseAnswer(GameID string) {
	game := bot.CurrentGames[GameID]

	game.CardTzarIndex = -1
	bot.CurrentGames[GameID] = game

}
