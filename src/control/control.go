package control

import (
	_ "carrier"
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
	"time"
)

type controlInstance struct {
	fleet               map[string]*rpc.Client
	redis               *redis.Client
	db                  *gorm.DB
	controlSocketServer socketio.Server
	calls               map[string]*Call
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

	this.db = &dbconn
	this.db.LogMode(true)

	if this.db.DB().Ping() != nil {
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

	this.redis = redis.NewTCPClient(redisopt)

	if _, err := this.redis.Ping().Result(); err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to redis (%s).", err))
	}
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
		log.Printf("(Control) RPC Interface: %s:%d", viper.GetStringMap("rpc")["ip"], viper.GetStringMap("rpc")["port"])
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
	ss := this.controlSocketServer

	ss.On("connect", ConnectHandler)
	ss.On("call_init", CallInitHandler)
}
