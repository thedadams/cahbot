package main

import (
	"cahbot/secrets"
	"cahbot/tgbotapi"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"html"
	"log"
	"strconv"
	"strings"
)

// This is the starting point for handling an update from chat.
func (bot *CAHBot) HandleUpdate(update *tgbotapi.Update) {
	messageType := bot.DetectKindMessageRecieved(&update.Message)
	log.Printf("[%s] Message type: %s", update.Message.From.UserName, messageType)
	if messageType == "command" {
		bot.ProccessCommand(&update.Message)
	} else if messageType == "message" || messageType == "photo" || messageType == "video" || messageType == "audio" || messageType == "contact" || messageType == "document" || messageType == "location" || messageType == "sticker" {
		bot.ForwardMessageToGroup(&update.Message)
	}
}

// This method forwards a message from a player to the rest of the group.
func (bot *CAHBot) ForwardMessageToGroup(m *tgbotapi.Message) {
	tx, err := bot.db_conn.Begin()
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
	rows, err := tx.Query("SELECT users.chat_id FROM users, games, players WHERE players.user_id = users.id AND players.game_id = games.id")
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
	defer rows.Close()
	defer tx.Rollback()
	var ID int
	for rows.Next() {
		if err := rows.Scan(&ID); err != nil {
			log.Printf("ERROR: %v", err)
		}
		bot.ForwardMessage(tgbotapi.NewForward(ID, m.Chat.ID, m.MessageID))
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
func (bot *CAHBot) ProccessCommand(m *tgbotapi.Message) {
	log.Printf("Processing command....")
	ChatID := ""
	// Get the command.
	switch strings.ToLower(strings.Replace(strings.Fields(m.Text)[0], "/", "", 1)) {
	case "start":
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Welcome to Cards Against Humanity for Telegram.  To create a new game, use the command '/create'.  If you create a game, you will be given a 6 character id you can share with friends so they can join you.  You can also join a game using the '/join <id>' command where the '<id>' is replaced with a game id created by someone else.  To see all available commands, use '/help'."))
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "While you are in a game, any (non-command) message you send to me will be automatically forwarded to everyone else in the game so you're all in the loop."))
		log.Printf("Adding user with ID %v to the database.", m.From.ID)
		bot.AddUserToDatabase(m.From, m.Chat.ID)
	case "help":
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "A help message should go here."))
	case "create":
		if value, ok := bot.CurrentGames[ChatID]; ok {
			if value.HasBegun {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "You are already in a game.  Use command '/leave' to leave the previous game."))
			} else {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "You are already in a game.  Use command '/leave' to leave the previous game or '/resume' to resume it."))
			}
		} else {
			ID := bot.CreateNewGame(ChatID, m.From)
			if ID != "" {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "The game was created successfully.  Tell your friends to use the command '/join "+ID+"' to join your game."))
			} else {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "An error occurred while trying to create the game.  The game was not created."))
			}
		}
	case "begin", "resume":
		if _, ok := bot.CurrentGames[ChatID]; ok {
			bot.BeginGame(ChatID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "stop":
		if _, ok := bot.CurrentGames[ChatID]; ok {
			bot.StopGame(ChatID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "pause":
		if value, ok := bot.CurrentGames[ChatID]; ok {
			if value.HasBegun {
				bot.PauseGame(ChatID)
			} else {
				ID, _ := strconv.Atoi(ChatID)
				bot.SendMessage(tgbotapi.NewMessage(ID, "The current game is already paused.  Use command '/resume' to resume it."))
			}
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "join":
		if _, ok := bot.CurrentGames[ChatID]; ok {
			bot.AddPlayerToGame(ChatID, m.From, m.MessageID, false)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "leave":
		if _, ok := bot.CurrentGames[ChatID]; ok {
			bot.RemovePlayerFromGame(ChatID, m.From)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "next":
		if _, ok := bot.CurrentGames[ChatID]; ok {
			bot.StartRound(ChatID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "mycards":
		if game, ok := bot.CurrentGames[ChatID]; ok {
			if _, yes := game.Players[strconv.Itoa(m.From.ID)]; yes {
				bot.ListCardsForUserWithMessage(ChatID, bot.CurrentGames[ChatID].Players[strconv.Itoa(m.From.ID)], "Your cards are listed in the keyboard area.")
			} else {
				message := tgbotapi.NewMessage(m.Chat.ID, m.From.String()+", you are not in the current game, so I cannot show you your cards.  Use command '/join' to join the game.")
				message.ReplyToMessageID = m.MessageID
				bot.SendMessage(message)
			}
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "scores":
		if value, ok := bot.CurrentGames[ChatID]; ok {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Here are the current scores:\n"+value.Scores()))
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "settings":
		if _, ok := bot.CurrentGames[ChatID]; ok {
			bot.SendGameSettings(ChatID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "changesettings":
		if _, ok := bot.CurrentGames[ChatID]; ok {
			bot.ChangeGameSettings(ChatID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "whoistzar":
		if value, ok := bot.CurrentGames[ChatID]; ok {
			if bot.CurrentGames[ChatID].CardTzarIndex == -1 {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "It looks like the game hasn't started yet so we don't have a Tzar.  Use command '/begin' to start the game."))
			} else {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "The current Card Tzar is "+value.Players[value.CardTzarOrder[value.CardTzarIndex]].Player.String()+"."))
			}
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "feedback":
		bot.ReceiveFeedback(ChatID)
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
	case "status":
		if len(strings.Fields(m.Text)) > 1 {
			hasher := sha512.New()
			if strings.EqualFold(base64.URLEncoding.EncodeToString(hasher.Sum([]byte(strings.Fields(m.Text)[1]))), secrets.AppPass) {
				message := "There are currently " + strconv.Itoa(len(bot.CurrentGames)) + " games being played."
				log.Printf("Sending status message...")
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, message))
			}
		} else {
			bot.WrongCommand(m.Chat.ID)
		}
	default:
		bot.WrongCommand(m.Chat.ID)
	}
}

// This method adds a user to the database. It does not link them to a game.
func (bot *CAHBot) AddUserToDatabase(User tgbotapi.User, ChatID int) {
	// Check to see if the user is already in the database.
	var OldChatID int
	tx, err := bot.db_conn.Begin()
	if err != nil {
		log.Printf("Cannot connect to the database.")
	}
	err = tx.QueryRow("SELECT chat_id FROM users WHERE id=$1", User.ID).Scan(&OldChatID)
	switch {
	case err == sql.ErrNoRows:
		// The user is not in the database so we add them.
		tx.Exec("INSERT INTO users (id, first_name, last_name, username, chat_id, points, cards_in_hand, current_tzar, current_answer) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)", User.ID, User.FirstName, User.LastName, User.UserName, ChatID, 0, nil, false, "")
	case err != nil:
		// An unknown error occurred.
		log.Printf("ERROR %T: %v", err, err)
	default:
		log.Printf("User with id %v is already in the database.", User.ID)
		tx.Exec("UPDATE users SET chat_id=$1 WHERE id=$2", ChatID, User.ID)
	}
	err = tx.Commit()
	if err != nil {
		log.Printf("Unable to commit add of user with id %v to the database.", User.ID)
	}
}

// This method creates a new game.
func (bot *CAHBot) CreateNewGame(ChatID string, User tgbotapi.User) string {
	tx, err := bot.db_conn.Begin()
	var GameID string
	for {
		GameID = GetRandomID()
		var tmp string
		err := tx.QueryRow("SELECT id FROM games WHERE id=$1", GameID).Scan(&tmp)
		if err == sql.ErrNoRows {
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
		return ""
	}
	tx.Exec("INSERT INTO games(id, question_cards, answer_cards, qcards_left, acards_left, tzar_order, tzar_index, current_qcard, has_begun, waiting_for_answers, mystery_player, trade_in_cards, num_cards_to_trade, pick_worst, num_cards_in_hand, points_to_win) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)", GameID, ArrayTransforForPostgres(ShuffledQuestionCards), ArrayTransforForPostgres(ShuffledAnswerCards), len(ShuffledQuestionCards), len(ShuffledAnswerCards), "{"+strconv.Itoa(User.ID)+"}", -1, -1, false, false, false, false, 1, false, 7, 7)
	err = tx.Commit()
	if err != nil {
		log.Printf("Game could not be created. ERROR: %v", err)
		return ""
	}
	log.Printf("Game with id %v created successfully!", GameID)
	//bot.AddPlayerToGame(ChatID, User, true)
	return GameID
}

// This method begins an already created game.
func (bot *CAHBot) BeginGame(ChatID string) {
	// If there is only one person in the game, the app will crash if we continue.
	if len(bot.CurrentGames[ChatID].Players) < 2 {
		log.Printf("We could not start the game because there aren't enough players to do so.  Only %v player. ", len(bot.CurrentGames[ChatID].Players))
		bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "There aren't enough people in the game to start it.  Please have others join using the '/join' command."))
	} else {
		log.Printf("Starting game for Chat ID %v.", ChatID)
		// There is a bug in Go that does not allow for things like bot.CurrentGames[ChatID].HasBegun = true.  This is a workaround.
		tmp := bot.CurrentGames[ChatID]
		tmp.HasBegun = true
		bot.CurrentGames[ChatID] = tmp
		if DoWeHaveAllAnswers(bot.CurrentGames[ChatID].Players) {
			log.Printf("Asking the Card Tzar, %v, to pick the best and/or worse answer.", bot.CurrentGames[ChatID].Players[bot.CurrentGames[ChatID].CardTzarOrder[bot.CurrentGames[ChatID].CardTzarIndex]].Player)
			bot.TzarChooseAnswer(ChatID)
		} else {
			bot.StartRound(ChatID)
		}
	}
}

// This method handles the starting/resuming of a round.
func (bot *CAHBot) StartRound(ChatID string) {
	// Check to see if the game is running and if we are waiting for answers.
	if bot.CurrentGames[ChatID].HasBegun {
		if bot.CurrentGames[ChatID].WaitingForAnswers {
			bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "We are waiting for other players to answer the question."))
		} else {
			// Once we get here, we are either starting a game, resuming a game, or going onto another round.
			// Check to see if someone won.
			if winner, ans := DidSomeoneWin(bot.CurrentGames[ChatID]); ans {
				// Someone won, so we end the game.
				log.Printf("%v won the game with ID %v.", winner, ChatID)
				bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "We have a winner!  Congratulations to "+winner.Player.String()+" on the victory.  We are now ending the game."))
				bot.StopGame(ChatID)
			} else {
				if bot.CurrentGames[ChatID].CardTzarIndex == -1 {
					log.Printf("Start a new game for chat ID %v.", ChatID)
					bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "Get ready.  We are starting the game!"))
					bot.MoveCardTzar(ChatID)
				}
				if bot.CurrentGames[ChatID].QuestionCard == -1 {
					bot.GetQuestionCard(ChatID)
					bot.DisplayQuestionCard(ChatID)
				}
				for _, value := range bot.CurrentGames[ChatID].Players {
					if !value.IsCardTzar && value.AnswerBeingPlayed == "" {
						log.Printf("Asking %v for an answer card.", value)
						bot.ListCardsForUserWithMessage(ChatID, value, "Please pick an answer for the question.")
					}
				}
			}
		}
	} else {
		bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "The game is not currently running.  Use command '/resume' to start it up."))
	}
}

