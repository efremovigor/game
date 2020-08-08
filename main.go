package main

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"
	"try-to-game/lib"
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
			game = &lib.Game{Connection: connections, Width: lib.GameWidth, Height: lib.GameHeight, Bullets: make(map[string]map[[16]byte]*lib.BulletGame), Enemies: make(map[[16]byte]*lib.Enemy)}
			lib.UniGames[playerConnection.SessionId] = game
		}
		lib.Games[playerConnection.SessionId] = game
		connections := game.Connection
		connections[playerConnection.SessionId] = playerConnection

		game.Builds = []lib.Build{lib.Build{X: 100, Y: 100, Width: 300, Height: 200, Type: 1}, lib.Build{X: 400, Y: 300, Width: 100, Height: 300, Type: 1}}
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

	fmt.Print("Listening to 127.0.0.1:3000")
	go lib.RunServer("127.0.0.1:3000", lib.RequestChan)
	for {
		select {
		case request := <-lib.RequestChan:
			handleRequest(request)
		}
	}
}
