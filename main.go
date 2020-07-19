package main

import (
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
	Player        lib.Player            `json:"player"`
	Others        map[string]lib.Player `json:"others"`
	PlayerBullets map[string]lib.Bullet `json:"bullets"`
	OthersBullets map[string]lib.Bullet `json:"othersBullets"`
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
		playerConnection.Player = &lib.Player{X: 15, Y: 15, W: 20, H: 20, ID: playerConnection.SessionId, Name: request.Request.Payload.Name, Hp: lib.MaxHp, MaxHp: lib.MaxHp}
		if request.Request.Lobby != "" {
			found, ok := lib.Games[request.Request.Lobby]
			if ok && len(found.Connection) < lib.MaxUserInLobby {
				game = found
			}
		}
		if game == nil {
			connections := make(map[string]*lib.PlayerConnection)
			game = &lib.Game{Connection: connections, Width: lib.GameWidth, Height: lib.GameHeight, Bullets: make(map[string]map[[16]byte]*lib.BulletGame)}
			lib.UniGames[playerConnection.SessionId] = game
		}
		lib.Games[playerConnection.SessionId] = game
		connections := game.Connection
		connections[playerConnection.SessionId] = playerConnection

		response := ResponseInfoState{Type: lib.SignalStartTheGame, Info: ResponseInfoStateInfo{Player: *playerConnection.Player}}
		playerConnection.Connection.PushData(response)

		go func(playerConnection *lib.PlayerConnection, game *lib.Game) {
			defer fmt.Println("sender data -" + playerConnection.SessionId + "closed")
			for {
				time.Sleep(10 * time.Millisecond)
				var others = make(map[string]lib.Player)
				var bullets = make(map[string]lib.Bullet)
				var othersBullets = make(map[string]lib.Bullet)
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
						bulletGame.MoveBullet(connections, sessionId)
						bullet := lib.Bullet{X: bulletGame.Bullet.X, Y: bulletGame.Bullet.Y, Deleted: bulletGame.Deleted}
						if sessionId == playerConnection.SessionId {
							bullets[string(bulletKey[:])] = bullet
						} else {
							othersBullets[string(bulletKey[:])] = bullet
						}
					}
				}
				game.Lock.Unlock()
				response := ResponseInfoState{Type: lib.SignalInfoTheGame, Info: ResponseInfoStateInfo{Player: *playerConnection.Player, Others: others, PlayerBullets: bullets, OthersBullets: othersBullets}}
				playerConnection.Connection.PushData(response)
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