// This method pauses a started game.
func (bot *CAHBot) PauseGame(ChatID string) {
	log.Printf("Pausing game for Chat %v...", ChatID)
	// There is a bug in Go that does not allow for things like bot.CurrentGames[ChatID].HasStarted = false.  This is a workaround.
	tmp := bot.CurrentGames[ChatID]
	tmp.HasBegun = false
	bot.CurrentGames[ChatID] = tmp
	bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "The game has been paused.  Use command '/resume' to resume."))
}

// This method stops and ends an already created game.
func (bot *CAHBot) StopGame(ChatID string) {
	bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "The game has been stopped.  Here are the scores:\n"+bot.CurrentGames[ChatID].Scores()+"Thanks for playing!"))
	log.Printf("Deleting a game with Chat ID %v...", ChatID)
	delete(bot.CurrentGames, ChatID)
}

// Sends a message show the players the question card.
func (bot *CAHBot) DisplayQuestionCard(ChatID string) {
	log.Printf("Sending question card to game with ID %v...", ChatID)
	var message string = "Here is the question card:\n"
	message += bot.AllQuestionCards[bot.CurrentGames[ChatID].QuestionCard].Text
	bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, html.UnescapeString(message)))
}

// This method handles the Tzar choosing an answer.
func (bot *CAHBot) TzarChooseAnswer(ChatID string) {
	game := bot.CurrentGames[ChatID]

	game.CardTzarIndex = -1
	bot.CurrentGames[ChatID] = game

}

