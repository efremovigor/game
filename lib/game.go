package lib

import (
	"crypto/md5"
	"fmt"
	"math"
	"sync"
	"time"
)

const GameWidth = 1000
const GameHeight = 900
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
const MaxHp = 100
const BulletSpeed = 5
const MaxDistanceBulletOutScreen = BulletSpeed * 10
const PlayerSpeed = 5
const PlayerWidth = 26
const PlayerHeight = 37
const PlayerStartPositionX = PlayerWidth/2 + 1
const PlayerStartPositionY = PlayerHeight/2 + 1

var Connections = make(map[string]*PlayerConnection)
var Games = make(map[string]*Game)
var UniGames = make(map[string]*Game)

type PlayerConnection struct {
	Connection *ConnectionReceiver
	SessionId  string
	Request    chan LoginJsonRequest
	InGame     bool
	Player     *Player
}

type Game struct {
	Connection map[string]*PlayerConnection
	Bullets    map[string]map[[16]byte]*BulletGame
	Enemies    map[[16]byte]*Enemy
	Builds     []Build
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

type Enemy struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	X                   int    `json:"x"`
	Y                   int    `json:"y"`
	W                   int    `json:"w"`
	H                   int    `json:"h"`
	Hp                  int    `json:"hp"`
	MaxHp               int    `json:"maxHp"`
	Destination         Node
	LatestConsideration int64
	Path                []Node
}

func (enemy Enemy) getPositionKey() string {
	return fmt.Sprintf("%d%d", enemy.X, enemy.Y)
}

type Node struct {
	X        int
	Y        int
	Distance float64
	next     map[string]*Node
	Back     *Node
}

func (node Node) getPositionKey() string {
	return fmt.Sprintf("%d%d", node.X, node.Y)
}

type StringPositionInterface interface {
	getPositionKey() string
}

type Searching struct {
	ComeFrom       Enemy
	Destination    Player
	Path           *Node
	VisitedPoints  map[string]Node
	CheckingPoints map[string]Node
	MinDistance    float64
	CurrentNode    *Node
}

func (searching *Searching) setNearestPoint() bool {
	var nearestNode Node
	searching.CurrentNode = nil
	for _, node := range searching.CheckingPoints {
		if node.Distance <= searching.MinDistance {
			nearestNode = node
			break
		}
		if nearestNode.Distance == 0 || nearestNode.Distance > node.Distance {
			nearestNode = node
		}
	}
	searching.CurrentNode = &nearestNode

	fmt.Println("----------------------")
	fmt.Println("Ближайшая точка")
	fmt.Println(searching.CurrentNode.getPositionKey(), ":", searching.CurrentNode.Distance)
	fmt.Println("++++++++++++++++++++++")

	searching.VisitedPoints[searching.CurrentNode.getPositionKey()] = searching.CheckingPoints[searching.CurrentNode.getPositionKey()]
	delete(searching.CheckingPoints, searching.CurrentNode.getPositionKey())
	if searching.CurrentNode == nil {
		return false
	}
	return true
}

func (searching *Searching) Handle() Node {
	searching.VisitedPoints = make(map[string]Node)
	searching.CheckingPoints = make(map[string]Node)

	node := Node{X: searching.ComeFrom.X, Y: searching.ComeFrom.Y, next: make(map[string]*Node)}
	searching.VisitedPoints[searching.ComeFrom.getPositionKey()] = node
	siblings := searching.getSiblings(node)
	if len(siblings) > 0 {
		for _, sibling := range siblings {
			node.next[sibling.getPositionKey()] = &sibling
			searching.CheckingPoints[sibling.getPositionKey()] = sibling
		}
	}

	for searching.setNearestPoint() {
		siblings := searching.getSiblings(*searching.CurrentNode)
		if len(siblings) > 0 {
			for _, sibling := range siblings {
				if sibling.getPositionKey() == searching.Destination.getPositionKey() {
					return sibling
				}
				fmt.Println(sibling.getPositionKey(), ":", sibling.Distance)

				searching.CurrentNode.next[sibling.getPositionKey()] = &sibling
				searching.CheckingPoints[sibling.getPositionKey()] = sibling
			}
		}
		searching.VisitedPoints[searching.CurrentNode.getPositionKey()] = *searching.CurrentNode
	}
	return Node{}
}

func (searching *Searching) getSiblings(node Node) (siblings []Node) {
	newSiblings := []Node{node.getLeftSibling(), node.getRightSibling(), node.getUpSibling(), node.getDownSibling()}
	for _, sibling := range newSiblings {
		if _, ok := searching.VisitedPoints[sibling.getPositionKey()]; !ok && sibling.X > 0 && sibling.Y > 0 && sibling.X < GameWidth && sibling.Y < GameHeight {
			sibling.Back = &node
			sibling.next = make(map[string]*Node)
			sibling.Distance = math.Sqrt(math.Pow(float64(searching.Destination.X-sibling.X), 2) + math.Pow(float64(searching.Destination.Y-sibling.Y), 2))
			if searching.MinDistance > sibling.Distance {
				searching.MinDistance = sibling.Distance
			}
			siblings = append(siblings, sibling)
		}
	}
	return
}

