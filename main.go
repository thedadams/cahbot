package main

import (
    "cahbot/tgbotapi"
    "log"
)

func main() {
    bot, err := NewCAHBot(Token)
    if err != nil {
        log.Panic(err)
    }

    bot.Debug = true

    log.Printf("Authorized on account %s", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates, err := bot.UpdatesChan(u)

    for update := range updates {
        go bot.HandleUpdate(&update)
    }
}