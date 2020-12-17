/**
 * Shea Odland, Von Castro, Eric Wedemire
 * CMPT315
 * Group Project: Codenames
 */
var socket: WebSocket;

const path = window.location.pathname;
const id = path.split("/")[2]

socket = new WebSocket("ws://localhost:8008/games?id=" + id);

socket.addEventListener('message', function (event) {
    let gameData = JSON.parse(event.data);
    const board: HTMLDivElement | null = document.querySelector('.board');
    if (gameData.status) {
        // redirect to 404 page
        window.location.assign("/notfound")
    }   
    //joining game, assassin field only appears here
    if (gameData.assassin) {
        dealCards(gameData);
        updateScoreboard(gameData);
        attachListeners();
    }
    //server sends update after card selection
    if (gameData.blueScore || gameData.redScore) {
        updateView(gameData);
        updateScoreboard(gameData)
    }
    assignTurn(gameData);
    checkGameState(gameData);
});

function nextGame() {
    socket.send('NEXTGAME');
}

function skipTurn() {
    socket.send('SKIP');
}

function dealCards(gameData: any) {
    let cards: any[] = [];
    let cardTypes = ["assassin", "civilian", "red", "blue"]
    for (let [key, value] of Object.entries(gameData)) {
        if (cardTypes.includes(key)) {
            let valueString = value + ""
            let words = valueString.split(" ");
            for (let i = 0; i < words.length;) {
                let selected = "unselected";
                if (words[i].includes("!")) {
                    // remove exclamation point
                    words[i] = words[i].slice(1);
                    selected = "selected"
                }
                if (Number.isInteger(Number(words[i + 1]))) {
                    let card = { word: words[i], wordCategory: key, position: words[i + 1], status: selected }
                    cards.push(card)
                    i += 2
                } else {
                    let card = { word: words[i] + " " + words[i + 1], wordCategory: key, position: words[i + 2], status: selected }
                    cards.push(card)
                    i += 3
                }
            }
        }
    }
    assignWords(cards.sort((a, b) => (Number(a.position) > Number(b.position)) ? 1 : -1), gameData);
}

function assignWords(cards: object[], gameData: any) {
    const tmpl = document.querySelector("#board-template").innerHTML;
    if (tmpl && cards.length != 0) {
        const renderFn = doT.template(tmpl);
        const renderResult = renderFn({ "cards": cards });
        document.querySelector(".board").innerHTML = renderResult;
        let wordCards = document.querySelectorAll(".board .wordCard");
        wordCards.forEach(function (wordCard) {
            let element = <HTMLElement>wordCard;
                        
            //add hover styling
            element.classList.add("playerView")
            if (element.classList[3] == "selected") {
                alterCardStyle(element);

            } else {
                if (!gameData.assassin.includes("!")) {
                    element.addEventListener("click", checkCard);
                }
            }
        });
    }
}

function checkGameState(gameData: any) {
    // check if game has ended
    if (gameData.gameover === "true") {
        let scoreboard = document.querySelector(".player-turn");
        // check if blue won
        if (gameData.blueScore <= 0) {
            if (scoreboard) {
                scoreboard.innerHTML = "Victory for Blue!";
            }
        }
        // check if red won
        else if (gameData.redScore <= 0) {
            if (scoreboard) {
                scoreboard.innerHTML = "Victory for Red!";
            }
        }
        // assassin probably clicked
        else {
            if (scoreboard) {
                let winner = scoreboard.innerHTML.slice(0, -7);
                scoreboard.innerHTML = "Victory for " + winner;
            }
        }
        // remove listener on skip button
        let skipButton = document.querySelector("#btn-skip-turn");
        if (skipButton) {
            skipButton.removeEventListener("click", skipTurn);
        }
        // remove listeners on cards
        let wordCards = document.querySelectorAll(".board .wordCard.tile");
        wordCards.forEach(function (wordCard) {
            let element = <HTMLDivElement>wordCard;
            element.removeEventListener("click", checkCard);
        });
    }
}

function assignTurn(gameData: any) {
    let turn = document.querySelector(".player-turn")
    if (turn) {
        let turnString = gameData.turn + ""
        turn.innerHTML = turnString[0].toUpperCase() + turnString.slice(1) + "'s turn"
    }
}

function updateScoreboard(gameData: any) {
    // bug: scoreboard is undefined until a card selection event is triggered
    let redBoard = document.querySelector(".red-scoreboard");
    if (redBoard) { redBoard.innerHTML = gameData.redScore };
    let blueBoard = document.querySelector(".blue-scoreboard");
    if (blueBoard) { blueBoard.innerHTML = gameData.blueScore };
}

/* This function is written with the premise that word cards will be made up of 
* three classes 'word-card unselected color(blue or red)'. If a word card is 
* unselected, it will be beige. Once selected, the function does an assassin check, 
* then changes the unselected class to selected, activating the card's color. 
*/
function checkCard(event: MouseEvent) {
    // Grab div clicked
    let card = event.currentTarget as HTMLElement;

    // Format card to "cardType cardWord" so the API can understand and respond appropriately
    let cardType = card.classList[2];
    let cardWord = card.textContent;
    let cardSelection = cardType + " " + cardWord;
    // card.classList.remove("unselected")
    // card.classList.add("selected")
    // Send the card selected to the backend to be marked selected
    socket.send(cardSelection);
}

