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

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// This is the code that we will use to write the bot to file.
	// Need to figure out how to interrupt the update loop to do this safely.
	/*fileJson, _ := json.Marshal(generic)
	  err := ioutil.WriteFile("output.json", fileJson, 0644)
	  if err != nil {
	      fmt.Printf("WriteFileJson ERROR: %+v", err)
	  }*/

	updates, err := bot.UpdatesChan(u)

	for update := range updates {
		go bot.HandleUpdate(&update)
	}
}
