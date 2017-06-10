package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/thedadams/telegram-bot-api"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	bot, err := NewCAHBot(Token)
	if err != nil {
		log.Panic(err)
	}
	defer bot.DBConn.Close()

	// Remove when deployed
	// bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	c := time.Tick(60 * time.Minute)
	go func() {
		for _ = range c {
			log.Printf("Cleaning up old games.")
			tx, err := bot.DBConn.Begin()
			if err != nil {
				log.Printf("ERROR: %v", err)
				log.Printf("Failed to clean up old games.")
				tx.Rollback()
				continue
			}
			rows, err := tx.Query("SELECT clean_up_old_games()")
			if err != nil {
				log.Printf("ERROR: %v", err)
				log.Printf("Failed to clean up old games.")
				tx.Rollback()
				rows.Close()
				continue
			}
			for rows.Next() {
				var GameID string
				var UserID int64
				if err := rows.Scan(&GameID); err != nil {
					log.Printf("ERROR: %v", err)
					log.Printf("Failed to clean up old games with id %v.", GameID)
					continue
				} else {
					GameID = strings.Replace(strings.Replace(GameID, "(", "", 1), ")", "", 1)
					UserID, _ = strconv.ParseInt(strings.Split(GameID, ",")[1], 10, 64)
					GameID = strings.Split(GameID, ",")[0]
					log.Printf("Game with id %v deleted.  Let user with id %v know about it.", GameID, UserID)
					bot.Send(tgbotapi.NewMessage(UserID, "Your game has been deleted because of inactivity."))
				}
			}
			if err := rows.Err(); err != nil {
				log.Printf("ERROR: %v", err)
			}
			rows.Close()
			tx.Commit()
		}
	}()

	for update := range updates {
		go bot.HandleUpdate(&update)
	}
}