// This method asks the Card Tzar to make a choice.
func (bot *CAHBot) GetQuestionCard(ChatID string) {
	Game := bot.CurrentGames[ChatID]
	Game.QuestionCard = Game.ShuffledQuestionCards[Game.NumQCardsLeft]
	Game.NumQCardsLeft -= 1
	Game.WaitingForAnswers = true
	log.Printf("The question card is %v: %v", Game.QuestionCard, bot.AllQuestionCards[Game.QuestionCard])
	if Game.NumQCardsLeft == -1 {
		log.Printf("Reshuffling question cards...")
		ReshuffleQCards(Game)
	}
	// This is the dumb Go map bug again.
	bot.CurrentGames[ChatID] = Game
}

// This method lists a user's cards using a custom keyboard in the Telegram API.  If we need them to respond to a question, this is handled.
func (bot *CAHBot) ListCardsForUserWithMessage(ChatID string, Player PlayerGameInfo, text string) {
	log.Printf("Showing the user %v their cards.", Player.Player.String())
	message := tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, text)
	cards := make([][]string, len(bot.CurrentGames[ChatID].Players[strconv.Itoa(Player.Player.ID)].Cards))
	for i := range cards {
		cards[i] = make([]string, 1)
	}
	if Player.Player.UserName != "" {
		message.Text += "@" + Player.Player.UserName
	} else {
		message.ReplyToMessageID = Player.ReplyID
	}
	for i := 0; i < len(bot.CurrentGames[ChatID].Players[strconv.Itoa(Player.Player.ID)].Cards); i++ {
		cards[i][0] = html.UnescapeString(bot.AllAnswerCards[bot.CurrentGames[ChatID].Players[strconv.Itoa(Player.Player.ID)].Cards[i]].Text)
	}
	message.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{cards, true, true, true}
	bot.SendMessage(message)
}

