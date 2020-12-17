# CMPT315_Project_Group4
##### Shea Odland, Von Castro, Eric Wedemire
##### CMPT315 - Web Application Development
##### Group Project: Codenames


## Configuration
```
    HOST = localhost
    PORT = 6379
```
## Database

We used a Docker based redis database for our implementation. Initially it was spun up on WSL with the command:
```
    sudo docker run --name projectDB -p 6379:6379 -d redis
```

## Functionality

#### Server-bound messages (client->server):
When creating a new game, the server expects the body of a request to include the requested name of the new game formatted in JSON as:
```javascript
    {"gameID": createGameBody}
```
after checking if the gameID is available, the server will reply with
either:
```
    1. HTTP.StatusOK
    2. "FAILURE: Game ID: REQUESTED_ID already exists"
        &
       HTTP.StatusBadRequest
```

After a successful call to create a game is made, the client is expected to make a subsequent connection request through the WebSocket API to:
```
    ws://FULLHOST/games?id=GAME_ID
```

When clients are sending card selections to the server, they are expected to be given as a space-seperated string in the form of:
```golang
   "cardType cardWord"
where:
    cardType = "red"|"blue"|"civilian"|"assassin"
    cardWord = any single word string  
```

If a skip command is given, the server expects the following message:
```golang
    "SKIP"
```
#### Client-bound messages (server->client):
If a client attempts to create a WebSocket to game that does not exist, the server will notify the client with the following message:
```golang
    {
        "status": "404 No current game called GAMEID"
    }
```

When the server is updating a client after a card selection has been made, clients can expect information back in the form of JSON with values as follows:
```javascript
    {
        "gameId": string,
        "lastSelection": string,
        "redScore": int,
        "blueScore": int,
        "turn": string,
        "gameover": "false"|"true"
    } 
```

If a turn has been skipped, the client side WebSocket will receive a response message formatted as a single value JSON:
```javascript
    {
        "turn": string,
    }

```

Upon joining a game with a WebSocket connection, the server with reply with the full game state of all relevant information structured as:
```javascript
    {
        "assassin": string,
        "blue": space-seperated words,
        "civilian": space-seperated words,
        "red": space-seperated words,
        "blueScore": int,
        "redScore": int,
        "turn": "red"|"blue"
        "gameover": "false"|"true"
    }
```

Words for each type will be structured as:
```javascript
    "africa !agent !air alien amazon"
```
where each word is seperated by a single space and those words that have been previously selected are denoted by a ! at the beginning of the word
