package main

import (
	"crypto/md5"
	"fmt"
	"game/lib"
	"strings"
	"time"
)

type ResponseInfoState struct {
	Type string                `json:"type"`
	Info ResponseInfoStateInfo `json:"info"`
}
type ResponseInfoStateInfo struct {
	Player        lib.Player                   `json:"player"`
	Others        map[string]lib.Player        `json:"others"`
	PlayerBullets map[string]lib.Bullet        `json:"bullets"`
	OthersBullets map[string]lib.Bullet        `json:"othersBullets"`
	Builds        []lib.Build                  `json:"builds"`
	Enemies       map[string]lib.EnemyResponse `json:"enemies"`
}

type ResponseStartGameState struct {
	Type string       `json:"type"`
	Conf ResponseConf `json:"conf"`
}
type ResponseConf struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	ID     string `json:"id"`
}

type ResponseLobbyList struct {
	Type    string      `json:"type"`
	Lobbies []LobbyInfo `json:"lobbies"`
}
type LobbyInfo struct {
	Max   int    `json:"max"`
	Free  int    `json:"free"`
	Title string `json:"title"`
	Id    string `json:"id"`
}

func handleRequest(request lib.UserRequest) {
	playerConnection := getPlayConnection(request)
	switch request.Request.Type {
	case lib.RequestInit:
		response := ResponseStartGameState{Type: lib.SignalConfTheGame, Conf: ResponseConf{Width: lib.GameWidth, Height: lib.GameHeight}}
		playerConnection.Connection.PushData(response)
	case lib.RequestTypeLobbyList:
		response := ResponseLobbyList{Type: lib.SignalLobbyList, Lobbies: []LobbyInfo{}}
		for id, game := range lib.UniGames {
			response.Lobbies = append(response.Lobbies, LobbyInfo{Max: lib.MaxUserInLobby, Free: lib.MaxUserInLobby - len(game.Connection), Id: id, Title: lib.Connections[id].Player.Name})
		}
		playerConnection.Connection.PushData(response)
	case lib.RequestTypeNewPlayer:
		var game *lib.Game
		playerConnection.InGame = true
		playerConnection.Player = &lib.Player{X: lib.PlayerStartPositionX, Y: lib.PlayerStartPositionY, W: lib.PlayerWidth, H: lib.PlayerHeight, ID: playerConnection.SessionId, Name: request.Request.Payload.Name, Hp: lib.MaxHp, MaxHp: lib.MaxHp}
		if request.Request.Lobby != "" {
			found, ok := lib.Games[request.Request.Lobby]
			if ok && len(found.Connection) < lib.MaxUserInLobby {
				game = found
			}
		}
		if game == nil {
			connections := make(map[string]*lib.PlayerConnection)
			game = &lib.Game{Connection: connections, Width: lib.GameWidth, Height: lib.GameHeight, Bullets: make(map[string]map[[16]byte]*lib.BulletGame), Enemies: make(map[[16]byte]*lib.Enemy), CrucialPoints: make(map[string]lib.CrucialPoint), CrucialPointsDistance: make(map[string]float64), Builds: []lib.Build{}}
			game.AddBuild(100, 100, 300, 200)
			game.AddBuild(400, 300, 100, 300)
			game.AddCrucialPoint(290, 320)
			game.AddCrucialPoint(365, 415)
			game.AddCrucialPoint(80, 320)
			game.AddCrucialPoint(365, 620)
			game.AddCrucialPoint(80, 80)
			game.AddCrucialPoint(520, 620)
			game.AddCrucialPoint(520, 270)
			game.AddCrucialPoint(420, 80)
			game.AddCrucialPoint(420, 270)

			game.AddSiblingToCrucialPoint(game.GetCrucialPoint(290, 320), game.GetCrucialPoint(365, 415), game.GetCrucialPoint(80, 320), game.GetCrucialPoint(365, 620))
			game.AddSiblingToCrucialPoint(game.GetCrucialPoint(365, 415), game.GetCrucialPoint(290, 320), game.GetCrucialPoint(80, 320), game.GetCrucialPoint(365, 620))
			game.AddSiblingToCrucialPoint(game.GetCrucialPoint(80, 320), game.GetCrucialPoint(290, 320), game.GetCrucialPoint(365, 415), game.GetCrucialPoint(365, 620), game.GetCrucialPoint(80, 80))
			game.AddSiblingToCrucialPoint(game.GetCrucialPoint(365, 620), game.GetCrucialPoint(80, 320), game.GetCrucialPoint(290, 320), game.GetCrucialPoint(365, 415), game.GetCrucialPoint(520, 620))
			game.AddSiblingToCrucialPoint(game.GetCrucialPoint(80, 80), game.GetCrucialPoint(80, 320), game.GetCrucialPoint(420, 80))
			game.AddSiblingToCrucialPoint(game.GetCrucialPoint(520, 620), game.GetCrucialPoint(365, 620), game.GetCrucialPoint(520, 270))
			game.AddSiblingToCrucialPoint(game.GetCrucialPoint(420, 80), game.GetCrucialPoint(80, 80), game.GetCrucialPoint(420, 270), game.GetCrucialPoint(520, 270))
			game.AddSiblingToCrucialPoint(game.GetCrucialPoint(520, 270), game.GetCrucialPoint(520, 620), game.GetCrucialPoint(420, 270), game.GetCrucialPoint(420, 80))
			game.AddSiblingToCrucialPoint(game.GetCrucialPoint(420, 270), game.GetCrucialPoint(520, 270), game.GetCrucialPoint(420, 80))

			lib.UniGames[playerConnection.SessionId] = game
		}
		lib.Games[playerConnection.SessionId] = game
		connections := game.Connection
		connections[playerConnection.SessionId] = playerConnection

		response := ResponseInfoState{Type: lib.SignalStartTheGame, Info: ResponseInfoStateInfo{Player: *playerConnection.Player, Builds: game.Builds}}
		playerConnection.Connection.PushData(response)

		go func(playerConnection *lib.PlayerConnection, game *lib.Game) {
			defer fmt.Println("sender data -" + playerConnection.SessionId + "closed")
			for {
				time.Sleep(10 * time.Millisecond)
				var others = make(map[string]lib.Player)
				var bullets = make(map[string]lib.Bullet)
				var othersBullets = make(map[string]lib.Bullet)
				var enemies = make(map[string]lib.EnemyResponse)
				connections := game.Connection
				for key, connection := range connections {
					if key == playerConnection.SessionId {
						continue
					}
					others[connection.Player.ID] = *connection.Player
				}
				game.Lock.Lock()
				for sessionId, playerBullets := range game.Bullets {
					for bulletKey, bulletGame := range playerBullets {
						bulletGame.MoveBullet(game, sessionId)
						bullet := lib.Bullet{X: bulletGame.Bullet.X, Y: bulletGame.Bullet.Y, Deleted: bulletGame.Deleted}
						if sessionId == playerConnection.SessionId {
							bullets[string(bulletKey[:])] = bullet
						} else {
							othersBullets[string(bulletKey[:])] = bullet
						}
					}
				}

				for id, enemy := range game.Enemies {
					enemies[string(id[:])] = lib.EnemyResponse{X: enemy.X, Y: enemy.Y, W: enemy.W, H: enemy.H, ID: enemy.ID, Name: enemy.Name, Hp: enemy.Hp, MaxHp: enemy.MaxHp}
				}

				if len(game.Enemies) == 0 {
					// генерация врага
				}

				game.Lock.Unlock()
				response := ResponseInfoState{Type: lib.SignalInfoTheGame, Info: ResponseInfoStateInfo{Player: *playerConnection.Player, Others: others, PlayerBullets: bullets, OthersBullets: othersBullets, Enemies: enemies}}
				playerConnection.Connection.PushData(response)
			}
		}(playerConnection, game)
		go func(playerConnection *lib.PlayerConnection, game *lib.Game) {
			enemy := &lib.Enemy{X: 600, Y: 600, W: 10, H: 10, Hp: 100, MaxHp: 100, Path: make(chan lib.Node, 5)}
			game.Enemies[md5.Sum([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))] = enemy

			go func(enemy *lib.Enemy, game *lib.Game) {
				for {
					var nearestToEnemyPoint, nearestToPlayerPoint *lib.NearestCrucialPoint
					//player := &playerConnection.Player
					for _, point := range game.CrucialPoints {
						distanceToEnemy := lib.GetDistance(enemy, point)
						if nearestToEnemyPoint == nil || nearestToEnemyPoint.Distance > distanceToEnemy {
							nearestToEnemyPoint = &lib.NearestCrucialPoint{Distance: distanceToEnemy, CrucialPoint: point}
						}
						distanceToPlayer := lib.GetDistance(playerConnection.Player, point)
						if nearestToPlayerPoint == nil || nearestToPlayerPoint.Distance > distanceToPlayer {
							nearestToPlayerPoint = &lib.NearestCrucialPoint{Distance: distanceToPlayer, CrucialPoint: point}
						}
					}
					visitedPoints := make(map[string]lib.PathCrucialPoint)
					checkingPoints := make(map[string]lib.PathCrucialPoint)

					path := lib.PathCrucialPoint{X: nearestToEnemyPoint.X, Y: nearestToEnemyPoint.Y, Sibling: make(map[string]lib.PathCrucialPoint)}

					visitedPoints[nearestToEnemyPoint.GetKey()] = path
					//for nearestToEnemyPoint.X != nearestToPlayerPoint.X && nearestToEnemyPoint.Y != nearestToPlayerPoint.Y && (*player).X == playerConnection.Player.X && (*player).Y == playerConnection.Player.Y {
					if len(nearestToEnemyPoint.Sibling) > 0 {
						for _, sibling := range nearestToEnemyPoint.Sibling {
							if nearestToPlayerPoint.GetKey() == sibling.GetKey() {
								//todo::arrived
								return
							}
							point := lib.PathCrucialPoint{X: sibling.X, Y: sibling.Y, Distance: path.Distance + game.GetDistance(sibling, path), Sibling: make(map[string]lib.PathCrucialPoint)}
							fmt.Println("Добаляем точку в чекинг лист ", point.GetKey(), " На основе точки", sibling.GetKey())
							path.Sibling[sibling.GetKey()] = point
							checkingPoints[sibling.GetKey()] = point
							fmt.Println("Все точки в поиске", checkingPoints)
						}
					}
					//}
					for len(checkingPoints) > 0 {
						var pointWithMinimalDistance *lib.PathCrucialPoint
						for _, point := range checkingPoints {
							if pointWithMinimalDistance == nil || point.Distance < pointWithMinimalDistance.Distance {
								pointWithMinimalDistance = &point
							}
						}
						visitedPoints[pointWithMinimalDistance.GetKey()] = *pointWithMinimalDistance
						if nearestToPlayerPoint.GetKey() == pointWithMinimalDistance.GetKey() {
							//todo::arrived
							return
						}
						for _, sibling := range game.CrucialPoints[pointWithMinimalDistance.GetKey()].Sibling {
							if nearestToPlayerPoint.GetKey() == sibling.GetKey() {
								//todo::arrived
								return
							}
							point := lib.PathCrucialPoint{X: sibling.X, Y: sibling.Y, Distance: path.Distance + game.GetDistance(sibling, path), Sibling: make(map[string]lib.PathCrucialPoint)}
							path.Sibling[sibling.GetKey()] = point
							checkingPoints[sibling.GetKey()] = point
						}
						delete(checkingPoints, pointWithMinimalDistance.GetKey())
					}
					//asdasdasd
					searching := lib.Searching{ComeFrom: *enemy, Destination: *playerConnection.Player, Builds: game.Builds}
					searching.Handle(playerConnection)
				}
			}(enemy, game)

			go func(enemy *lib.Enemy, game *lib.Game) {
				for {
					time.Sleep(10 * time.Millisecond)
					for i := 0; i < 3; i++ {
						node := <-enemy.Path
						enemy.X = node.X
						enemy.Y = node.Y
					}
				}
			}(enemy, game)

		}(playerConnection, game)

		go func(playerConnection *lib.PlayerConnection) {
			defer fmt.Println("calculate data -" + playerConnection.SessionId + "closed")

			for {
				select {
				case request := <-playerConnection.Request:
					commands := strings.Split(request.Payload.Name, "-")
					if len(commands) > 0 {
						for _, command := range commands {
							if request.Payload.Name == lib.CommandShoot {
								playerConnection.Shoot(game, request.Payload.Bullet)
							} else {
								playerConnection.Move(game, command)
							}
						}
					}
				}
			}
		}(playerConnection)

		go func(game *lib.Game) {
			for {
				time.Sleep(5000 * time.Millisecond)
				game.Lock.Lock()
				for _, playerBullets := range game.Bullets {
					for bulletKey, bulletGame := range playerBullets {
						if bulletGame.Deleted == true {
							delete(playerBullets, bulletKey)
						}
					}
				}
				game.Lock.Unlock()
			}
		}(game)
	case lib.RequestTypeNewCommand:
		playerConnection.Request <- request.Request
	}
}

func getPlayConnection(request lib.UserRequest) *lib.PlayerConnection {
	var playerConnection, ok = lib.Connections[request.SessionId]
	if !ok {
		playerConnection = &lib.PlayerConnection{Connection: request.Receiver, Request: make(chan lib.LoginJsonRequest), SessionId: request.SessionId}
		lib.Connections[request.SessionId] = playerConnection
	}
	return playerConnection
}

func main() {
	lib.RequestChan = make(chan lib.UserRequest)

	fmt.Println("Listening to 127.0.0.1:3000")
	go lib.RunServer("127.0.0.1:3000", lib.RequestChan)
	for {
		select {
		case request := <-lib.RequestChan:
			handleRequest(request)
		}
	}
}