func (node Node) getLeftSibling() Node {
	return Node{X: node.X - 1, Y: node.Y}
}

func (node Node) getRightSibling() Node {
	return Node{X: node.X + 1, Y: node.Y}
}

func (node Node) getUpSibling() Node {
	return Node{X: node.X, Y: node.Y - 1}
}

func (node Node) getDownSibling() Node {
	return Node{X: node.X, Y: node.Y + 1}
}

func (node Node) GetPath() []Node {
	var nodes []Node
	nodes = append(nodes, Node{X: node.X, Y: node.Y})
	for node.Back != nil {
		node = *node.Back
		nodes = append(nodes, Node{X: node.X, Y: node.Y})
	}

	return nodes
}

func (bullet *BulletGame) MoveBullet(game *Game, sessionId string) {
	for i := 0; i < 10; i++ {
		bullet.Bullet.X += bullet.XStep / 10
		bullet.Bullet.Y += bullet.YStep / 10
		if bullet.Bullet.X > GameWidth+MaxDistanceBulletOutScreen || bullet.Bullet.X < -MaxDistanceBulletOutScreen || bullet.Bullet.Y > GameHeight+MaxDistanceBulletOutScreen || bullet.Bullet.Y < -MaxDistanceBulletOutScreen {
			bullet.Deleted = true
		} else {
			for _, player := range game.Connection {
				distance := math.Sqrt(math.Pow(bullet.Bullet.Y-float64(player.Player.Y), 2) + math.Pow(bullet.Bullet.X-float64(player.Player.X), 2))
				if distance < 15 && bullet.Deleted == false && sessionId != player.SessionId {
					bullet.Deleted = true
					player.Player.Hp -= 10
					if player.Player.Hp < 0 {
						player.Player.Hp = 0
					}
					return
				}
			}
			for _, build := range game.Builds {
				if float64(build.X) < bullet.Bullet.X && float64(build.Y) < bullet.Bullet.Y && float64(build.X+build.Width) > bullet.Bullet.X && float64(build.Y+build.Height) > bullet.Bullet.Y {
					bullet.Deleted = true
					return
				}
			}
		}
	}
}

func (playerConnection *PlayerConnection) Shoot(game *Game, requestBullet Bullet) {
	game.Lock.Lock()
	defer game.Lock.Unlock()
	if playerConnection.Player.LatestShoot+int64(time.Second/2) >= time.Now().UnixNano() {
		return
	}

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
	bullet.Bullet.X = float64(bullet.Player.X)
	bullet.Bullet.Y = float64(bullet.Player.Y)
	bulletKey := md5.Sum([]byte(fmt.Sprintf("%d%d%f%f%d", playerConnection.Player.X, playerConnection.Player.Y, requestBullet.X, requestBullet.Y, time.Now().UnixNano())))
	bullets[bulletKey] = &bullet
	playerConnection.Player.LatestShoot = time.Now().UnixNano()
}

func (playerConnection *PlayerConnection) Move(game *Game, command string) {
	game.Lock.Lock()
	defer game.Lock.Unlock()
	x := playerConnection.Player.X
	y := playerConnection.Player.Y
	switch command {
	case CommandUp:
		y = playerConnection.Player.Y - PlayerSpeed
		if y < PlayerStartPositionY {
			y = playerConnection.Player.Y
		}
	case CommandDown:
		y = playerConnection.Player.Y + PlayerSpeed
		if y > game.Height-PlayerStartPositionY {
			y = game.Height - PlayerStartPositionY
		}
	case CommandLeft:
		x = playerConnection.Player.X - PlayerSpeed
		if x < PlayerStartPositionX {
			x = playerConnection.Player.X
		}
	case CommandRight:
		x = playerConnection.Player.X + PlayerSpeed
		if x > game.Width-PlayerStartPositionY {
			x = game.Width - PlayerStartPositionY
		}
	}

	for _, build := range game.Builds {
		xBegin := x - (PlayerWidth / 2)
		xEnd := x + (PlayerWidth / 2)
		yBegin := y - (PlayerHeight / 2)
		yEnd := y + (PlayerHeight / 2)
		if build.X+build.Width > xBegin && build.X < xBegin && build.Y < yBegin && build.Y+build.Height > yBegin {
			return
		}
		if build.X+build.Width > xEnd && build.X < xEnd && build.Y < yEnd && build.Y+build.Height > yEnd {
			return
		}
		if build.X+build.Width > xBegin && build.X < xBegin && build.Y < yEnd && build.Y+build.Height > yEnd {
			return
		}
		if build.X+build.Width > xEnd && build.X < xEnd && build.Y < yBegin && build.Y+build.Height > yBegin {
			return
		}
	}

	playerConnection.Player.X = x
	playerConnection.Player.Y = y
}

type Player struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	W           int    `json:"w"`
	H           int    `json:"h"`
	Hp          int    `json:"hp"`
	MaxHp       int    `json:"maxHp"`
	LatestShoot int64
}

func (player Player) getPositionKey() string {
	return fmt.Sprintf("%d%d", player.X, player.Y)
}

type Build struct {
	Type   int `json:"type"`
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
