import { Container, Graphics } from "pixi.js";

export class ContainerPlayer extends Container {

  SetPlayer(player) {
    this.player = player;
    this.addChild(player);
  }

  SetTitle(title) {
    this.title = title;
    this.addChild(title);
  }

  SetContainerHp(containerHp) {
    this.ContainerHp = containerHp;
    this.addChild(containerHp);
  }
}

export class PlayerHpContainer extends Container {

  constructor(playerHp, playerMaxHp) {
    super();
    this.playerHp = playerHp;
    this.playerMaxHp = playerMaxHp;
    this.RenderHp();
  }

  ChangeHp(hp) {
    if (hp < 0) {
      return;
    }
    this.playerHp = hp;
    this.RenderHp();
  }

  RenderHp() {
    let maxHp = new Graphics();

    maxHp.beginFill(0x4e4747);
    maxHp.drawRect(-20, -20, 5, 40);
    maxHp.endFill();
    let hp = new Graphics();

    hp.beginFill(0xde3249);
    hp.drawRect(-20, -20, 5, (40 / this.playerMaxHp) * this.playerHp);
    hp.endFill();

    this.addChild(maxHp);
    this.addChild(hp);
    this.rotation = 3.15;
  }
}