function updateView(gameData: any) {
    let lastSelection = gameData.lastSelection;
    let wordCards = document.querySelectorAll(".wordCard");
    wordCards.forEach(function (wordCard) {
        let element = <HTMLElement>wordCard;
        if (element.innerHTML == lastSelection) {
            // update view for everyone
            element.classList.remove("unselected")
            element.classList.add("selected")
            alterCardStyle(element);
            element.removeEventListener("click", checkCard);
        }
    });

}

function spyMasterView() {
    let cards: NodeListOf<HTMLElement> = document.querySelectorAll(".board .wordCard");
    cards.forEach(function (card: HTMLElement) {

        let cardClasses = card.classList;
        card.classList.remove("playerView");
        card.removeEventListener("click", checkCard);

        if (cardClasses[0] && cardClasses[1] != "assassin") {
            if (cardClasses[2] == "blue")  {
                card.classList.add("spymasterBlue")
            } else if (cardClasses[2] == "red") {
                card.classList.add("spymasterRed")
            } else if (cardClasses[2] == "civilian") {
                card.classList.add("spymasterCivilian")
            } else {
                card.classList.add("spymasterAssassin")
            }
        }
    });
}

function playerView() {
    let cards: NodeListOf<HTMLElement> = document.querySelectorAll(".board .wordCard");
    cards.forEach(function (card) {
        card.addEventListener("click", checkCard)
        card.classList.remove("spymasterBlue")
        card.classList.remove("spymasterRed")
        card.classList.remove("spymasterCivilian")
        card.classList.remove("spymasterAssassin")
        let cardClasses = card.classList;
        if (cardClasses[3] == "selected") {
            alterCardStyle(card);
        }
        else {
            //add hover css back
            card.classList.add("playerView")
        }
    });
}

// create element for link-container-template
// <script type="text/x-dot-template" id = "link-container-template" >
//     <div>Share this link with your friends: <a href="" class="link" > {{=it.link }}</a></div >
// </script>

function createLinkTemplate(linkTemplate: HTMLScriptElement) {
    let a = document.createElement("a");
    a.className = "link";
    a.textContent = "{{=it.link}}";

    let div = document.createElement("div");
    div.textContent = "Share this link with your friends";
    div.appendChild(a);

    linkTemplate.appendChild(div);
}

// create element for board-template
// <script type="text/x-dot-template" id = "board-template" >
//     {{ ~it.cards: value: index }}
//     <div class="{{=value["wordCategory"]}}" "tile{{=index+1}}" > {{=value["word"] }}</div>
//     { { ~} }
// </script>

function createBoardTemplate(boardTemplate: HTMLScriptElement): string {

    let div = document.createElement("div");
    div.className = "wordCard tile {{=value['wordCategory']}} {{=value['status']}}";
    div.id = "tile{{=index+1}}"
    div.textContent = '{{=value["word"]}}';
    boardTemplate.insertAdjacentText('afterbegin', '{{~it.cards:value:index}}');
    boardTemplate.appendChild(div);
    boardTemplate.insertAdjacentText('beforeend', '{{~}}');
    let gameBody = document.querySelector("body");
    if (gameBody) {
        let gameId = gameBody.id;
        return gameId;
    }
    return "";
}

function alterCardStyle(element: HTMLElement) {
    switch (element.classList[2]) {
        case "blue":
            element.classList.add("spymasterBlue")
            element.classList.remove("playerView")
            break;
        case "red":
            element.classList.add("spymasterRed")
            element.classList.remove("playerView")
            break;
        case "assassin":
            element.classList.add("spymasterAssassin")
            element.classList.remove("playerView")
            break;
        case "civilian":
            element.classList.add("spymasterCivilian")
            element.classList.remove("playerView")
            break;
        default:
            return;
    }
}


function attachListeners() {
    const spyBtn: HTMLInputElement | null = document.querySelector("#btn-spymaster");
    if (spyBtn) {
        spyBtn.addEventListener("click", spyMasterView);
    }
    const playBtn: HTMLInputElement | null = document.querySelector("#btn-player");
    if (playBtn) {
        playBtn.addEventListener("click", playerView);
    }
    const nxtBtn: HTMLInputElement | null = document.querySelector("#btn-next-game");
    if (nxtBtn) {
        nxtBtn.addEventListener("click", nextGame);
    }

    const skipBtn: HTMLInputElement | null = document.querySelector("#btn-skip-turn");
    if (skipBtn) {
        skipBtn.addEventListener("click", skipTurn);
    }
}

function createTemplates() {
    const linkTemp: HTMLScriptElement | null = document.querySelector("#link-container-template");
    if (linkTemp) { createLinkTemplate(linkTemp) }

    const boardTemp: HTMLScriptElement | null = document.querySelector("#board-template");
    if (boardTemp) {
        let gameId = createBoardTemplate(boardTemp);
        dealCards(gameId);
    }
}

createTemplates();
attachListeners();