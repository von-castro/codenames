package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

// newSocketConnection connects to a front end websocket to pass game data back
//and forth to. It expects http requests to made using as query in the form of
//games?id=GAMEID. Upon successful connection, the user to prompted that
//connection was successful
func newSocketConnection(writer http.ResponseWriter, request *http.Request) {
	var newUser User
	var err error

	// grabbing gameID
	vars := mux.Vars(request)
	gameID := vars["gameID"]

	//creating WebSocket
	log.Println("Attempting new client WebSocket creation for game:", gameID)
	newUser.Connection, err = upgrader.Upgrade(writer, request, nil)
	if err != nil {
		message := "Could not open websocket connection: " + err.Error()
		encodeAndSendError(writer, request, http.StatusBadRequest, message)
		return
	}
	newUser.GameID = gameID
	currentGame := activeGames[gameID]
	if currentGame == nil {
		var errorState ErrorState
		errorState.Status = "404 No current game called " + gameID
		message, err := json.Marshal(errorState)
		if err != nil {
			message := "Could not encode fail status into JSON:" + err.Error()
			encodeAndSendError(writer, request, http.StatusBadRequest, message)
			return
		}
		newUser.Connection.WriteMessage(1, message)
		//encodeAndSendError(writer, request, http.StatusBadRequest, message)
		return
	}

	currentGame.Connections = append(currentGame.Connections, newUser)

	//give current game state to new connection
	currentGameState, err := json.Marshal(database.HGetAll(ctx, gameID).Val())
	if err != nil {
		log.Println("Error encoding initial status message:", err)
		return
	}
	newUser.Connection.WriteMessage(1, currentGameState)

	//calling listener in go routine
	log.Println("SUCCESS: created socket:", gameID)
	go listenOnSocket(newUser)
}

// listenOnSocket takes a User struct as an argument and will read messages
//sent from the client connection. Messages are expected to be the key-value
//pair that was altered on card selection
//
// listen expects messages from the clients to be structured in one of two ways
// 		1. "SKIP" - this message is used when individuals have a loss of
//			as governed by the Codenames rules
//		2. "KEY VALUE" - this message is the most common, it signifies:
//				key
func listenOnSocket(user User) {
	for {
		messageType, message, err := user.Connection.ReadMessage()
		if err != nil || messageType != 1 {
			log.Println("Message:", message, "; not understood by server:", err)
			checkGameStatus(user)
			return
		}
		databaseUpdate(user, string(message))
	}
}

// checkGameStatus will run after a user has closed their browser window,
//ending their connection to a current game session. This function will also
//completely remove the remove a game from the map of active games and the
//database
func checkGameStatus(user User) {
	session := activeGames[user.GameID]

	//locking session mutex
	session.mutex.Lock()

	//close game as last connection is closing
	if len(session.Connections) <= 1 {
		log.Println("Last user left game: " + user.GameID + ". Tearing down game")
		del := database.Del(ctx, user.GameID)
		if err := del.Err(); err != nil {
			if err == redis.Nil {
				log.Println("key does not exists")
				return
			}
		}
		delete(activeGames, user.GameID)
		log.Println("Successfully eneded game (" + user.GameID + ")")
	} else {
		for i, element := range session.Connections {
			if element == user {
				l := len(session.Connections)
				session.Connections[l-1], session.Connections[i] = session.Connections[i], session.Connections[l-1]
				session.Connections = session.Connections[:l-1]
				log.Println("Removing user", user, "from game:", user.GameID)
				break
			}
		}
	}
	//unlocking session mutex
	session.mutex.Unlock()

	//close user connection
	user.Connection.Close()
}
