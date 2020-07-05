package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
	"try-to-game/lib"
)

type ResponseInfoState struct {
	Type string                `json:"type"`
	Info ResponseInfoStateInfo `json:"info"`
}
type ResponseInfoStateInfo struct {
	Player lib.Player            `json:"player"`
	Others map[string]lib.Player `json:"others"`
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
	Id    string `json:"title"`
}

func handleRequest(request lib.UserRequest) {
	playerConnection := getPlayConnection(request)
	switch request.Request.Type {
	case lib.RequestInit:
		response := ResponseStartGameState{Type: lib.SignalConfTheGame, Conf: ResponseConf{Width: lib.GameWidth, Height: lib.GameHeight}}
		playerConnection.Connection.PushData(response)
	case lib.RequestTypeNewCommand:
		playerConnection.Command <- request.Request.Payload.Name
	case lib.RequestTypeLobbyList:
		response := ResponseLobbyList{Type: lib.RequestTypeLobbyList, Lobbies: []LobbyInfo{}}
		for id, game := range lib.Games {
			response.Lobbies = append(response.Lobbies, LobbyInfo{Max: lib.MaxUserInLobby, Free: lib.MaxUserInLobby - len(game.Connection), Id: id, Title: lib.Connections[id].Name})
		}
		playerConnection.Connection.PushData(response)
	case lib.RequestTypeNewPlayer:
		var game *lib.Game
		playerConnection.InGame = true
		playerConnection.Player = &lib.Player{X: 15, Y: 15, W: 20, H: 20, ID: playerConnection.SessionId}
		if len(lib.Games) > 0 {
			keys := reflect.ValueOf(lib.Games).MapKeys()
			game = lib.Games[keys[rand.Intn(len(keys))].Interface().(string)]
		} else {
			connections := make(map[string]*lib.PlayerConnection)
			game = &lib.Game{Connection: connections, Width: lib.GameWidth, Height: lib.GameHeight}
		}
		lib.Games[playerConnection.SessionId] = game
		connections := game.Connection
		connections[playerConnection.SessionId] = playerConnection

		response := ResponseStartGameState{Type: lib.SignalStartTheGame}
		playerConnection.Connection.PushData(response)

		go func(playerConnection *lib.PlayerConnection, game *lib.Game) {
			defer fmt.Println("sender data -" + playerConnection.SessionId + "closed")
			for {
				time.Sleep(10 * time.Millisecond)
				var others = make(map[string]lib.Player)
				connections := game.Connection
				fmt.Println("count connections:", len(connections))
				for key, connection := range connections {
					fmt.Println("connection: ", playerConnection.SessionId)

					if key == playerConnection.SessionId {
						continue
					}
					others[connection.Player.ID] = *connection.Player
				}
				response := ResponseInfoState{Type: lib.SignalInfoTheGame, Info: ResponseInfoStateInfo{Player: *playerConnection.Player, Others: others}}
				playerConnection.Connection.PushData(response)
			}
		}(playerConnection, game)

		go func(playerConnection *lib.PlayerConnection) {
			defer fmt.Println("calculate data -" + playerConnection.SessionId + "closed")

			for {
				select {
				case command := <-playerConnection.Command:
					commands := strings.Split(command, "-")
					if len(commands) > 0 {
						for _, command := range commands {
							playerConnection.Move(*game, command)
						}
					} else {
						playerConnection.Move(*game, command)
					}
				}
			}
		}(playerConnection)
	}
}

func getPlayConnection(request lib.UserRequest) *lib.PlayerConnection {
	var playerConnection, ok = lib.Connections[request.SessionId]
	if !ok {
		playerConnection = &lib.PlayerConnection{Name: request.Request.Payload.Name, Connection: request.Receiver, Command: make(chan string), SessionId: request.SessionId}
		lib.Connections[request.SessionId] = playerConnection
	}
	return playerConnection
}

func main() {
	lib.RequestChan = make(chan lib.UserRequest)

	go lib.RunServer("127.0.0.1:3000", lib.RequestChan)
	for {
		select {
		case request := <-lib.RequestChan:
			handleRequest(request)
		}
	}
}
