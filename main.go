package main

import (
	"cahbot/secrets"
	"cahbot/tgbotapi"
	"log"
)

func main() {
	bot, err := NewCAHBot(secrets.Token)
	if err != nil {
		log.Panic(err)
	}
	defer bot.db_conn.Close()

	// Remove when deployed
	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.UpdatesChan(u)

	for update := range updates {
		go bot.HandleUpdate(&update)
	}
}
