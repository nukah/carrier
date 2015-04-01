package control

import (
	"carrier"
	"fmt"
	"github.com/googollee/go-socket.io"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"gopkg.in/redis.v2"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
)

type controlInstance struct {
	fleet               map[string]*rpc.Client
	redis               *redis.Client
	db                  *gorm.DB
	controlSocketServer socketio.Server
	calls               map[int]*Call
	mutex               sync.RWMutex
}

func (ci *controlInstance) initDb() {
	if !viper.IsSet("db") {
		log.Fatal("(Configuration) Database configuration is missing")
	}

	dbconf := viper.GetStringMap("db")
	dbconn, _ := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d dbname=%s password=%s user=%s connect_timeout=5",
		dbconf["host"],
		dbconf["port"],
		dbconf["database"],
		dbconf["password"],
		dbconf["username"],
	))

	ci.db = &dbconn
	ci.db.LogMode(true)

	if ci.db.DB().Ping() != nil {
		log.Fatal(fmt.Sprintf("(Configuration) Error connecting to database."))
	}
}

func (ci *controlInstance) initRedis() {
	if !viper.IsSet("redis") {
		log.Fatal("(Configuration) Redis configuration is missing")
	}
	redisconf := viper.GetStringMap("redis")
	redisopt := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisconf["host"], redisconf["port"]),
		Password: cast.ToString(redisconf["password"]),
		DB:       int64(cast.ToInt(redisconf["database"])),
	}

	ci.redis = redis.NewTCPClient(redisopt)

	if _, err := ci.redis.Ping().Result(); err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to redis (%s).", err))
	}
}

func (ci *controlInstance) handleFleet() {
	fleet := ci.redis.PubSub()
	err := fleet.Subscribe(carrier.REDIS_FLEET_CHAT_KEY)
	if err != nil {
		log.Printf("(Fleet) Subscribe to pubsub failed: %s", err)
	}
	go func() {
		for {
			in, err := fleet.Receive()
			if err != nil {
				log.Printf("(Fleet) pubsub error: %s", err)
			}
			switch t := in.(type) {
			case *redis.Message:
				carrierHost := ci.redis.HGet(carrier.REDIS_CARRIERS_KEY, t.Payload).Val()
				conn, err := rpc.DialHTTP("tcp", carrierHost)
				if err != nil {
					log.Printf("(Fleet) Error connecting to carrier(%s): %s.", t.Payload, err)
				} else {
					log.Printf("(Fleet) New carrier is on uplink: %s", t.Payload)
					ci.fleet[t.Payload] = conn
				}
			}
			time.Sleep(time.Second)
		}
	}()
}

func (ci *controlInstance) startRPC() {
	controlRPC := new(ControlRPC)

	rpc.Register(controlRPC)
	rpc.HandleHTTP()

	rpcHandler, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetStringMap("rpc")["ip"], viper.GetStringMap("rpc")["port"]))
	if err != nil {
		log.Printf("(Control) Error while initializing RPC: %s", err)
	}
	go func() {
		log.Fatal(http.Serve(rpcHandler, nil))
	}()
}

func (ci *controlInstance) initSocket() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal("Socket server failed to initialized")
	}
	server.SetPingTimeout(time.Second * 60)
	server.SetPingInterval(time.Second * 30)
	server.SetMaxConnection(500)

	ci.controlSocketServer = *server
	ci.setupSocketHandlers()
}

func (ci *controlInstance) setupSocketHandlers() {
	ci.controlSocketServer.On("connection", func(ss socketio.Socket) {
		ConnectHandler(ss)

		ss.On("call_init", CallInitHandler)
	})
}
