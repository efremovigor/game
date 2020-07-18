import { Sprite } from "pixi.js";
import { loader } from "./index.js";

export class Player extends Sprite {
  constructor() {
    super();
    this.texture = loader.resources["bunny"].texture;
    this.anchor.set(0.5);
    this.x = 0;
    this.y = 0;
  }
}

export class Bullet extends Sprite {
  constructor(isEnemy) {
    super();
    this.texture = isEnemy
      ? loader.resources["bulletEnemies"].texture
      : loader.resources["bullet"].texture;
    this.anchor.set(0.5);
    this.dead = false;
  }
}
