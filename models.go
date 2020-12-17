/*
 * Shea Odland, Von Castro, Eric Wedemire
 * CMPT315
 * Group Project: Codenames
 */

package main

import (
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

// User struct will
type User struct {
	GameID     string          `json:"gameId"`
	Connection *websocket.Conn `json:"connection,omitempty"`
}

// Game maintains the set of active clients and broadcasts messages to the
//clients.
type Game struct {
	GameID      string `json:"gameId"`
	Connections []User `json:"connections,omitempty"`
	mutex       sync.Mutex
}

// GameState is used to pass updates to client connections everytime a word is
//selected.
type GameState struct {
	GameID        string `json:"gameId"`
	LastSelection string `json:"lastSelection"`
	RedScore      int    `json:"redScore"`
	BlueScore     int    `json:"blueScore"`
	Turn          string `json:"turn"`
	GameOver      string `json:"gameover"`
}

// TurnState is used to pass updates to client connections everytime a word is
//selected.
type TurnState struct {
	Turn string `json:"turn"`
}

// ErrorState is used to notify socket connections that a game is not found.
type ErrorState struct {
	Status string `json:"status"`
}

// activeGames keeps a record of all active game sessions
var activeGames map[string]*Game = make(map[string]*Game)

// constants for DB connection
const (
	DBHOST = "localhost"
	DBPORT = "6379"
)

// constants for web hosting
const (
	WEBHOST = "localhost"
	WEBPORT = "8008"
)

// FULLHOST const for ease of use when using full URI path
const FULLHOST = WEBHOST + ":" + WEBPORT

//package-private access to database connection
var database *redis.Client

//upgrader params for upgrading HTTP to WebSocket Connections
var upgrader = websocket.Upgrader{}
