/**
 * Shea Odland, Von Castro, Eric Wedemire
 * CMPT315
 * Group Project: Codenames
 */

function createGame() {
    let idInput = <HTMLInputElement>document.querySelector("#game-id");
    let gameId = idInput.value;
    const createGameBody: string = gameId;
    const jsonBody: string = JSON.stringify({ "gameID": createGameBody });
    const myInit: RequestInit = {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', },
        body: jsonBody,
    };

    const apiCall = new Request("http://localhost:8008/api/v1/games", myInit);
    fetch(apiCall)
        .then(response => {
            if (response.status === 400) {
            } window.location.assign('/games/' + gameId);
        });
}

const goForm: HTMLInputElement | null = document.querySelector("#form");
if (goForm) {
    goForm.addEventListener("submit", createGame);
};