package main

import (
	"fmt"
	"try-to-game/lib"
)

func handleRequest(request lib.UserRequest) {
	playerConnection := getPlayConnection(request)
	fmt.Println(playerConnection)
}

func getPlayConnection(request lib.UserRequest) *lib.PlayerConnection {
	var playerConnection, ok = lib.Connections[request.SessionId]
	if !ok {
		playerConnection = &lib.PlayerConnection{Name: request.Request.Payload.Username, Connection: request.Receiver, Command: make(chan string), SessionId: request.SessionId}
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
