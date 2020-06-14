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
	Player lib.Player `json:"player"`
}

type ResponseStartGameState struct {
	Type string       `json:"type"`
	Conf ResponseConf `json:"conf"`
}
type ResponseConf struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func handleRequest(request lib.UserRequest) {
	playerConnection := getPlayConnection(request)
	fmt.Println(playerConnection)
	if playerConnection.InGame {
		playerConnection.Command <- request.Request.Payload.Name
		return
	}

	if request.Request.Type == lib.RequestTypeNewPlayer {
		game := &lib.Game{Player: &lib.Player{X: 15, Y: 15, W: 20, H: 20}, Connection: *playerConnection, Weight: 1000, Height: 500}

		response := ResponseStartGameState{Type: lib.SignalStartTheGame, Conf: ResponseConf{Width: game.Weight, Height: game.Height}}
		game.Connection.Connection.PushData(response)

		playerConnection.InGame = true
		go func(game *lib.Game) {
			for {
				time.Sleep(100 * time.Millisecond)
				response := ResponseInfoState{Type: lib.SignalInfoTheGame, Info: ResponseInfoStateInfo{Player: *game.Player}}
				game.Connection.Connection.PushData(response)
			}
		}(game)

		go func(game *lib.Game) {
			for {
				select {
				case command := <-game.Connection.Command:
					commands := strings.Split(command, "-")
					if len(commands) > 0 {
						for _, command := range commands {
							game.Move(command)
						}
					} else {
						game.Move(command)
					}

				}
			}
		}(game)
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
