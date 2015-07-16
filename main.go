package main

import (
	"cahbot/secrets"
	"cahbot/tgbotapi"
	"log"
	"strconv"
	"strings"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
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

	err = bot.UpdatesChan(u)
	c := time.Tick(1 * time.Minute)
	go func() {
		for _ = range c {
			log.Printf("Cleaning up old games.")
			tx, err := bot.db_conn.Begin()
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
				var gameid string
				var userid int
				if err := rows.Scan(&gameid); err != nil {
					log.Printf("ERROR: %v", err)
					log.Printf("Failed to clean up old games with id %v.", gameid)
					continue
				} else {
					gameid = strings.Replace(strings.Replace(gameid, "(", "", 1), ")", "", 1)
					userid, _ = strconv.Atoi(strings.Split(gameid, ",")[1])
					gameid = strings.Split(gameid, ",")[0]
					log.Printf("Game with id %v deleted.  Let user with id %v know about it.", gameid, userid)
					bot.SendMessage(tgbotapi.NewMessage(userid, "Your game has been deleted because of inactivity."))
				}
			}
			if err := rows.Err(); err != nil {
				log.Printf("ERROR: %v", err)
			}
			rows.Close()
			tx.Commit()
		}
	}()

	for update := range bot.Updates {
		go bot.HandleUpdate(&update)
	}
}
