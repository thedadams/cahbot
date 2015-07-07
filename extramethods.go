package main

import (
	"cahbot/tgbotapi"
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
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "There is no game being played here.  Use command '/create' to start a new one."))
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
	switch strings.ToLower(strings.Replace(strings.Fields(m.Text)[0], "/", "", 1)) {
	case "create":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			if bot.CurrentGames[m.Chat.ID].HasStarted {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "There is already a game created here.  Use command '/stop' to end the previous game."))
			} else {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "There is already a game created here.  Use command '/stop' to end the previous game or '/resume' to resume."))
			}
		} else {
			bot.CreateNewGame(m.Chat.ID, m.From)
		}
	case "start", "resume":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.StartGame(m.Chat.ID)
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
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			if bot.CurrentGames[m.Chat.ID].HasStarted {
				bot.PauseGame(m.Chat.ID)
			} else {
				bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "The current game is already paused.  Use command '/resume' to resume it."))
			}
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "join":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.AddPlayerToGame(m.Chat.ID, m.From)
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
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.ListCardsForUser(m.Chat.ID, m.From)
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "scores":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Here are the current scores:\n"+bot.CurrentGames[m.Chat.ID].Scores()))
		} else {
			bot.SendNoGameMessage(m.Chat.ID)
		}
	case "gamesettings":
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
	case "feedback":
		bot.ReceiveFeedback(m.Chat.ID)
	default:
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Sorry, I don't know that command."))
	}
}

// This method starts a new game.
func (bot *CAHBot) CreateNewGame(ChatID int, User tgbotapi.User) {
	log.Printf("Creating a new game for Chat ID %v.", ChatID)
	// Get the keys for the All Cards map.
	ShuffledCards := make([]int, len(bot.AllCards))
	for i := 0; i < len(ShuffledCards); i++ {
		ShuffledCards[i] = i
	}
	shuffle(ShuffledCards)
	bot.CurrentGames[ChatID] = CAHGame{ShuffledCards, map[int]PlayerGameInfo{User.ID: PlayerGameInfo{User, 0, make([]int, bot.CurrentGames[ChatID].Settings.NumCardsInHand), true, false, false}}, 0, GameSettings{false, false, false, 7}, false}
	log.Println("Game for Chat ID %v created successfully!", ChatID)
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "The game was created successfully."))
}

// This method starts an already created game.
func (bot *CAHBot) StartGame(ChatID int) {
	log.Printf("Starting game for Chat ID %v.", ChatID)
	// There is a bug in Go that does not allow for things like bot.CurrentGames[ChatID].HasStarted = true.  This is a workaround.
	tmp := bot.CurrentGames[ChatID]
	tmp.HasStarted = true
	bot.CurrentGames[ChatID] = tmp
	if DoWeHaveAllAnswers(bot.CurrentGames[ChatID].Players) {
		log.Printf("Asking the Card Tzar, %v, to pick the best and/or worse answer.", bot.CurrentGames[ChatID].Players[bot.CurrentGames[ChatID].CardTzarIndex])
		bot.AskCardTzarForChoice(ChatID, bot.CurrentGames[ChatID].Players[bot.CurrentGames[ChatID].CardTzarIndex])
	} else {
		for _, value := range bot.CurrentGames[ChatID].Players {
			if !value.IsCardTzar && value.WaitingForCard {
				log.Printf("Asking %v for an answer card.", value)
				bot.AskForCardFromPlayer(ChatID, value)
			}
		}
	}
}

// This method asks a player for a card.
func (bot *CAHBot) AskForCardFromPlayer(ChatID int, Player PlayerGameInfo) {

}

// This method asks the Card Tzar to make a choice.
func (bot *CAHBot) AskCardTzarForChoice(ChatID int, Player PlayerGameInfo) {

}

// This method pauses a started game.
func (bot *CAHBot) PauseGame(ChatID int) {
	log.Printf("Pausing game for Chat %v...", ChatID)
	// There is a bug in Go that does not allow for things like bot.CurrentGames[ChatID].HasStarted = false.  This is a workaround.
	tmp := bot.CurrentGames[ChatID]
	tmp.HasStarted = false
	bot.CurrentGames[ChatID] = tmp
}

// This method stops and ends an already created game.
func (bot *CAHBot) StopGame(ChatID int) {
	bot.SendMessage(tgbotapi.NewMessage(ChatID, "The game has been stopped.  Here are the scores:\n"+bot.CurrentGames[ChatID].Scores()+"Thanks for playing!"))
	log.Printf("Deleting a game with Chat ID %v...", ChatID)
	delete(bot.CurrentGames, ChatID)
}

func (bot *CAHBot) ListCardsForUser(ChatID int, User tgbotapi.User) {

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
func (bot *CAHBot) AddPlayerToGame(ChatID int, User tgbotapi.User) {
	if _, ok := bot.CurrentGames[ChatID].Players[User.ID]; ok {
		bot.SendMessage(tgbotapi.NewMessage(ChatID, User.String()+" is already playing.  Use command '/leave' to remove yourself."))
	} else {
		log.Printf("Adding %v to the game %v...", User, ChatID)
		bot.CurrentGames[ChatID].Players[User.ID] = PlayerGameInfo{User, 0, make([]int, bot.CurrentGames[ChatID].Settings.NumCardsInHand), false, false, false}
		bot.SendMessage(tgbotapi.NewMessage(ChatID, "Welcome to the game, "+User.String()+"!"))
	}
}

// Remove a player from a game if the player is playing.
func (bot *CAHBot) RemovePlayerFromGame(ChatID int, User tgbotapi.User) {
	if _, ok := bot.CurrentGames[ChatID].Players[User.ID]; ok {
		bot.SendMessage(tgbotapi.NewMessage(ChatID, "Thanks for playing, "+User.String()+"!  You collected "+strconv.Itoa(bot.CurrentGames[ChatID].Players[User.ID].Points)+"."))
		log.Printf("Removing %v from the game %v...", User, ChatID)
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
