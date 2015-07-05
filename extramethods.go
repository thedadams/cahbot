package main

import (
    "strings"
    "tgbotapi"
    "log"
)

// This creates a new CAHBot, which is basically a wrapper for tgbotapi.BotAPI.
// We need this wrapper to add the desired methods.
func NewCAHBot(token string) (*CAHBot, error) {
    GenericBot, err := tgbotapi.NewBotAPI(token)
    return &CAHBot{GenericBot}, err
}

// This is the starting point for handling an update from chat.
func(bot *CAHBot) HandleUpdate(update *tgbotapi.Update) {
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
    }
}

// This method starts a new game.
func (bot *CAHBot) StartNewGame(ChatID int) {
    log.Printf("Starting a new game.")
    bot.SendMessage(tgbotapi.NewMessage(ChatID, "I hear you would like to start a new game."))
}