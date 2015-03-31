package carrier

import (
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
	"os"
	"time"
)

const (
	REDIS_USER_SOCKET_SESSION_KEY = "formation:users"
	REDIS_USER_CARRIER_KEY        = "users:carrier:session"
	REDIS_FLEET_CHAT_KEY          = "fleet"
	REDIS_CARRIERS_KEY            = "formation:carriers"
)

type carrierInstance struct {
	id                  string
	carrierFleet        map[string]*rpc.Client
	carrierFleetPubSub  *redis.PubSub
	carrierSocketServer socketio.Server
	redis               *redis.Client
	db                  *gorm.DB
}

func (ci *carrierInstance) shutDown() {
	ci.redis.HDel("formation:carriers", ci.id)
	for client := range SocketsMap {
		go func() {
			if res := removeSocketAuthorization(client); res != "" {
				log.Println(res)
			}
		}()
	}
	ci.db.Close()
	ci.redis.Close()
	os.Exit(1)
}

func (ci *carrierInstance) initDb() {
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

func (ci *carrierInstance) initRedis() {
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

func (ci *carrierInstance) initSocket() {
	transports := []string{"websocket"}
	server, err := socketio.NewServer(transports)
	if err != nil {
		log.Fatal("Socket server failed to initialized")
	}
	server.SetPingTimeout(time.Second * 30)
	server.SetPingInterval(time.Second * 10)
	server.SetMaxConnection(10000)

	ci.carrierSocketServer = *server
	ci.setupSocketHandlers()
}

func (ci *carrierInstance) setupSocketHandlers() {
	this.carrierSocketServer.On("connection", func(ss socketio.Socket) {
		ConnectHandler(ss)

		ss.On("authorize", AuthorizationHandler)
		ss.On("disconnect", DisconnectionHandler)
		ss.On("call_accept", CallAcceptHandler)
		ss.On("call_stop", CallStopHandler)
	})
}

func (ci *carrierInstance) interConnect(id string) {
	carrierHost := ci.redis.HGet(REDIS_CARRIERS_KEY, id).Val()
	if carrierHost != fmt.Sprintf("%s:%d", viper.GetStringMap("rpc")["ip"], viper.GetStringMap("rpc")["port"]) {
		conn, err := rpc.DialHTTP("tcp", carrierHost)
		if err != nil {
			log.Printf("(Formation) Error connecting to carrier(%s): %s. Removing invalid carrier from formation.", carrierHost, err)
			ci.redis.HDel(REDIS_CARRIERS_KEY, id)
		} else {
			log.Printf("(Formation) Communication with carrier(%s)@(%s) established.", id, carrierHost)
			ci.carrierFleet[id] = conn
		}
	}
}

func (ci *carrierInstance) startRPC() {
	carrierRPC := new(CarrierRPC)
	rpc.Register(carrierRPC)
	rpc.HandleHTTP()

	rpcHandler, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetStringMap("rpc")["ip"], viper.GetStringMap("rpc")["port"]))
	if err != nil {
		log.Printf("(Formation) Error while initializing listener: %s", err)
	}
	this.redis.HSet("formation:carriers", this.id, fmt.Sprintf("%s:%d", viper.GetStringMap("rpc")["ip"], viper.GetStringMap("rpc")["port"]))
	go func() {
		log.Fatal(http.Serve(rpcHandler, nil))
	}()
}
