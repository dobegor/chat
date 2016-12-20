package main

import (
	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/dobegor/chat/server"
	"github.com/gorilla/websocket"
	"github.com/pressly/chi"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	log.SetHandler(cli.Default)

	h := &server.Hub{}

	h.Init()

	r := chi.NewRouter()

	r.Get("/chat/:username", func(w http.ResponseWriter, r *http.Request) {
		username := chi.URLParam(r, "username")

		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.WithFields(log.Fields{
				"username": username,
				"err":      err.Error(),
				"time":     time.Now().String(),
			}).Error("Error upgrading connection")

			return
		}

		h.AddConnection(username, conn)

	})

	log.Info("Starting server")

	http.ListenAndServe(":9090", r)
}
