let PIXI = require('pixi.js');
let loader = PIXI.Loader.shared;
loader.add('bunny', 'static/bunny.png');
loader.add('bullet', 'static/bullet.png');
loader.add('bulletEnemies', 'static/enemiesBullet.png');
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
let mousePosition = {x: 0, y: 0};
let bullets = [];
let bulletsSocketInfo = [];
let othersBullets = [];
let othersBulletsSocketInfo = [];

document.getElementById("greet").hidden = false;
document.getElementById("choose-multi").addEventListener('click', startGame);

function startGame() {
    document.getElementById("greet").hidden = true;
    document.getElementById("app").hidden = false;

    socket = new WebSocket('ws://127.0.0.1:3000/ws');
    socket.onmessage = function (event) {
        const messageText = event.data;
        // const message = JSON.parse(messageText);
        // console.log(message);
    };

    socket.onopen = function (e) {
        socket.send('{"type":"init"}');
        socket.send('{"type":"lobbyList"}');
    };

    socket.onmessage = function (event) {
        let response = JSON.parse(event.data);
        switch (response.type) {
            case 'SIGNAL_CONF_THE_GAME':
                app = new PIXI.Application({
                    width: response.conf.width,
                    height: response.conf.height,
                    backgroundColor: 0x53b4ff,
                    resolution: window.devicePixelRatio || 1,
                });
                document.getElementById("app").appendChild(app.view);
                app.stage.interactive = true;
                app.stage.on("pointermove", function (e) {
                    mousePosition.x = e.data.global.x;
                    mousePosition.y = e.data.global.y;
                });
                break;
            case 'SIGNAL_LOBBY_LIST':
                clean();
                let i = 1;
                response.lobbies.forEach(function (element) {
                    let lobbyLink = new PIXI.Text(i.toString() + '.' + element.title + '(' + (element.max - element.free).toString() + '/' + element.max.toString() + ')');
                    lobbyLink.x = 10;
                    lobbyLink.y = (i - 1) * 15 + 3;
                    lobbyLink.style = new PIXI.TextStyle({
                        fill: 0x223AFF,
                        fontSize: 15
                    });
                    lobbyLink.alpha = 0.3;
                    lobbyLink.lobbyId = element.id;
                    lobbyLink.mouseover = function () {
                        this.alpha = 1;
                    };
                    lobbyLink.mouseout = function () {
                        this.alpha = 0.3;
                    };
                    lobbyLink.interactive = true;
                    lobbyLink.click = function () {
                        socket.send('{"type":"newPlayer","lobby":"' + lobbyLink.lobbyId.toString() + '","payload":{"name":"' + document.getElementById("username").value + '"}}');
                    };
                    app.stage.addChild(lobbyLink);
                    i++
                });
                let lobbyLink = new PIXI.Text('Create new lobby');
                lobbyLink.x = 10;
                lobbyLink.y = (i - 1) * 15 + 3;
                lobbyLink.style = new PIXI.TextStyle({
                    fill: 0x223AFF,
                    fontSize: 15
                });
                lobbyLink.alpha = 0.3;

                lobbyLink.mouseover = function () {
                    this.alpha = 1;
                };
                lobbyLink.mouseout = function () {
                    this.alpha = 0.2;
                };
                lobbyLink.interactive = true;
                lobbyLink.click = function () {
                    socket.send('{"type":"newPlayer","payload":{"name":"' + document.getElementById("username").value + '"}}');
                };
                app.stage.addChild(lobbyLink);

                break;
            case 'SIGNAL_START_THE_GAME':
                clean();
                player = PIXI.Sprite.from(loader.resources['bunny'].texture);
                player.anchor.set(0.5);
                player.x = 0;
                player.y = 0;
                app.stage.addChild(player);
                app.ticker.add(appLoop);
                break;
            case 'SIGNAL_INFO_THE_GAME':
                playerSocketInfo = response.info.player;
                otherPlayerSocketInfo = response.info.others;
                bulletsSocketInfo = response.info.bullets;
                othersBulletsSocketInfo = response.info.othersBullets;
                 console.log(response.info.othersBullets);
                break;
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
        if (value.getLastTime + 1500 < Date.now()) {
            app.stage.removeChild(otherPlayers[key]);
            delete otherPlayers[key];
        }
    }

}, 1);

function clean() {
    while (app.stage.children.length > 0) {
        let child = app.stage.getChildAt(0);
        app.stage.removeChild(child);
    }
}

function appLoop() {
    moveObject(playerSocketInfo,player)
    for (let [key, value] of Object.entries(otherPlayerSocketInfo)) {
        if (!otherPlayers[key]) {
            otherPlayers[key] = createPlayer()
            app.stage.addChild(otherPlayers[key]);
        } else {
            moveObject(value, otherPlayers[key]);
        }
        otherPlayers[key].getLastTime = Date.now();
    }

    for (let [key, value] of Object.entries(bulletsSocketInfo)) {
        if (!bullets[key]) {
            bullets[key] = createBullet(value,false);
            app.stage.addChild(bullets[key]);
        } else {
            moveObject(value, bullets[key]);
        }
    }

    for (let [key, value] of Object.entries(othersBulletsSocketInfo)) {
        if (!othersBullets[key]) {
            othersBullets[key] = createBullet(value,true);
            app.stage.addChild(othersBullets[key]);
        } else {
            moveObject(value, othersBullets[key]);
        }
    }

    for (let i = 0, c = bullets.length; i < c; i++) {
        if (bullets[i] && bullets[i].dead) {
            app.stage.removeChild(bullets[i]);
            bullets.splice(i, 1);
        }
    }

    for (let i = 0, c = othersBullets.length; i < c; i++) {
        if (othersBullets[i] && othersBullets[i].dead) {
            app.stage.removeChild(othersBullets[i]);
            bullets.splice(i, 1);
        }
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

    if (keysPressed['Space']) {
        socket.send('{"type":"command","payload":{"name":"shoot","bullet":{"x":' + mousePosition.x.toString() + ',"y":' + mousePosition.y.toString() + '}}}');
    }

    if (dir === '') {
        return;
    }

    socket.send('{"type":"command","payload":{"name":"' + dir + '"}}');
}

function createBullet(bulletSocket,isEnemy) {
    let bullet = new PIXI.Sprite.from(isEnemy ? loader.resources['bulletEnemies'].texture : loader.resources['bullet'].texture);
    bullet.anchor.set(0.5);
    bullet.dead = false;
    bullet.x = bulletSocket.x;
    bullet.y = bulletSocket.y;
    return bullet;
}

function createPlayer() {
    let player = PIXI.Sprite.from(loader.resources['bunny'].texture);
    player.anchor.set(0.5);
    player.x = 0;
    player.y = 0;
    return player
}

function moveObject(objectSocket, object) {
    let x = objectSocket.x - object.x;
    let y = objectSocket.y - object.y;
    if (x > 0) {
        object.x = Math.floor(object.x + x / 2);
    }
    if (x < 0) {
        object.x = Math.floor(object.x + x / 2);
    }
    if (y > 0) {
        object.y = Math.floor(object.y + y / 2);
    }
    if (y < 0) {
        object.y = Math.floor(object.y + y / 2);
    }
}

document.addEventListener('keydown', (event) => {
    keysPressed[event.code] = true;
});

document.addEventListener('keyup', (event) => {
    delete keysPressed[event.code];
});