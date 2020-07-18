import { Text, TextStyle } from "pixi.js";
import { socket } from "./index.js";

export class Lobby extends Text {
  //lobbyId

  constructor(text, style, canvas) {
    super(text, style, canvas);
    this.interactive = true;
    this.alpha = 0.2;
    this.style = new TextStyle({
      fill: 0x223aff,
      fontSize: 15,
    });

    this.mouseover = function () {
      this.alpha = 1;
    };
    this.mouseout = function () {
      this.alpha = 0.2;
    };
    this.click = function () {
      if (this.lobbyId !== undefined) {
        socket.send(
          `{"type":"newPlayer","lobby":"${this.lobbyId.toString()}",` +
            `"payload":{"name":"${document.getElementById("username").value}"}}`
        );
      } else {
        socket.send(
          `{"type":"newPlayer","payload":{"name":"${
            document.getElementById("username").value
          }"}}`
        );
      }
    };
    this.setCoordinates = function (i) {
      this.x = 10;
      this.y = (i - 1) * 15 + 3;
    };
  }
}
