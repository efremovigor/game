package lib

const RequestTypeNewCommand = "command"
const RequestTypeNewPlayer = "newPlayer"

const commandFire = "fire"
const commandUp = "up"
const commandDown = "down"
const commandLeft = "left"
const commandRight = "right"

const SignalStartTheGame = "SIGNAL_START_THE_GAME"
const SignalInfoTheGame = "SIGNAL_INFO_THE_GAME"

var Connections = make(map[string]*PlayerConnection)
var Games = make(map[string]*Game)

type PlayerConnection struct {
	Connection *ConnectionReceiver
	SessionId  string
	Name       string
	Command    chan string
	InGame     bool
}

type Game struct {
	Connection PlayerConnection
	Player     *Player
	Weight     int
	Height     int
}

func (game *Game) Move(command string) {
	switch command {
	case "up":
		game.Player.Y -= 5
		if game.Player.Y < 0 {
			game.Player.Y = 0
		}
	case "down":
		game.Player.Y += 5
		if game.Player.Y > game.Height-game.Player.H {
			game.Player.Y = game.Height - game.Player.H
		}
	case "left":
		game.Player.X -= 5
		if game.Player.X < 0 {
			game.Player.X = 0
		}
	case "right":
		game.Player.X += 5
		if game.Player.X > game.Weight-game.Player.W {
			game.Player.X = game.Weight - game.Player.W
		}
	}
}

type Player struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}
