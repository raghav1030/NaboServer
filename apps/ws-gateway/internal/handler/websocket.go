package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Upgrade err", err)
		return
	}

	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()

		if err != nil {
			log.Println("Read error: ", err)
			break
		}

		log.Println("Received msg: ", msg)

		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("Write error: ", err)
			break
		}
	}
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	log.Println("Websocket server started at : 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
