let PIXI = require('pixi.js');
let loader = PIXI.Loader.shared;
loader.add('bunny', 'static/bunny.png');
loader.load();

let app;
let player;
let playerSocketInfo = {
    x: 0,
    y: 0,
};
let otherPlayers = {};
let otherPlayerSocketInfo = {};
let socket;

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
            player = PIXI.Sprite.from(loader.resources['bunny'].texture);
            player.anchor.set(0.5);
            player.x = 0;
            player.y = 0;
            app.stage.addChild(player);
            app.ticker.add(appLoop);
        }
        if (response.type === 'SIGNAL_INFO_THE_GAME') {
            playerSocketInfo = response.info.player;
            otherPlayerSocketInfo = response.info.others;
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
setInterval(function () {

    for (let [key, value] of Object.entries(otherPlayers)) {
        if(value.getLastTime + 1500 < Date.now()){
            console.log('remove'+key);
            app.stage.removeChild(otherPlayers[key]);
            delete otherPlayers[key];
        }
        console.log('count:'+Object.entries(otherPlayers).length);

    }

},1);

function appLoop() {
    let x = playerSocketInfo.x - player.x;
    let y = playerSocketInfo.y - player.y;
    if (x > 0) {
        player.x = Math.floor(player.x + x / 2);
    }
    if (x < 0) {
        player.x = Math.floor(player.x + x / 2);
    }
    if (y > 0) {
        player.y = Math.floor(player.y + y / 2);
    }
    if (y < 0) {
        player.y = Math.floor(player.y + y / 2);
    }
    console.log(Object.entries(otherPlayerSocketInfo).length);
    for (let [key, value] of Object.entries(otherPlayerSocketInfo)) {
        if(!otherPlayers[key]){
            otherPlayers[key] = PIXI.Sprite.from(loader.resources['bunny'].texture);
            otherPlayers[key].anchor.set(0.5);
            otherPlayers[key].x = 0;
            otherPlayers[key].y = 0;
            app.stage.addChild(otherPlayers[key]);
        }else{
            let x = value.x - otherPlayers[key].x;
            let y = value.y - otherPlayers[key].y;
            if (x > 0) {
                otherPlayers[key].x = Math.floor(otherPlayers[key].x + x / 2);
            }
            if (x < 0) {
                otherPlayers[key].x = Math.floor(otherPlayers[key].x + x / 2);
            }
            if (y > 0) {
                otherPlayers[key].y = Math.floor(otherPlayers[key].y + y / 2);
            }
            if (y < 0) {
                otherPlayers[key].y = Math.floor(otherPlayers[key].y + y / 2);
            }
        }
        otherPlayers[key].getLastTime = Date.now();
    }

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
}

document.addEventListener('keydown', (event) => {
    keysPressed[event.code] = true;
});

document.addEventListener('keyup', (event) => {
    delete keysPressed[event.code];
});