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
	Connection            map[string]*PlayerConnection
	Bullets               map[string]map[[16]byte]*BulletGame
	Enemies               map[[16]byte]*Enemy
	Builds                []Build
	CrucialPoints         map[string]CrucialPoint
	CrucialPointsDistance map[string]float64
	Width                 int
	Height                int
	Lock                  sync.Mutex
}

func (game *Game) GetCrucialPoint(x int, y int) CrucialPoint {
	return game.CrucialPoints[CrucialPoint{X: x, Y: y}.GetKey()]
}

func (game *Game) AddCrucialPoint(x int, y int) {
	point := CrucialPoint{X: x, Y: y, Sibling: make(map[string]*CrucialPoint)}
	game.CrucialPoints[point.GetKey()] = point
}

func (game *Game) AddSiblingToCrucialPoint(point CrucialPoint, siblings ...CrucialPoint) {
	for _, sibling := range siblings {
		game.CrucialPoints[point.GetKey()].Sibling[sibling.GetKey()] = &sibling
		distance := GetDistance(point, sibling)
		game.CrucialPointsDistance[fmt.Sprintf("%s|%s", point.GetKey(), sibling.GetKey())] = distance
		game.CrucialPointsDistance[fmt.Sprintf("%s|%s", sibling.GetKey(), point.GetKey())] = distance
	}
}

func (game *Game) GetDistance(point1 CoordinateKeyInterface, point2 CoordinateKeyInterface) float64 {
	return game.CrucialPointsDistance[fmt.Sprintf("%s|%s", point1.GetKey(), point2.GetKey())]
}

func (game *Game) AddBuild(x int, y int, w int, h int) {
	game.Builds = append(game.Builds, Build{X: x, Y: y, Width: w, Height: h, Type: 1})
}

type BulletGame struct {
	Bullet  Bullet
	Player  Player
	delta   float64
	XStep   float64
	YStep   float64
	Deleted bool
}
type EnemyResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	W     int    `json:"w"`
	H     int    `json:"h"`
	Hp    int    `json:"hp"`
	MaxHp int    `json:"maxHp"`
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
	Path                chan Node
}

func (enemy Enemy) getX() int {
	return enemy.X
}

func (enemy Enemy) getY() int {
	return enemy.Y
}

func (enemy Enemy) getW() int {
	return enemy.W
}

