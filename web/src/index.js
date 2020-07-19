import {Application, Loader, Text, TextStyle} from "pixi.js";
import {Lobby} from "./lobby";
import {Bullet, Player} from "./sprites";
import {ContainerPlayer, PlayerHpContainer} from "./containers";

export let socket;
export let loader = Loader.shared;

loader.add("bunny", "static/img/bunny.png");
loader.add("bullet", "static/img/bullet.png");
loader.add("bulletEnemies", "static/img/enemiesBullet.png");
loader.load();

let app;
let player;
let playerSocketInfo = {
    x: 0,
    y: 0,
};
let otherPlayers = [];
let otherPlayerSocketInfo = [];
let keysPressed = {};
let mousePosition = {x: 0, y: 0};
let bullets = [];
let bulletsSocketInfo = [];
let othersBullets = [];
let othersBulletsSocketInfo = [];

document.getElementById("greet").hidden = false;
document.getElementById("choose-multi").addEventListener("click", startGame);

function startGame() {
    document.getElementById("greet").hidden = true;
    document.getElementById("app").hidden = false;

    socket = new WebSocket("ws://127.0.0.1:3000/ws");
    socket.onmessage = function (e) {
        const messageText = e.data;
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
            case "SIGNAL_CONF_THE_GAME":
                app = new Application({
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
            case "SIGNAL_LOBBY_LIST":
                clean();
                let i = 1;
                response.lobbies.forEach(function (element) {
                    let lobbyLink = new Lobby(
                        `${i.toString()}.${element.title}(${(
                            element.max - element.free
                        ).toString()}/${element.max.toString()})`
                    );
                    lobbyLink.setCoordinates(i);
                    lobbyLink.lobbyId = element.id;
                    app.stage.addChild(lobbyLink);
                    i++;
                });
                let lobbyLink = new Lobby("Create new lobby");
                lobbyLink.setCoordinates(i);
                app.stage.addChild(lobbyLink);

                break;
            case "SIGNAL_START_THE_GAME":
                clean();
                player = createPlayer(response.info.player);
                app.stage.addChild(player);
                app.ticker.add(appLoop);
                break;
            case "SIGNAL_INFO_THE_GAME":
                playerSocketInfo = response.info.player;
                for (let [key, value] of Object.entries(response.info.others)) {
                    otherPlayerSocketInfo[key] = value
                }
                for (let [key, value] of Object.entries(response.info.bullets)) {
                    bulletsSocketInfo[key] = value
                }
                for (let [key, value] of Object.entries(response.info.othersBullets)) {
                    othersBulletsSocketInfo[key] = value
                }
                break;
        }
    };

    socket.onclose = function (event) {
        if (event.wasClean) {
            console.log(
                `[close] Соединение закрыто чисто, код=${event.code} причина=${event.reason}`
            );
        } else {
            // например, сервер убил процесс или сеть недоступна
            // обычно в этом случае event.code 1006
            console.log("[close] Соединение прервано");
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
    moveObject(playerSocketInfo, player);
    player.ContainerHp.ChangeHp(playerSocketInfo.hp)

    for (const property in bulletsSocketInfo) {

    }
        otherPlayerSocketInfo.forEach(function (item,index) {
        if (!otherPlayers[index]) {
            otherPlayers[index] = createPlayer(item);
            app.stage.addChild(otherPlayers[index]);
        } else {
            moveObject(item, otherPlayers[index]);
        }
        otherPlayers[index].getLastTime = Date.now();
        otherPlayers[index].ContainerHp.ChangeHp(item.hp)
    });

    bulletsSocketInfo.forEach(function (item,index,object) {
        console.log(item);

        if (!bullets[index]) {
            bullets[index] = createBullet(item, false);
            app.stage.addChild(bullets[index]);
        } else {
            moveObject(item, bullets[index]);
            if (item.deleted) {
                object.splice(index,1);
            }
        }
    });

    othersBulletsSocketInfo.forEach(function (item,index,object) {
        if (!othersBullets[index]) {
            othersBullets[index] = createBullet(item, false);
            app.stage.addChild(othersBullets[index]);
        } else {
            moveObject(item, othersBullets[index]);
            if (item.deleted) {
                object.splice(index,1);
            }
        }
    });

    // for (let i = 0, c = bullets.length; i < c; i++) {
    //     console.log(bullets[i]);
    //     if (bullets[i] && bullets[i].deleted) {
    //         app.stage.removeChild(bullets[i]);
    //         bullets.splice(i, 1);
    //     }
    // }
    //
    // for (let i = 0, c = othersBullets.length; i < c; i++) {
    //     console.log(othersBullets[i]);
    //     if (othersBullets[i] && othersBullets[i].deleted) {
    //         app.stage.removeChild(othersBullets[i]);
    //         bullets.splice(i, 1);
    //     }
    // }

    let dir = "";
    if (keysPressed["KeyD"] && keysPressed["KeyD"] === true) {
        dir = "right";
    }
    if (keysPressed["KeyS"] && keysPressed["KeyS"] === true) {
        if (dir !== "") {
            dir += "-";
        }
        dir += "down";
    }
    if (keysPressed["KeyW"] && keysPressed["KeyW"] === true) {
        if (dir !== "") {
            dir += "-";
        }
        dir += "up";
    }
    if (keysPressed["KeyA"] && keysPressed["KeyA"] === true) {
        if (dir !== "") {
            dir += "-";
        }
        dir += "left";
    }

    if (keysPressed["Space"]) {
        socket.send(`{"type":"command","payload":{"name":"shoot","bullet":{"x":${mousePosition.x.toString()},"y":${mousePosition.y.toString()}}}}`);
    }

    if (dir === "") {
        return;
    }

    socket.send(`{"type":"command","payload":{"name":"${dir}"}}`);
}

function createBullet(bulletSocket, isEnemy) {
    let bullet = new Bullet(isEnemy);
    bullet.x = bulletSocket.x;
    bullet.y = bulletSocket.y;
    bullet.deleted = bulletSocket.deleted;
    return bullet;
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

function createPlayer(objectSocket) {
    let container = new ContainerPlayer();
    let player = new Player();
    let title = new Text(objectSocket.name);
    title.style = new TextStyle({
        fill: 0x30728,
        fontSize: 10,
    });
    title.y = -35;
    title.x = -15;

    container.SetPlayer(player);
    container.SetTitle(title);
    container.SetContainerHp(new PlayerHpContainer(objectSocket.hp, objectSocket.maxHp));
    return container;
}

document.addEventListener("keydown", (event) => {
    keysPressed[event.code] = true;
});

document.addEventListener("keyup", (event) => {
    delete keysPressed[event.code];
});
