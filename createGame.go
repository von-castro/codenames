/*
 * Shea Odland, Von Castro, Eric Wedemire
 * CMPT315
 * Group Project: Codenames
 */

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type templateData struct {
	ID string
}

// createGame will generate a Game struct to store all subsequent connections
//to the game. createGame will always check if a game of the same name exists
//before allowing creation
func createGame(writer http.ResponseWriter, request *http.Request) {
	var err error
	var newGame Game

	// grabbing gameID
	err = json.NewDecoder(request.Body).Decode(&newGame)
	if err != nil {
		message := "Failure to decode client JSON: " + err.Error()
		encodeAndSendError(writer, request, http.StatusBadRequest, message)
		return
	}
	log.Println("Attempting game creation for ID:", newGame.GameID)
	//set game name for new game and add it to map of all active games
	if activeGames[newGame.GameID] == nil {
		activeGames[newGame.GameID] = &newGame
	} else {
		message := "FAILURE: Game ID: " + newGame.GameID + " already exists"
		encodeAndSendError(writer, request, http.StatusBadRequest, message)
		return
	}
	// generate words and place them into database object ---------------------

	// draw n number or random cards
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
	database.HSet(ctx, newGame.GameID, vals)

	get := database.HGetAll(ctx, newGame.GameID)
	if err := get.Err(); err != nil {
		if err == redis.Nil {
			log.Println("key does not exists")
		}
		panic(err)
	}
	log.Println("SUCCESS: created game:", newGame.GameID)
}

func (t *tempHandler) passTemplate(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	tdata := templateData{
		ID: vars["id"],
	}
	err := t.temp.ExecuteTemplate(w, "game.tmpl", tdata)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func drawCards() []string {
	dat, err := ioutil.ReadFile("./wordlist.txt")
	if err != nil {
		panic(err)
	}

	cards := make([]string, 25)
	words := strings.Split(string(dat), "\n")
	for i := 0; i < 25; i++ {
		// select random word and place in cards at index i
		j := rand.Intn(len(words))
		cards[i] = strings.TrimSuffix(words[j], "\r") + " " + strconv.Itoa(i+1)

		// remove selected word from words list
		words[j] = words[len(words)-1]
		words[len(words)-1] = ""
		words = words[:len(words)-1]
	}
	shuffleCards(cards)
	return cards
}

func shuffleCards(cards []string) []string {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cards), func(i, j int) { cards[i], cards[j] = cards[j], cards[i] })
	return cards
}
