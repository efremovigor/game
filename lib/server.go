package lib

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"reflect"
)

var upgrader = websocket.Upgrader{}
var store = sessions.NewFilesystemStore("./session", []byte("MTU4MjQ0NTc0NnxEdi1CQkFFQ180SU"))
var RequestChan chan UserRequest

type UserRequest struct {
	SessionId string
	Request   LoginJsonRequest
	Receiver  *ConnectionReceiver
}

type LoginJsonRequest struct {
	Type    string  `json:"type"`
	Lobby   string  `json:"lobby"`
	Payload Payload `json:"payload"`
}

type Payload struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Bullet Bullet `json:"bullet"`
}

type Bullet struct {
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Deleted bool    `json:"deleted"`
}

type ConnectionReceiver struct {
	conn         *websocket.Conn
	readChannel  chan string
	WriteChannel chan []byte
	closeConnect chan int
}

func (receiver ConnectionReceiver) PushData(jsonObject interface{}) {
	responseJson, err := json.Marshal(jsonObject)
	if err != nil {
		fmt.Println(err)
	}
	receiver.WriteChannel <- responseJson
}

func getSession(r *http.Request) (session *sessions.Session) {
	session, _ = store.Get(r, "session-Name")
	return
}

func webHandler(w http.ResponseWriter, _ *http.Request) {
	tmpl, _ := template.ParseFiles("./static/index.html")
	if err := tmpl.Execute(w, ""); err != nil {
		log.Fatalf("404: %v", err)
	}
}

func webSocketHandler(w http.ResponseWriter, r *http.Request) {
	session := getSession(r)
	if session.IsNew {
		if err := session.Save(r, w); err != nil {
			panic(err)
		}
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	receiver := &ConnectionReceiver{conn: conn, readChannel: make(chan string), WriteChannel: make(chan []byte), closeConnect: make(chan int)}

	go func() {

		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		go func(receiver *ConnectionReceiver) {
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					receiver.closeConnect <- 0
					break
				}
				receiver.handleRequest(message, session)
			}
		}(receiver)

		go func(receiver *ConnectionReceiver) {
			for {
				err := conn.WriteMessage(1, <-receiver.WriteChannel)
				if err != nil {
					receiver.closeConnect <- 0
					break
				}
			}
		}(receiver)

		for {
			select {
			case <-receiver.closeConnect:
				game, ok := Games[session.ID]
				if ok {
					fmt.Println("remove connection: ", session.ID)
					delete(game.Connection, session.ID)
					if len(game.Connection) == 0 {
						delete(Games, session.ID)
						delete(UniGames, session.ID)
					} else if _, ok := Games[session.ID]; ok {
						gameKeys := reflect.ValueOf(game.Connection).MapKeys()
						newMainPlayer := gameKeys[rand.Intn(len(gameKeys))].Interface()
						Games[newMainPlayer.(string)] = game
						UniGames[newMainPlayer.(string)] = game
						delete(Games, session.ID)
						delete(UniGames, session.ID)
						fmt.Println("Mainer changed from " + session.ID + " to " + newMainPlayer.(string))
					}
				}
				if _, ok := Connections[session.ID]; ok {
					delete(Connections, session.ID)
				}
				break
			}
		}
	}()
}

func (receiver *ConnectionReceiver) handleRequest(message []byte, session *sessions.Session) {
	var request LoginJsonRequest
	err := json.Unmarshal(message, &request)
	if err != nil {
		fmt.Println("unreadable message -" + string(message))
		return
	}
	RequestChan <- UserRequest{Request: request, SessionId: session.ID, Receiver: receiver}
}

func RunServer(socket string, requests chan UserRequest) {
	RequestChan = requests

	router := mux.NewRouter()
	router.HandleFunc("/", webHandler)
	router.HandleFunc("/ws", webSocketHandler)
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	server := http.Server{
		Addr:    socket,
		Handler: router,
	}
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe(): %v", err)
	}
}
