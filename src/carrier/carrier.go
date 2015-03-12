package carrier

import (
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/nukah/go-socket.io"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/twinj/uuid"
	"gopkg.in/redis.v2"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
)

type CarrierInstance struct {
	ID        string
	Formation map[string]*rpc.Client
}

const (
	RPC_PORT = 25000
)

var (
	Redis        *redis.Client
	DB           *gorm.DB
	Environment  string
	SocketsMap   map[*socketio.NameSpace]int
	UsersMap     map[int]map[*socketio.NameSpace]bool
	SocketServer *socketio.SocketIOServer
	Carrier      *CarrierInstance
)

func init() {
	flag.StringVar(&Environment, "e", "development", "Configuration environment")
	flag.Parse()

	viper.SetConfigName(Environment)
	viper.AddConfigPath("../config")

	if viper.ReadInConfig() != nil {
		log.Fatal("(Configuration) Error while loading configuration")
	}
}

func preparationForShutdown() {
	term_channel := make(chan os.Signal, 1)
	signal.Notify(term_channel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGCONT)
	go func() {
		<-term_channel
		log.Printf("(Shutdown) Removing carrier from formation: %s", Carrier.ID)
		DB.Close()
		Redis.Close()
		os.Exit(1)
	}()
}

func initializeCarrier() {
	carrierFormation := Redis.HGetAllMap("carriers:formation").Val()
	for id, host := range carrierFormation {
		if host != fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], RPC_PORT) {
			connection, err := rpc.DialHTTP("tcp", host)
			if err != nil {
				log.Printf("(Formation) Error connecting to carrier(%s): %s. Removing invalid carrier from formation.", host, err)
				Redis.HDel("carriers:formation", id)
			} else {
				Carrier.Formation[id] = connection
			}
		}
	}

	userRPC := new(UserRPC)
	rpc.Register(userRPC)
	rpc.HandleHTTP()

	rpcHandler, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], RPC_PORT))
	if err != nil {
		log.Printf("(Formation) Error while initializing listener: %s", err)
	}
	go http.Serve(rpcHandler, nil)

	Redis.HSet("carriers:formation", Carrier.ID, fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], RPC_PORT))

	SocketsMap = make(map[*socketio.NameSpace]int)
	UsersMap = make(map[int]map[*socketio.NameSpace]bool)
}

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

	DB = &dbConnection

	if Environment == "development" {
		DB.LogMode(true)
	}

	if DB.DB().Ping() != nil {
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

	Redis = redis.NewTCPClient(redisOptions)

	if _, err := Redis.Ping().Result(); err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to redis (%s).", err))
	}
}

func initializeSocketServer() {
	transports := socketio.NewTransportManager()

	for _, v := range cast.ToStringSlice(viper.GetStringMap("sockets")["transports"]) {
		transports.RegisterTransport(v)
	}

	sioConfig := &socketio.Config{}
	sioConfig.Transports = transports
	sioConfig.ClosingTimeout = 50
	sioConfig.HeartbeatTimeout = 10000

	SocketServer = socketio.NewSocketIOServer(sioConfig)

	SetupSocketHandlers(SocketServer)
}

func SetupSocketHandlers(socketServer *socketio.SocketIOServer) {
	socketServer.On("connect", ConnectHandler)
	socketServer.On("authorize", AuthorizationHandler)
	socketServer.On("disconnect", DisconnectionHandler)
}

func Init() {
	Carrier = &CarrierInstance{
		ID:        uuid.Formatter(uuid.NewV4(), uuid.Clean),
		Formation: make(map[string]*rpc.Client),
	}

	initializeDatabase()
	initializeRedis()
	initializeCarrier()
	initializeSocketServer()

	preparationForShutdown()

	http_server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"]),
		Handler: SocketServer,
	}

	log.Printf("(Startup) Carrier(%s) firing up on %s:%d/socketIo", Carrier.ID, viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"])
	log.Fatal(http_server.ListenAndServe())
}
