package api

import (
	"net/http"

	"github.com/go-logr/logr"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func websocketHandler(log logr.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info("Received request", "url", r.URL)
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}

			err = c.WriteMessage(mt, message)
			if err != nil {
				break
			}
		}
	})
}
