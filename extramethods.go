package main

import (
	"cahbot/tgbotapi"
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
	}
}

// Send a 'There is no game' message
func (bot *CAHBot) SendNoGameMessage(ChatID int) {
	log.Printf("Telling them there is no game right now.")
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "There is no game being played here.  Use command '/create' to create a new one."))
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
		return "contant"
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
	// Get the command.
	switch strings.ToLower(strings.Replace(strings.Fields(m.Text)[0], "/", "", 1)) {
	case "start":
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Welcome to Cards Against Humanity for Telegram.  To create a new game, use the command '/create'.  To see all available commands, use '/help'."))
	case "help":
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "A help message should go here."))
	case "create":
		if value, ok := bot.CurrentGames[m.Chat.ID]; ok {
			if value.HasBegun {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "There is already a game created here.  Use command '/stop' to end the previous game."))
			} else {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "There is already a game created here.  Use command '/stop' to end the previous game or '/resume' to resume."))
			}
		} else {
			bot.CreateNewGame(m.Chat.ID, m.From, m.MessageID)
		}
	case "begin", "resume":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.BeginGame(m.Chat.ID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "stop":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.StopGame(m.Chat.ID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "pause":
		if value, ok := bot.CurrentGames[m.Chat.ID]; ok {
			if value.HasBegun {
				bot.PauseGame(m.Chat.ID)
			} else {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "The current game is already paused.  Use command '/resume' to resume it."))
			}
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "join":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.AddPlayerToGame(m.Chat.ID, m.From, m.MessageID, false)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "leave":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.RemovePlayerFromGame(m.Chat.ID, m.From)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "mycards":
		if game, ok := bot.CurrentGames[m.Chat.ID]; ok {
			if _, yes := game.Players[m.From.ID]; yes {
				bot.ListCardsForUser(m.Chat.ID, bot.CurrentGames[m.Chat.ID].Players[m.From.ID])
			} else {
				message := tgbotapi.NewMessage(m.Chat.ID, m.From.String()+", you are not in the current game, so I cannot show you your cards.  Use command '/join' to join the game.")
				message.ReplyToMessageID = m.MessageID
				bot.SendMessage(message)
			}
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "scores":
		if value, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Here are the current scores:\n"+value.Scores()))
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "settings":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.SendGameSettings(m.Chat.ID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "changesettings":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.ChangeGameSettings(m.Chat.ID)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "whoistzar":
		if value, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "The current Card Tzar is "+value.Players[value.CardTzarOrder[value.CardTzarIndex]].Player.String()+"."))
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "feedback":
		bot.ReceiveFeedback(m.Chat.ID)
	default:
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Sorry, I don't know that command."))
	}
}

// This method creates a new game.
func (bot *CAHBot) CreateNewGame(ChatID int, User tgbotapi.User, MessageID int) {
	log.Printf("Creating a new game for Chat ID %v.", ChatID)
	// Get the keys for the All Cards map.
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
	bot.CurrentGames[ChatID] = CAHGame{ChatID, ShuffledQuestionCards, ShuffledAnswerCards, len(ShuffledQuestionCards) - 1, len(ShuffledAnswerCards) - 1, make(map[int]PlayerGameInfo), []int{User.ID}, 0, -1, GameSettings{false, false, false, 7}, false, false}
	log.Printf("Game for Chat ID %v created successfully!%v", ChatID)
	bot.AddPlayerToGame(ChatID, User, MessageID, true)
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "The game was created successfully."))
}

// This method begins an already created game.
func (bot *CAHBot) BeginGame(ChatID int) {
	log.Printf("Starting game for Chat ID %v.", ChatID)
	// There is a bug in Go that does not allow for things like bot.CurrentGames[ChatID].HasBegun = true.  This is a workaround.
	tmp := bot.CurrentGames[ChatID]
	tmp.HasBegun = true
	bot.CurrentGames[ChatID] = tmp
	if DoWeHaveAllAnswers(bot.CurrentGames[ChatID].Players) {
		log.Printf("Asking the Card Tzar, %v, to pick the best and/or worse answer.", bot.CurrentGames[ChatID].Players[bot.CurrentGames[ChatID].CardTzarOrder[bot.CurrentGames[ChatID].CardTzarIndex]].Player)
		bot.GetQuestionCard(ChatID, true)
	} else {
		for _, value := range bot.CurrentGames[ChatID].Players {
			if !value.IsCardTzar && value.CardBeingPlayed == -1 {
				log.Printf("Asking %v for an answer card.", value)
				bot.ListCardsForUser(ChatID, value)
			}
		}
	}
}

// This method asks the Card Tzar to make a choice.
func (bot *CAHBot) GetQuestionCard(ChatID int, display bool) {
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
	if display {
		bot.DisplayQuestionCard(Game.ChatID)
	}
}

// Sends a message show the players the question card.
func (bot *CAHBot) DisplayQuestionCard(ChatID int) {
	log.Printf("Sending question care to game with ID %v...", ChatID)
	var message string = "Here is the question card:\n"
	message += bot.AllQuestionCards[bot.CurrentGames[ChatID].QuestionCard].Text
	bot.SendMessage(tgbotapi.NewMessage(ChatID, html.UnescapeString(message)))
}

