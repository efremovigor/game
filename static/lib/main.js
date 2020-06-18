var PIXI = require('pixi.js');

let app;
let player;
let socket;
let socketInfo = {
    x: 0,
    y: 0,
};
let keysPressed = {};

document.getElementById("greet").hidden = false;
document.getElementById("choose-multi").addEventListener('click', startGame);

function startGame() {
    document.getElementById("greet").hidden = true;
    document.getElementById("app").hidden = false;

    socket = new WebSocket('ws://127.0.0.1:3000/ws');
    socket.onmessage = function (event) {
        const messageText = event.data;
        const message = JSON.parse(messageText);
        console.log(message);
    };

    socket.onopen = function (e) {
        console.log("[open] Соединение установлено");
        console.log("Отправляем данные на сервер");
        socket.send('{"type":"newPlayer","payload":{"name":"test1"}}');
    };

    socket.onmessage = function (event) {
        let response = JSON.parse(event.data);
        if (response.type === 'SIGNAL_START_THE_GAME') {
            app = new PIXI.Application({
                width: response.conf.width,
                height: response.conf.height,
                backgroundColor: 0x1099bb,
                resolution: window.devicePixelRatio || 1,


            });
            document.getElementById("app").appendChild(app.view);
            player = PIXI.Sprite.from('static/bunny.png');
            player.anchor.set(0.5);
            player.x = 0;
            player.y = 0;
            app.stage.addChild(player);
            app.ticker.add(appLoop);
        }
        if (response.type === 'SIGNAL_INFO_THE_GAME') {
            socketInfo = response.info.player;
        }
    };

    socket.onclose = function (event) {
        if (event.wasClean) {
            console.log(`[close] Соединение закрыто чисто, код=${event.code} причина=${event.reason}`);
        } else {
            // например, сервер убил процесс или сеть недоступна
            // обычно в этом случае event.code 1006
            console.log('[close] Соединение прервано');
        }
    };

    socket.onerror = function (error) {
        console.log(`[error] ${error.message}`);
    };
}

function appLoop() {
    let x = socketInfo.x - player.x;
    let y = socketInfo.y - player.y;
    // player.x += x;
    // player.y += y;
    if (keysPressed['KeyD'] && keysPressed['KeyD'] === true) {
        player.x += x;
    }
    if (keysPressed['KeyS'] && keysPressed['KeyS'] === true) {
        player.y += y;
    }
    if (keysPressed['KeyW'] && keysPressed['KeyW'] === true) {
        player.y -= y;
    }
    if (keysPressed['KeyA'] && keysPressed['KeyA'] === true) {
        player.x -= x;

    }
}

document.addEventListener('keydown', (event) => {
    keysPressed[event.code] = true;
});

document.addEventListener('keyup', (event) => {
    delete keysPressed[event.code];
});

document.addEventListener('keydown', function (event) {
    let dir = '';
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

    socket.send('{"type":"command","payload":{"name":"' + dir + '"}}');

});