// This method lists the answers for everyone and allows the Tzar to choose one.
func (bot *CAHBot) ListAnswers(ChatID string) {
	Tzar := bot.CurrentGames[ChatID].Players[bot.CurrentGames[ChatID].CardTzarOrder[bot.CurrentGames[ChatID].CardTzarIndex]]
	cards := BuildAnswerList(bot.CurrentGames[ChatID])
	text := "Here are the submitted answers:\n\n"
	for i := range cards {
		text += cards[i][0] + "\n"
	}
	log.Printf("Showing everyone the answers submitted for game %v.", ChatID)
	message := tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, text)
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

func (bot *CAHBot) SendGameSettings(ChatID string) {
	log.Printf("Sending game settings for %v.", ChatID)
	bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "Game settings: \n"+bot.CurrentGames[ChatID].Settings.String()))
}

func (bot *CAHBot) ChangeGameSettings(ChatID string) {

}

func (bot *CAHBot) ReceiveFeedback(ChatID string) {

}

// Add a player to a game if the player is not playing.
func (bot *CAHBot) AddPlayerToGame(ChatID string, User tgbotapi.User, MessageID int, MakeTzar bool) {
	if len(bot.CurrentGames[ChatID].Players) > 9 {
		bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "Player limit of 10 reached, we can not add any more players."))
	} else {
		if _, ok := bot.CurrentGames[ChatID].Players[strconv.Itoa(User.ID)]; ok {
			bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, User.String()+" is already playing.  Use command '/leave' to remove yourself."))
		} else {
			log.Printf("Adding %v to the game %v...", User, ChatID)
			game := bot.CurrentGames[ChatID]
			PlayerHand := make([]int, 0, bot.CurrentGames[ChatID].Settings.NumCardsInHand)
			PlayerHand = DealPlayerHand(game, PlayerHand)
			bot.CurrentGames[ChatID].Players[strconv.Itoa(User.ID)] = PlayerGameInfo{User, MessageID, 0, PlayerHand, MakeTzar, ""}
			game.CardTzarOrder = append(bot.CurrentGames[ChatID].CardTzarOrder, strconv.Itoa(User.ID))
			bot.CurrentGames[ChatID] = game
			bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "Welcome to the game, "+User.String()+"!"))
		}
	}
}