func (enemy Enemy) getH() int {
	return enemy.H
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

type CrucialPoint struct {
	X       int
	Y       int
	Sibling map[string]*CrucialPoint
}

type NearestCrucialPoint struct {
	CrucialPoint
	Distance float64
}

type PathCrucialPoint struct {
	X        int
	Y        int
	Distance float64
	Sibling  map[string]*PathCrucialPoint
}

func (point CrucialPoint) getX() int {
	return point.X
}

func (point CrucialPoint) getY() int {
	return point.Y
}

func (point CrucialPoint) GetKey() string {
	return fmt.Sprintf("%d-%d", point.X, point.Y)
}

func (point PathCrucialPoint) GetKey() string {
	return fmt.Sprintf("%d-%d", point.X, point.Y)
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
	Builds         []Build
	Path           *Node
	VisitedPoints  map[string]Node
	CheckingPoints map[string]Node
	CurrentNode    *Node
}

func (searching *Searching) setNearestPoint() bool {
	var nearestNode Node
	searching.CurrentNode = nil

	if len(searching.CheckingPoints) == 0 {
		return false
	}

	for _, node := range searching.CheckingPoints {
		if nearestNode.Distance == 0 || nearestNode.Distance > node.Distance {
			nearestNode = node
		}
	}

	searching.CurrentNode = &nearestNode

	searching.VisitedPoints[searching.CurrentNode.getPositionKey()] = searching.CheckingPoints[searching.CurrentNode.getPositionKey()]
	delete(searching.CheckingPoints, searching.CurrentNode.getPositionKey())
	if searching.CurrentNode == nil {
		return false
	}

	for _, build := range searching.Builds {
		if build.CheckCollision(Enemy{X: searching.VisitedPoints[searching.CurrentNode.getPositionKey()].X, Y: searching.VisitedPoints[searching.CurrentNode.getPositionKey()].Y, W: searching.ComeFrom.W, H: searching.ComeFrom.H}) {
			fmt.Println("ошибка ", searching.VisitedPoints[searching.CurrentNode.getPositionKey()])
		}
	}
	searching.ComeFrom.Path <- searching.VisitedPoints[searching.CurrentNode.getPositionKey()]
	return true
}

func (searching *Searching) Handle(player *PlayerConnection) {
	searching.VisitedPoints = make(map[string]Node)
	searching.CheckingPoints = make(map[string]Node)

	node := Node{X: searching.ComeFrom.X, Y: searching.ComeFrom.Y, next: make(map[string]*Node)}
	if node.getPositionKey() == searching.Destination.getPositionKey() {
		return
	}
	searching.VisitedPoints[searching.ComeFrom.getPositionKey()] = node
	siblings := searching.getSiblings(node)
	if len(siblings) > 0 {
		for _, sibling := range siblings {
			if sibling.getPositionKey() == searching.Destination.getPositionKey() {
				return
			}
			node.next[sibling.getPositionKey()] = &sibling
			searching.CheckingPoints[sibling.getPositionKey()] = sibling
		}
	}

	for searching.setNearestPoint() {
		if player.Player.X != searching.Destination.X || player.Player.Y != searching.Destination.Y {
			return
		}
		siblings := searching.getSiblings(*searching.CurrentNode)
		if len(siblings) > 0 {
			for _, sibling := range siblings {
				if sibling.getPositionKey() == searching.Destination.getPositionKey() {
					return
				}
				searching.CurrentNode.next[sibling.getPositionKey()] = &sibling
				searching.CheckingPoints[sibling.getPositionKey()] = sibling
			}
		}
		searching.VisitedPoints[searching.CurrentNode.getPositionKey()] = *searching.CurrentNode
	}
}

func (searching *Searching) getSiblings(node Node) (siblings []Node) {
	newSiblings := []Node{node.getLeftSibling(), node.getRightSibling(), node.getUpSibling(), node.getDownSibling()}
	for _, sibling := range newSiblings {
		for _, build := range searching.Builds {
			if build.CheckCollision(Enemy{X: sibling.X, Y: sibling.Y, W: searching.ComeFrom.W, H: searching.ComeFrom.H}) {
				searching.VisitedPoints[sibling.getPositionKey()] = sibling
			}
		}
		if _, ok := searching.VisitedPoints[sibling.getPositionKey()]; !ok && sibling.X > 0 && sibling.Y > 0 && sibling.X < GameWidth && sibling.Y < GameHeight {
			sibling.Back = &node
			sibling.next = make(map[string]*Node)
			sibling.Distance = GetDistance(searching.Destination, sibling)
			siblings = append(siblings, sibling)
		}
	}
	return
}

func GetDistance(dest CoordinateInterface, target CoordinateInterface) float64 {
	return math.Sqrt(math.Pow(float64(dest.getX()-target.getX()), 2) + math.Pow(float64(dest.getY()-target.getY()), 2))
}

func (node Node) getX() int {
	return node.X
}

func (node Node) getY() int {
	return node.Y
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

func (bullet *BulletGame) MoveBullet(game *Game, sessionId string) {
	for i := 0; i < 10; i++ {
		bullet.Bullet.X += bullet.XStep / 10
		bullet.Bullet.Y += bullet.YStep / 10
		if bullet.Bullet.X > GameWidth+MaxDistanceBulletOutScreen || bullet.Bullet.X < -MaxDistanceBulletOutScreen || bullet.Bullet.Y > GameHeight+MaxDistanceBulletOutScreen || bullet.Bullet.Y < -MaxDistanceBulletOutScreen {
			bullet.Deleted = true
		} else {
			for _, player := range game.Connection {
				distance := GetDistance(bullet.Bullet, player.Player)
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
		if build.CheckCollision(Player{X: x, Y: y, W: playerConnection.Player.W, H: playerConnection.Player.H}) {
			return
		}
	}

	playerConnection.Player.X = x
	playerConnection.Player.Y = y
}

type CoordinateInterface interface {
	getX() int
	getY() int
}

type CoordinateKeyInterface interface {
	GetKey() string
}

type CollisionObjectInterface interface {
	CoordinateInterface
	getW() int
	getH() int
}

func (build Build) CheckCollision(object CollisionObjectInterface) bool {
	xBegin := object.getX() - (object.getW() / 2)
	xEnd := object.getX() + (object.getW() / 2)
	yBegin := object.getY() - (object.getH() / 2)
	yEnd := object.getY() + (object.getH() / 2)
	if build.X+build.Width > xBegin && build.X < xBegin && build.Y < yBegin && build.Y+build.Height > yBegin {
		return true
	}
	if build.X+build.Width > xEnd && build.X < xEnd && build.Y < yEnd && build.Y+build.Height > yEnd {
		return true
	}
	if build.X+build.Width > xBegin && build.X < xBegin && build.Y < yEnd && build.Y+build.Height > yEnd {
		return true
	}
	if build.X+build.Width > xEnd && build.X < xEnd && build.Y < yBegin && build.Y+build.Height > yBegin {
		return true
	}
	return false
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

func (bullet Bullet) getX() int {
	return int(bullet.X)
}

func (bullet Bullet) getY() int {
	return int(bullet.Y)
}

func (player Player) getX() int {
	return player.X
}

func (player Player) getY() int {
	return player.Y
}

func (player Player) getW() int {
	return player.W
}

func (player Player) getH() int {
	return player.H
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
