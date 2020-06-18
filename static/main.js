document.getElementById("greet").hidden = false;

document.getElementById("choose-multi").addEventListener('click', startGame);

function startGame() {
    document.getElementById("greet").hidden = true;
    document.getElementById("game").hidden = false;
    document.getElementById("game-canvas").hidden = false;

    game.socket = new WebSocket('ws://127.0.0.1:3000/ws');
    game.socket.onmessage = function (event) {
        const messageText = event.data;
        const message = JSON.parse(messageText);
        console.log(message);
    };

    game.socket.onopen = function (e) {
        console.log("[open] Соединение установлено");
        console.log("Отправляем данные на сервер");
        game.socket.send('{"type":"newPlayer","payload":{"name":"test1"}}');
    };

    game.socket.onmessage = function (event) {
        console.log(`[message] Данные получены с сервера: ${event.data}`);
        let response = JSON.parse(event.data);
        if (response.type === 'SIGNAL_START_THE_GAME') {
            game.canvas.width = response.conf.width;
            game.canvas.height = response.conf.height;
            game.interval = setInterval(() => gameLoop(), 1 / 60 * 1000);
        }
        if (response.type === 'SIGNAL_INFO_THE_GAME') {
            game.player.h = response.info.player.h;
            game.player.w = response.info.player.w;
            player = response.info.player

        }
    };

    game.socket.onclose = function (event) {
        if (event.wasClean) {
            console.log(`[close] Соединение закрыто чисто, код=${event.code} причина=${event.reason}`);
        } else {
            // например, сервер убил процесс или сеть недоступна
            // обычно в этом случае event.code 1006
            console.log('[close] Соединение прервано');
        }
    };

    game.socket.onerror = function (error) {
        console.log(`[error] ${error.message}`);
    };


}

let game = {};
game.canvas = document.getElementById("game-canvas");
game.ctx = game.canvas.getContext("2d");
game.canvas.width = 0;
game.canvas.height = 0;
game.player = {
    x: 0,
    y: 0,
    w: 0,
    h: 0,
};
let player = {
    x: 0,
    y: 0,
    w: 0,
    h: 0,
};

function gameLoop() {

    if (game.player.x < player.x) {
        game.player.x++
    }
    if (game.player.x > player.x) {
        game.player.x--
    }
    if (game.player.y < player.y) {
        game.player.y++
    }
    if (game.player.y > player.y) {
        game.player.y--
    }
    console.log(game.player);

    game.ctx.clearRect(0, 0, game.canvas.width, game.canvas.height);
    game.ctx.fillStyle = "silver";
    game.ctx.fillRect(0, 0, game.canvas.width, game.canvas.height);
    game.ctx.fillStyle = "green";
    game.ctx.fillRect(game.player.x, game.player.y, game.player.w, game.player.h);
}

let keysPressed = {};

document.addEventListener('keydown', (event) => {
    keysPressed[event.code] = true;
});

document.addEventListener('keyup', (event) => {
    delete keysPressed[event.code];
});

document.addEventListener('keydown', function (event) {
    let dir = '';
    console.log(keysPressed);

    if (keysPressed['KeyD'] && keysPressed['KeyD'] === true) {
        dir = 'right';
    }
    if (keysPressed['KeyS'] && keysPressed['KeyS'] === true) {
        if (dir !== '') {
            dir += '-'
        }
        dir += 'down';
    }
    if (keysPressed['KeyW'] && keysPressed['KeyW'] === true) {
        if (dir !== '') {
            dir += '-'
        }
        dir += 'up';
    }
    if (keysPressed['KeyA'] && keysPressed['KeyA'] === true) {
        if (dir !== '') {
            dir += '-'
        }
        dir += 'left';
    }
    if (dir === '') {
        return;
    }
    console.log('{"type":"command","payload":{"name":"' + dir + '"}}');

    game.socket.send('{"type":"command","payload":{"name":"' + dir + '"}}');

});