// This method pauses a started game.
func (bot *CAHBot) PauseGame(ChatID int) {
	log.Printf("Pausing game for Chat %v...", ChatID)
	// There is a bug in Go that does not allow for things like bot.CurrentGames[ChatID].HasStarted = false.  This is a workaround.
	tmp := bot.CurrentGames[ChatID]
	tmp.HasBegun = false
	bot.CurrentGames[ChatID] = tmp
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "The game has been paused.  Use command '/resume' to resume."))
}

// This method stops and ends an already created game.
func (bot *CAHBot) StopGame(ChatID int) {
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "The game has been stopped.  Here are the scores:\n"+bot.CurrentGames[ChatID].Scores()+"Thanks for playing!"))
	log.Printf("Deleting a game with Chat ID %v...", ChatID)
	delete(bot.CurrentGames, ChatID)
}

// This method lists the cards for the user.  If we need them to respond to a question, this is handled.
func (bot *CAHBot) ListCardsForUser(ChatID int, Player PlayerGameInfo) {
	log.Printf("Showing the user %v their cards.", Player.Player.String())
	message := tgbotapi.NewMessage(ChatID, "")
	cards := make([][]string, len(bot.CurrentGames[ChatID].Players[Player.Player.ID].Cards))
	for i := range cards {
		cards[i] = make([]string, 1)
	}
	if Player.Player.UserName != "" {
		message.Text += "@" + Player.Player.UserName
	} else {
		message.ReplyToMessageID = Player.ReplyID
	}
	for i := 0; i < len(bot.CurrentGames[ChatID].Players[Player.Player.ID].Cards); i++ {
		cards[i][0] = html.UnescapeString(bot.AllAnswerCards[bot.CurrentGames[ChatID].Players[Player.Player.ID].Cards[i]].Text)
	}
	message.ReplyMarkup = tgbotapi.ReplyKeyboardMarkup{cards, true, true, true}
	bot.SendMessage(message)
}

func (bot *CAHBot) SendGameSettings(ChatID int) {
	log.Printf("Sending game settings for %v.", ChatID)
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "Game settings: \n"+bot.CurrentGames[ChatID].Settings.String()))
}

func (bot *CAHBot) ChangeGameSettings(ChatID int) {

}

func (bot *CAHBot) ReceiveFeedback(ChatID int) {

}

// Add a player to a game if the player is not playing.
func (bot *CAHBot) AddPlayerToGame(ChatID int, User tgbotapi.User, MessageID int, MakeTzar bool) {
	if len(bot.CurrentGames[ChatID].Players) > 9 {
		bot.SendMessage(tgbotapi.NewMessage(ChatID, "Player limit of 10 reached, we can not add any more players."))
	} else {
		if _, ok := bot.CurrentGames[ChatID].Players[User.ID]; ok {
			bot.SendMessage(tgbotapi.NewMessage(ChatID, User.String()+" is already playing.  Use command '/leave' to remove yourself."))
		} else {
			log.Printf("Adding %v to the game %v...", User, ChatID)
			game := bot.CurrentGames[ChatID]
			PlayerHand := make([]int, 0, bot.CurrentGames[ChatID].Settings.NumCardsInHand)
			PlayerHand = DealPlayerHand(game, PlayerHand)
			bot.CurrentGames[ChatID].Players[User.ID] = PlayerGameInfo{User, MessageID, 0, PlayerHand, MakeTzar, -1}
			game.CardTzarOrder = append(bot.CurrentGames[ChatID].CardTzarOrder, User.ID)
			bot.CurrentGames[ChatID] = game
			bot.SendMessage(tgbotapi.NewMessage(ChatID, "Welcome to the game, "+User.String()+"!"))
		}
	}
}

// Remove a player from a game if the player is playing.
func (bot *CAHBot) RemovePlayerFromGame(ChatID int, User tgbotapi.User) {
	if value, ok := bot.CurrentGames[ChatID].Players[User.ID]; ok {
		bot.SendMessage(tgbotapi.NewMessage(ChatID, "Thanks for playing, "+User.String()+"!  You collected "+strconv.Itoa(value.Points)+" cards."))
		log.Printf("Removing %v from the game %v...", User, ChatID)
		for i := 0; i < len(bot.CurrentGames[ChatID].CardTzarOrder); i++ {
			if bot.CurrentGames[ChatID].CardTzarOrder[i] == User.ID {
				// This is a workaround because go assignments inside a map don't work.
				game := bot.CurrentGames[ChatID]
				game.CardTzarOrder = append(bot.CurrentGames[ChatID].CardTzarOrder[:i], bot.CurrentGames[ChatID].CardTzarOrder[i+1:]...)
				bot.CurrentGames[ChatID] = game
				break
			}
		}
		if bot.CurrentGames[ChatID].Players[User.ID].IsCardTzar {
			bot.MoveCardTzar(ChatID)
		}
		delete(bot.CurrentGames[ChatID].Players, User.ID)
		if len(bot.CurrentGames[ChatID].Players) == 0 {
			log.Printf("There are no more players in game %v.  We shall end it.", ChatID)
			bot.SendMessage(tgbotapi.NewMessage(ChatID, "There are no more people playing in this game. We are going to end it."))
			bot.StopGame(ChatID)
		}
	} else {
		bot.SendMessage(tgbotapi.NewMessage(ChatID, User.String()+" is not playing yet.  Use command '/add' to add yourself."))
	}
}

func (bot *CAHBot) MoveCardTzar(ChatID int) {
	Game := bot.CurrentGames[ChatID]
	Game.CardTzarIndex = (Game.CardTzarIndex + 1) % len(Game.CardTzarOrder)
	bot.CurrentGames[ChatID] = Game
}
