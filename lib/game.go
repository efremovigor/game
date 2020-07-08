package lib

import "sync"

const GameWidth = 800
const GameHeight = 600
const RequestTypeNewCommand = "command"
const RequestTypeNewPlayer = "newPlayer"
const RequestTypeLobbyList = "lobbyList"
const RequestInit = "init"

const commandFire = "fire"
const commandUp = "up"
const commandDown = "down"
const commandLeft = "left"
const commandRight = "right"

const SignalConfTheGame = "SIGNAL_CONF_THE_GAME"
const SignalLobbyList = "SIGNAL_LOBBY_LIST"
const SignalStartTheGame = "SIGNAL_START_THE_GAME"
const SignalInfoTheGame = "SIGNAL_INFO_THE_GAME"

const MaxUserInLobby = 4

var Connections = make(map[string]*PlayerConnection)
var Games = make(map[string]*Game)
var UniGames = make(map[string]*Game)

type PlayerConnection struct {
	Connection *ConnectionReceiver
	SessionId  string
	Name       string
	Command    chan string
	InGame     bool
	Player     *Player
}

type Game struct {
	Connection map[string]*PlayerConnection
	Width      int
	Height     int
	Lock       sync.Mutex
}

func (playerConnection *PlayerConnection) Move(game Game, command string) {
	switch command {
	case "up":
		playerConnection.Player.Y -= 10
		if playerConnection.Player.Y < 0 {
			playerConnection.Player.Y = 0
		}
	case "down":
		playerConnection.Player.Y += 10
		if playerConnection.Player.Y > game.Height-playerConnection.Player.H {
			playerConnection.Player.Y = game.Height - playerConnection.Player.H
		}
	case "left":
		playerConnection.Player.X -= 10
		if playerConnection.Player.X < 0 {
			playerConnection.Player.X = 0
		}
	case "right":
		playerConnection.Player.X += 10
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
