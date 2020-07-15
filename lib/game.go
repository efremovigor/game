package lib

import (
	"crypto/md5"
	"fmt"
	"math"
	"sync"
	"time"
)

const GameWidth = 800
const GameHeight = 600
const RequestTypeNewCommand = "command"
const RequestTypeNewPlayer = "newPlayer"
const RequestTypeLobbyList = "lobbyList"
const RequestInit = "init"

const CommandShoot = "shoot"
const CommandUp = "up"
const CommandDown = "down"
const CommandLeft = "left"
const CommandRight = "right"

const SignalConfTheGame = "SIGNAL_CONF_THE_GAME"
const SignalLobbyList = "SIGNAL_LOBBY_LIST"
const SignalStartTheGame = "SIGNAL_START_THE_GAME"
const SignalInfoTheGame = "SIGNAL_INFO_THE_GAME"

const MaxUserInLobby = 4
const BulletSpeed = 10
const PlayerSpeed = 5

var Connections = make(map[string]*PlayerConnection)
var Games = make(map[string]*Game)
var UniGames = make(map[string]*Game)

type PlayerConnection struct {
	Connection *ConnectionReceiver
	SessionId  string
	Name       string
	Request    chan LoginJsonRequest
	InGame     bool
	Player     *Player
}

type Game struct {
	Connection map[string]*PlayerConnection
	Bullets    map[string]map[[16]byte]*BulletGame
	Width      int
	Height     int
	Lock       sync.Mutex
}

type BulletGame struct {
	Bullet  Bullet
	Player  Player
	delta   float64
	XStep   float64
	YStep   float64
	Deleted bool
}

func (playerConnection *PlayerConnection) Shoot(game Game, requestBullet Bullet) {
	bullets, ok := game.Bullets[playerConnection.SessionId]
	if !ok {
		bullets = make(map[[16]byte]*BulletGame)
		game.Bullets[playerConnection.SessionId] = bullets
	}
	bullet := BulletGame{Bullet: requestBullet, Player: *playerConnection.Player}
	bullet.delta = math.Atan((requestBullet.Y - float64(playerConnection.Player.Y)) / (requestBullet.X - float64(playerConnection.Player.X)))
	var bulletLeft int
	if requestBullet.X-float64(playerConnection.Player.X) >= 0 {
		bulletLeft = 1
	} else {
		bulletLeft = -1
	}
	bullet.XStep = float64(bulletLeft*BulletSpeed) * math.Cos(bullet.delta)
	bullet.YStep = float64(bulletLeft*BulletSpeed) * math.Sin(bullet.delta)
	bullets[md5.Sum([]byte(string(playerConnection.Player.X)+string(playerConnection.Player.Y)+fmt.Sprintf("%f", requestBullet.X)+fmt.Sprintf("%f", requestBullet.Y)+string(time.Now().Unix())))] = &bullet
}

func (playerConnection *PlayerConnection) Move(game Game, command string) {
	switch command {
	case CommandUp:
		playerConnection.Player.Y -= PlayerSpeed
		if playerConnection.Player.Y < 0 {
			playerConnection.Player.Y = 0
		}
	case CommandDown:
		playerConnection.Player.Y += PlayerSpeed
		if playerConnection.Player.Y > game.Height-playerConnection.Player.H {
			playerConnection.Player.Y = game.Height - playerConnection.Player.H
		}
	case CommandLeft:
		playerConnection.Player.X -= PlayerSpeed
		if playerConnection.Player.X < 0 {
			playerConnection.Player.X = 0
		}
	case CommandRight:
		playerConnection.Player.X += PlayerSpeed
		if playerConnection.Player.X > game.Width-playerConnection.Player.W {
			playerConnection.Player.X = game.Width - playerConnection.Player.W
		}
	}
}

type Player struct {
	ID string `json:"id"`
	X  int    `json:"x"`
	Y  int    `json:"y"`
	W  int    `json:"w"`
	H  int    `json:"h"`
}
