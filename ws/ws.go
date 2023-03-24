package ws

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	HandshakeTimeout: time.Second * 5,
	ReadBufferSize:   4096,
	WriteBufferSize:  4096,
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		fmt.Printf("An error occurred while upgrading the http connection: %v", reason)
		http.Error(w, http.StatusText(status), status)
		_, err := w.Write([]byte("error occurred while upgrading the http connection"))
		if err != nil {
			fmt.Printf("Failed to write response: %v", reason)
		}
	},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},

}


func HandleWs( ws*websocket.Conn){
	
}