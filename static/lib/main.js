let PIXI = require('pixi.js');
let loader = PIXI.Loader.shared;
loader.add('bunny', 'static/bunny.png');
loader.add('bullet', 'static/bullet.png');
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
let lobbies;
let keysPressed = {};
let mousePosition = {x: 0, y: 0};
let bullets = [];

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
            console.log('remove' + key);
            app.stage.removeChild(otherPlayers[key]);
            delete otherPlayers[key];
        }
        console.log('count:' + Object.entries(otherPlayers).length);

    }

}, 1);

function clean() {
    while (app.stage.children.length > 0) {
        let child = app.stage.getChildAt(0);
        app.stage.removeChild(child);
    }
}

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
        if (!otherPlayers[key]) {
            otherPlayers[key] = PIXI.Sprite.from(loader.resources['bunny'].texture);
            otherPlayers[key].anchor.set(0.5);
            otherPlayers[key].x = 0;
            otherPlayers[key].y = 0;
            app.stage.addChild(otherPlayers[key]);
        } else {
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

    if (keysPressed['Space']) {
        let bullet = new PIXI.Sprite.from(loader.resources['bullet'].texture);
        bullet.anchor.set(0.5);
        bullet.dead = false;
        bullet.x = player.x;
        bullet.y = player.y;
        let alfa = Math.atan(mousePosition.y / mousePosition.x);
        bullet.Xstep = 2 * Math.cos(alfa);
        bullet.Ystep = 2 * Math.sin(alfa);
        app.stage.addChild(bullet);
        bullets.push(bullet);
    }

    for (let i = 0, c = bullets.length; i < c; i++) {
        bullets[i].x -= bullets[i].Xstep;
        if (bullets[i].x > app.width || bullets[i].x < 0) {
            bullets[i].dead = true;
        }
        bullets[i].y -= bullets[i].Ystep;
        if (bullets[i].y > app.height || bullets[i].y < 0) {
            bullets[i].dead = true;
        }}

    for (let i = 0, c = bullets.length; i < c; i++) {
        if (bullets[i] && bullets[i].dead) {
            app.stage.removeChild(bullets[i]);
            bullets.splice(i, 1);
        }
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