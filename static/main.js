document.getElementById("greet").hidden = false;

document.getElementById("choose-multi").addEventListener('click', startGame);

let game = {};
game.canvas = document.getElementById("game-canvas");
game.canvas.width = 1000;
game.canvas.height = 1000;
game.ctx = game.canvas.getContext("2d");
game.player = {
    x: 15,
    y: 15,
    w: 10,
    h: 10,
};



function gameLoop() {

    game.ctx.clearRect(0, 0, game.canvas.width, game.canvas.height);
    game.ctx.fillStyle = "green";
    game.ctx.fillRect(game.player.x, game.player.y, game.player.w, game.player.h);
}

document.addEventListener('keydown', function (event) {
    if (event.code === 'KeyD') {
        game.player.x += game.player.w
    }
    if (event.code === 'KeyS') {
        game.player.y += game.player.h
    }
    if (event.code === 'KeyW') {
        game.player.y -= game.player.h
    }
    if (event.code === 'KeyA') {
        game.player.x -= game.player.w
    }
});
// var conn = new WebSocket('ws://localhost:8080/echo');
// conn.onmessage = function(e){ console.log(e.data); };
// conn.onopen = () => conn.send('hello');

let socket = new WebSocket('ws://127.0.0.1:3000/ws');
socket.onmessage = function (event) {
    const messageText = event.data;
    const message = JSON.parse(messageText);
    console.log(message);
};

const interval = setInterval(socket.onopen = () => socket.send('hello'), 10 * 1000);

socket.onclose = function () {
    clearInterval(interval);
};

function startGame(socket) {
    document.getElementById("greet").hidden = true;
    document.getElementById("game").hidden = false;
    document.getElementById("game-canvas").hidden = false;

    game.interval = setInterval(() => gameLoop(), 100);
    socket.send('newPlayer');

}