// Remove a player from a game if the player is playing.
func (bot *CAHBot) RemovePlayerFromGame(ChatID string, User tgbotapi.User) {
	if value, ok := bot.CurrentGames[ChatID].Players[strconv.Itoa(User.ID)]; ok {
		bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "Thanks for playing, "+User.String()+"!  You collected "+strconv.Itoa(value.Points)+" cards."))
		log.Printf("Removing %v from the game %v...", User, ChatID)
		for i := 0; i < len(bot.CurrentGames[ChatID].CardTzarOrder); i++ {
			if bot.CurrentGames[ChatID].CardTzarOrder[i] == strconv.Itoa(User.ID) {
				// This is a workaround because go assignments inside a map don't work.
				game := bot.CurrentGames[ChatID]
				game.CardTzarOrder = append(bot.CurrentGames[ChatID].CardTzarOrder[:i], bot.CurrentGames[ChatID].CardTzarOrder[i+1:]...)
				bot.CurrentGames[ChatID] = game
				break
			}
		}
		if bot.CurrentGames[ChatID].Players[strconv.Itoa(User.ID)].IsCardTzar {
			bot.MoveCardTzar(ChatID)
		}
		delete(bot.CurrentGames[ChatID].Players, strconv.Itoa(User.ID))
		if len(bot.CurrentGames[ChatID].Players) == 0 {
			log.Printf("There are no more players in game %v.  We shall end it.", ChatID)
			bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "There are no more people playing in this game. We are going to end it."))
			bot.StopGame(ChatID)
		}
	} else {
		bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, User.String()+" is not playing yet.  Use command '/add' to add yourself."))
	}
}

func (bot *CAHBot) MoveCardTzar(ChatID string) {
	log.Printf("Switch the Card Tzar...")
	Game := bot.CurrentGames[ChatID]
	Game.CardTzarIndex = (Game.CardTzarIndex + 1)
	if Game.CardTzarIndex >= len(Game.CardTzarOrder) {
		Game.CardTzarIndex = Game.CardTzarIndex % len(Game.CardTzarOrder)
		if Game.Settings.TradeInCardsEveryRound {
			log.Printf("We are about to start a new round.  Letting the players trade in a card.")
			for key := range Game.Players {
				bot.ListCardsForUserWithMessage(ChatID, Game.Players[key], "Please choose "+strconv.Itoa(Game.Settings.NumCardsToTradeIn)+" card to trade in.")
			}
		}
	}
	log.Printf("The Card Tzar is now %v.", Game.Players[Game.CardTzarOrder[Game.CardTzarIndex]].Player)
	bot.SendMessage(tgbotapi.NewMessage(bot.CurrentGames[ChatID].ChatID, "The Card Tzar is now "+Game.Players[Game.CardTzarOrder[Game.CardTzarIndex]].Player.String()))
	bot.CurrentGames[ChatID] = Game
}
