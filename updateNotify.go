package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

// databaseUpdate receives a message from the frontend WebSocket connection in
//form of:
//		"SKIP"
//		"NEXTGAME"
//		"cardType cardClicked"
//Mostly commonly, the third message listed third will be sent by client
//connections, this will use the key(cardType)-value(cardClicked) pair to
//update the redis database accordingly and notify all users who are
//listening in that game
func databaseUpdate(user User, message string) {

	game := activeGames[user.GameID]

	//mutex lock
	game.mutex.Lock()

	//Skip turn button clicked
	if message == "SKIP" {
		skipTurn(user.GameID)

		//early unlock when turn skip called
		game.mutex.Unlock()
		return
	}

	//Next Game button clicked
	if message == "NEXTGAME" {
		nextGame(user.GameID)

		//early unlock when turn skip called
		game.mutex.Unlock()
		return
	}

	log.Println("Attempting card selection for:", message)

	//split message into [key, value]
	keyValue := strings.Split(message, " ")
	if len(keyValue)-1 != 1 {
		keyValue[1] = keyValue[1] + " " + keyValue[len(keyValue)-1]
	}

	//alter cardValue in database
	alterResult := alterCardState(user.GameID, keyValue)
	if alterResult == "" {
		return
	}

	turn := database.HGet(ctx, user.GameID, "turn").Val()

	//generate GameState object to be passed to user
	var gameState GameState
	gameState.GameID = user.GameID

	//database call to change score & turn and mark card as selected
	pipeline := database.TxPipeline()
	defer pipeline.Close() //close pipeline on failure
	var err error

	// changing score; civilian cards alter no points, and so that case will
	// simply fallthrough to change the turn
	gameState.RedScore, err = strconv.Atoi(database.HGet(ctx, user.GameID, "redScore").Val())
	if err != nil {
		log.Println("FAILURE: red score was not understood:", err)
		return
	}
	gameState.BlueScore, err = strconv.Atoi(database.HGet(ctx, user.GameID, "blueScore").Val())
	if err != nil {
		log.Println("FAILURE: blue score was not understood:", err)
		return
	}
	switch keyValue[0] {
	case "red":
		gameState.RedScore--
		if gameState.RedScore == 0 {
			gameState.GameOver = "true"
			pipeline.Do(ctx, "HSET", user.GameID, "gameover", "true")
		}
		pipeline.Do(ctx, "HSET", user.GameID, "redScore", gameState.RedScore)

	case "blue":
		gameState.BlueScore--
		if gameState.BlueScore == 0 {
			gameState.GameOver = "true"
			pipeline.Do(ctx, "HSET", user.GameID, "gameover", "true")
		}
		pipeline.Do(ctx, "HSET", user.GameID, "blueScore", gameState.BlueScore)

	case "assassin":
		gameState.GameOver = "true"
		pipeline.Do(ctx, "HSET", user.GameID, "gameover", "true")
	}

	// turn change only if card colour did not match turn colour
	if turn != keyValue[0] {
		switch turn {
		case "red":
			turn = "blue"
			gameState.Turn = "blue"
		case "blue":
			turn = "red"
			gameState.Turn = "red"
		}
		pipeline.Do(ctx, "HSET", user.GameID, "turn", turn)
		log.Println("turned changed to:", turn)
	} else {
		gameState.Turn = keyValue[0]
		log.Println("turned remained as:", turn)
	}

	// mark card as selected in game state
	pipeline.Do(ctx, "HSET", user.GameID, keyValue[0], alterResult)
	gameState.LastSelection = keyValue[1]

	//execute pipelined commands
	pipeline.Exec(ctx)

	//notify players about selection
	notify(user.GameID, gameState)

	//unlock mutex and set lock to available
	game.mutex.Unlock()
	log.Println("SUCCESS: card:", keyValue[0], keyValue[1], "was selected")
}

// skipTurn is called when clients send "NEXTGAME" messages through their
//WebSocket connection, it will create a new game board and return that new
//game information to all listeners
func nextGame(gameID string) {
	// generate words and place them into database object ---------------------

	// draw random cards
	words := drawCards()

	// select who goes first
	// who goes first determines how many cards a team needs to guess
	turns := [2]string{"red", "blue"}
	turn := turns[rand.Intn(2)]

	redScore := 9
	blueScore := 9

	if turn == "blue" {
		redScore--
	} else {
		blueScore--
	}

	// create strings for vals interface
	redCards := words[:redScore]
	blueCards := words[redScore : redScore+blueScore]
	civCards := words[redScore+blueScore : len(words)-1]
	assassin := words[len(words)-1]

	vals := map[string]interface{}{
		"blueScore": blueScore,
		"redScore":  redScore,
		"turn":      turn,
		"red":       strings.Join(redCards, " "),
		"blue":      strings.Join(blueCards, " "),
		"assassin":  assassin,
		"civilian":  strings.Join(civCards, " "),
		"gameover":  "false",
	}
	database.HSet(ctx, gameID, vals)

	get := database.HGetAll(ctx, gameID)
	if err := get.Err(); err != nil {
		if err == redis.Nil {
			log.Println("key does not exists")
		}
		panic(err)
	}
	log.Println("SUCCESS: created new game for ID:", gameID)

	notify(gameID, vals)
}

// skipTurn is called when clients send "SKIP" messages through their WebSocket
//connection, it will send a message to all listeners that the turn has been
//skipped.
func skipTurn(gameID string) {
	currentTurn := database.HGet(ctx, gameID, "turn").Val()
	switch currentTurn {
	case "red":
		currentTurn = "blue"
	case "blue":
		currentTurn = "red"
	}

	database.HSet(ctx, gameID, "turn", currentTurn)
	var turnStatus TurnState
	turnStatus.Turn = currentTurn

	notify(gameID, turnStatus)

	log.Println("turned skipped to:", currentTurn)
}

// notify will be called after a database entry has been updated following a
//slection on a card. This function will then notify all listeners on a game
//that a card has been selected
//
// Messages sent to WebSockets will be sent as JSON objects as such:
// {
// 	"gameId": string,
// 	"lastSelection": string,
// 	"redScore": int,
// 	"blueScore": int,
// 	"turn": string,
// 	"gameover": bool
// }
//
func notify(gameID string, status interface{}) {
	game := activeGames[gameID]
	outboud, err := json.Marshal(status)
	if err != nil {
		log.Println("Error encoding outbound message:", err)
	}
	for _, user := range game.Connections {
		user.Connection.WriteMessage(1, outboud)
	}
	log.Println("Sent:", string(outboud), "to client connections")
	return
}

// alterCardState will search through a cardType for that entry and mark it as
//chosen by adding a ! to the beginning of the word
func alterCardState(gameID string, keyValue []string) string {
	valuesFromKey := database.HGet(ctx, gameID, keyValue[0])
	if err := valuesFromKey.Err(); err != nil {
		if err == redis.Nil {
			log.Println("key does not exists")
		}
		log.Println("key does not exists")
		return ""
	}

	//replace cardValue with !cardValue for database insertion
	return strings.Replace(valuesFromKey.Val(), keyValue[1]+" ", "!"+keyValue[1]+" ", 1)
}
