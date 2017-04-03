package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
)

type commandStruct struct {
	Command string
}

var userAuthStore = make(map[string]string)

var txWatcherRedisConn redis.Conn

var reqCounter uint64

func pushHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	cmd := struct{ Command string }{""}
	err := decoder.Decode(&cmd)

	failOnError(err, "Decoding Failed")

	//Validate Command
	//Add it to Redis
	redisCmd := fmt.Sprintf("[%d] %s\n", reqCounter, cmd.Command)
	_, err = txWatcherRedisConn.Do("RPUSH", pendingTxKey, redisCmd)

	failOnError(err, "Failed to push to worker queue") // This will result in requests hanging
	if err != nil {
		// Apparently err is nil even on good pushes. #ThanksRedis
		fmt.Fprintf(w, "Action %s was not completed.", cmd.Command)
		return
	}
	reqCounter++
	fmt.Fprintf(w, "Successfully performed %s\n", cmd.Command)
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	cmd := struct {
		Pass string
		User string
	}{"", ""}
	err := decoder.Decode(&cmd)
	failOnError(err, "Failed to decode request")
	pass, found := userAuthStore[cmd.User]
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "User %s does not exist", cmd.User) // SET THE RIGHT STATUS CODES!
		return
	}
	if pass != cmd.Pass {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Password for User %s is incorrect", cmd.User)
		return
	}
	//fmt.Fprint(w, `{"Result":"Alice","Body":"Hello","Time":1294706395881547000}`)
	fmt.Fprintf(w, "User %s successfully logged in", cmd.User)
}

func createHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	cmd := struct {
		Pass string
		User string
	}{"", ""}
	err := decoder.Decode(&cmd)

	failOnError(err, "Decoding Failed")
	_, found := userAuthStore[cmd.User]
	if found {
		// user already exists
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "User %s already exists", cmd.User)
		return
	}
	userAuthStore[cmd.User] = cmd.Pass
	fmt.Fprint(w, "Success")
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var userSocketmap = make(map[string]*websocket.Conn)

var count int

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// frontend handshake to get user and hook them into the userMap for sockets
	_, message, err := conn.ReadMessage()
	failOnError(err, "Failed to handshake")
	fmt.Printf("Handshake from client is %s\n", message)
	userSocket, found := userSocketmap[string(message)]
	if found {
		userSocket.Close()
	}
	userSocketmap[string(message)] = conn
	greeting := fmt.Sprintf("Hello %d\n", count)
	bye := fmt.Sprintf("Goodbye %d\n", count)
	conn.WriteMessage(websocket.TextMessage, []byte(greeting))
	conn.WriteMessage(websocket.TextMessage, []byte(bye))
	fmt.Println(userSocketmap)
	// _, message, err = conn.ReadMessage()
	// failOnError(err, "Failed to offshake")
	// fmt.Printf("Failout from client is %s\n", message)
	count++
}

func incomingTxWatcher() {

	txWatcherRedisConn = redisPool.Get()
	port := fmt.Sprintf(":%d", config.WebSocketPort)
	fmt.Printf("Started watching on port %s\n", port)
	http.HandleFunc("/push", pushHandler)
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/create", createHandler)
	http.HandleFunc("/ws", wsHandler)
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.ListenAndServe(port, nil)
}
