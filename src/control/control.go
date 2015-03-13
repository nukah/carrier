package control

import (
	_ "carrier"
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/nukah/go-socket.io"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"gopkg.in/redis.v2"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type controlInstance struct {
	formation map[string]*rpc.Client
}

var (
	_redis   *redis.Client
	_db      *gorm.DB
	_control = &controlInstance{
		formation: make(map[string]*rpc.Client),
	}
	socketServer *socketio.SocketIOServer
)

func initializeDatabase() {
	if !viper.IsSet("db") {
		log.Fatal("(Configuration) Database configuration is missing")
	}

	dbConfig := viper.GetStringMap("db")
	dbConnection, _ := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d dbname=%s password=%s user=%s connect_timeout=5",
		dbConfig["host"],
		dbConfig["port"],
		dbConfig["database"],
		dbConfig["password"],
		dbConfig["username"],
	))

	_db = &dbConnection
	_db.LogMode(true)

	if _db.DB().Ping() != nil {
		log.Fatal(fmt.Sprintf("(Configuration) Error connecting to database."))
	}
}

func initializeRedis() {
	if !viper.IsSet("redis") {
		log.Fatal("(Configuration) Redis configuration is missing")
	}
	redisConfig := viper.GetStringMap("redis")
	redisOptions := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig["host"], redisConfig["port"]),
		Password: cast.ToString(redisConfig["password"]),
		DB:       int64(cast.ToInt(redisConfig["database"])),
	}

	_redis = redis.NewTCPClient(redisOptions)

	if _, err := _redis.Ping().Result(); err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to redis (%s).", err))
	}
}

func initializeRPC() {
	entity := new(ControlRPC)

	rpc.Register(entity)
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

func initializeControl() {
	carrierFormation := _redis.HGetAllMap("carriers:formation").Val()
	for id, host := range carrierFormation {
		connection, err := rpc.DialHTTP("tcp", host)
		if err != nil {
			log.Printf("(Control) Error connecting to carrier(%s): %s.", host, err)
		} else {
			log.Printf("(Formation) Communication with carrier(%s)@(%s) established.", id, host)
			_control.formation[id] = connection
		}
	}

	transports := socketio.NewTransportManager()
	transports.RegisterTransport("websocket")

	sioConfig := &socketio.Config{}
	sioConfig.Transports = transports
	sioConfig.ClosingTimeout = 50000
	sioConfig.HeartbeatTimeout = 10000

	socketServer = socketio.NewSocketIOServer(sioConfig)

	SetupSocketHandlers(socketServer)
}

func SetupSocketHandlers(socketServer *socketio.SocketIOServer) {
	// socketServer.On("connect", ConnectHandler)
	// socketServer.On("authorize", AuthorizationHandler)
	// socketServer.On("disconnect", DisconnectionHandler)
}

func Init() {
	var config string
	flag.StringVar(&config, "c", "control", "Configuration file")
	flag.Parse()

	viper.SetConfigName(config)
	viper.AddConfigPath("../config")

	if viper.ReadInConfig() != nil {
		log.Fatal("(Configuration) Error while loading configuration")
	}

	http_server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"]),
		Handler: socketServer,
	}
	initializeRedis()
	initializeRPC()
	initializeDatabase()
	initializeControl()

	log.Printf("(Control) Control lifting up on %s:%d/socketIo", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"])
	log.Fatal(http_server.ListenAndServe())

}
