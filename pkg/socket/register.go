package socket

import (
	socketio "github.com/googollee/go-socket.io"
	"github.com/gorilla/mux"
	"log"
	"waitingroom/internal/infra/container"
	"waitingroom/internal/infra/jwt"
)

func Register(r *mux.Router) {
	container.GetSocket().OnConnect("/", func(s socketio.Conn) error {
		log.Println("Conectado:", s.ID())
		return nil
	})

	container.GetSocket().OnEvent("/", "join", func(s socketio.Conn, token string) {
		claims, _ := jwt.VerifyToken(token)
		s.Join(claims.SessionID)
	})

	container.GetSocket().OnError("/", func(s socketio.Conn, e error) {
		log.Println("Erro:", e)
	})

	container.GetSocket().OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("Saiu:", s.ID(), reason)
	})

	r.Handle("/socket.io/", container.GetSocket())

	go func() {
		err := container.GetSocket().Serve()
		if err != nil {
			log.Println("error:", err)
		}
	}()
}
