package main

import (
	"cahbot/tgbotapi"
	"log"
	"strings"
)

// This is the starting point for handling an update from chat.
func (bot *CAHBot) HandleUpdate(update *tgbotapi.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	bot.DetectKindMessageRecieved(&update.Message)
}

// Here we detect the kind of message we received from the user.
func (bot *CAHBot) DetectKindMessageRecieved(m *tgbotapi.Message) string {
	if m.Text != "" {
		if strings.HasPrefix(m.Text, "/") {
			bot.ProccessCommand(m)
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
	switch strings.Replace(strings.Fields(m.Text)[0], "/", "", 1) {
	case "start":
		bot.StartNewGame(m.Chat.ID)
	case "stop":
		bot.StopGame(m.Chat.ID)
	case "join":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.AddPlayerToGame(m.Chat.ID, m.From)
		} else {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "There is no game being played here.  Use command '/start' to start a new one."))
		}
	case "leave":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.RemovePlayerFromGame(m.Chat.ID, m.From)
		} else {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "There is no game being played here.  Use command '/start' to start a new one."))
		}
	case "mycards":
		bot.ListCardsForUser(m.Chat.ID, m.From)
	case "scores":
		if _, ok := bot.CurrentGames[m.Chat.ID]; ok {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Here are the current scores:\n"+bot.CurrentGames[m.Chat.ID].Scores()))
		} else {
			bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "There is no game in this chat."))
		}
	case "gamesettings":
		bot.SendGameSettings(m.Chat.ID)
	case "changesettings":
		bot.ChangeGameSettings(m.Chat.ID)
	case "feedback":
		bot.ReceiveFeedback()
	default:
		bot.SendMessage(tgbotapi.NewMessage(m.Chat.ID, "Sorry, I don't know that command."))
	}
}

// This method starts a new game.
func (bot *CAHBot) StartNewGame(ChatID int) {
	// If there is already a game being played, we do not create another one.
	if _, ok := bot.CurrentGames[ChatID]; ok {
		bot.SendMessage(tgbotapi.NewMessage(ChatID), "There is already a game being player here.  Use command '/stop' to end the previous game.")
	} else {
		log.Println("Starting a new game for Chat ID " + ChatID + ".")
		// Get the keys for the All Cards map.
		ShuffledCards := make([]int, len(bot.AllCards))
		for i := 0; i < len(ShuffledCards); i++ {
			ShuffledCards[i] = i
		}
		bot.CurrentGames = append(bot.CurrentGames, CAHGame{shuffle(ShuffledCards), GetAllPlayersInChat(), 0, GameSettings{false, false, false, 7}})
		log.Println("Game for Chat ID " + ChatID + " created successfully!")
		bot.SendMessage(tgbotapi.NewMessage(ChatID, "The game was created successfully."))
	}

}

func (bot *CAHBot) StopGame(ChatID int) {

	if _, ok := bot.CurrentGames[ChatID]; ok {
		bot.setMessage(tgbotapi.NewMessage(ChatID, "The game has been stopped.  The winner was "+bot.CurrentGames[ChatID].Scores()+"."))
		log.Printf("Deleting a game with Chat ID " + ChatID + "...")
		delete(bot.CurrentGames, ChatID)
	} else {
		bot.SendMessage(tgbotapi.NewMessage(ChatID, "There is no game currently running in this chat.  Use command '/start' to start one."))
	}
}

func (bot *CAHBot) ListCardsForUser(ChatID int, User User) {

}

func (bot *CAHBot) SendGameSettings(ChatID int) {

}

func (bot *CAHBot) ChangeGameSettings(ChatID int) {

}

func (bot *CAHBot) ReceiveFeedback(ChatID int) {

}

// Add a player to a game if the player is not playing.
func (bot *CAHBot) AddPlayerToGame(ChatID int, User tgbotapi.User) {
	if _, ok := bot.CurrentGames[m.Chat.ID].Players[User.ID]; ok {
		bot.SendMessage(ChatID, User.String()+" is already playing.  Use command '/leave' to remove yourself.")
	} else {
		log.Printf("Adding %v to the game %v...", User, ChatID)
		bot.CurrentGames[ChatID].Players[User.ID] = PlayerGameInfo{User, 0, make([]int, bot.CurrentGames.SettingsNumCardsInHand, false, false)}
		bot.SendMessage(m.Chat.ID, "Welcome to the game, "+User.String()+"!")
	}
}

// Remove a player from a game if the player is playing.
func (bot *CAHBot) RemovePlayerFromGame(ChatID int, User tgbotapi.User) {
	if _, ok := bot.CurrentGames[m.Chat.ID].Players[User.ID]; ok {
		log.Printf("Removing %v from the game %v...", User, ChatID)
		delete(bot.CurrentGames[ChatID].Players, User.ID)
		bot.SendMessage(m.Chat.ID, "Thanks for playing, "+User.String()+"!")
	} else {
		bot.SendMessage(ChatID, User.String()+" is not playing yet.  Use command '/add' to add yourself.")
	}
}
