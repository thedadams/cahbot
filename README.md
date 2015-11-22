# CAH Bot
A Telegram bot, written in Go, for playing Cards Against Humanity.  This is very much still a work in progress.  Right now, the following commands are supported:

- /create -- Create a game.  This adds the person that invoked this action to the game.
- /join -- The user that invokes this action is added to the game, if there is one.
- /leave -- The user that invokes this action is removed from the game.
- /start -- Start a game.  Should be invoked after everyone is added.
- /stop -- Ends a game.  Also invoked if everyone leaves a game.
- /scores -- List the scores for the game, if there is one.
- /gamesettings -- List the settings for the game.
- /whoistczar -- Sends a message that reveals who the Card Tzar is.
- /mycards -- Shows the user the cards they are "holding."

The following commands are in progress:
- /changesettings -- Change the settings of the current game.

The Telegram Bot functionality comes from [Telegram Bot API](https://github.com/thedadams/telegram-bot-api), another of my repositories.  Most of the bot functionality is complete; the game play is left to code.  For example, starting a game doesn't actually start the game.

The card data was taken from https://github.com/samurailink3/hangouts-against-humanity, which is offered under the [Creative Commons Attribution-NonCommercial-ShareAlike 3.0 Unported License](http://creativecommons.org/licenses/by-nc-sa/3.0/deed.en_US).  The card data remains under that license.

The code written here licensed under the [MIT license](LICENSE).
