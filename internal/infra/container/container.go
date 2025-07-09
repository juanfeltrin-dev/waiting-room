package container

import (
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
	"github.com/redis/go-redis/v9"
	"net/http"
	"waitingroom/internal/infra/cache"
)

var (
	container       *Container
	secretKey       = []byte("secret-key")
	allowOriginFunc = func(r *http.Request) bool {
		return true
	}
)

type Container struct {
	cache  *redis.Client
	socket *socketio.Server
}

func Start() {
	container = &Container{
		cache: cache.NewRedisCache(),
		socket: socketio.NewServer(&engineio.Options{
			Transports: []transport.Transport{
				&polling.Transport{
					CheckOrigin: allowOriginFunc,
				},
				&websocket.Transport{
					CheckOrigin: allowOriginFunc,
				},
			},
		}),
	}
}

func GetContainer() *Container {
	return container
}

func GetCache() *redis.Client {
	return container.cache
}

func GetSocket() *socketio.Server {
	return container.socket
}

func GetSecretKey() []byte {
	return secretKey
}
