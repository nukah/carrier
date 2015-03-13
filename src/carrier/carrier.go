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

type carrierInstance struct {
	ID        string
	Formation map[string]*rpc.Client
	FleetChan *redis.PubSub
}

const (
	RPC_PORT = 25000
)

var (
	_redis       *redis.Client
	_db          *gorm.DB
	SocketsMap   map[*socketio.NameSpace]int
	UsersMap     map[int]map[*socketio.NameSpace]bool
	SocketServer *socketio.SocketIOServer
	_carrier     = &carrierInstance{
		ID:        uuid.Formatter(uuid.NewV4(), uuid.Clean),
		Formation: make(map[string]*rpc.Client),
	}
	_control *rpc.Client
)

func preparationForShutdown() {
	term_channel := make(chan os.Signal, 1)
	signal.Notify(term_channel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGCONT)
	go func() {
		<-term_channel
		log.Printf("(Shutdown) Removing carrier from formation: %s", _carrier.ID)
		_redis.HDel("carriers:formation", _carrier.ID)
		_db.Close()
		_redis.Close()
		os.Exit(1)
	}()
}

func enterSquadron(id string) {
	host := _redis.HGet("carriers:formation", id).Val()
	if host != fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], RPC_PORT) {
		connection, err := rpc.DialHTTP("tcp", host)
		if err != nil {
			log.Printf("(Formation) Error connecting to carrier(%s): %s. Removing invalid carrier from formation.", host, err)
			_redis.HDel("carriers:formation", id)
		} else {
			log.Printf("(Formation) Communication with carrier(%s)@(%s) established.", id, host)
			_carrier.Formation[id] = connection
		}
	}
}

func initializeCarrier() {
	carrierFormation := _redis.HKeys("carriers:formation").Val()
	for _, id := range carrierFormation {
		go enterSquadron(string(id))
	}

	_redis.HSet("carriers:formation", _carrier.ID, fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], RPC_PORT))

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

	_db = &dbConnection

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

func initializeSocketServer() {
	transports := socketio.NewTransportManager()
	transports.RegisterTransport("websocket")

	sioConfig := &socketio.Config{}
	sioConfig.Transports = transports
	sioConfig.ClosingTimeout = 500000
	sioConfig.HeartbeatTimeout = 10000

	SocketServer = socketio.NewSocketIOServer(sioConfig)

	SetupSocketHandlers(SocketServer)
}

func initializeControlConnection() {
	connection, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", viper.GetStringMap("control")["ip"], viper.GetStringMap("control")["port"]))
	if err != nil {
		log.Fatal(fmt.Sprintf("(Carrier) Can not connect to control tower @ %s. Error: %s", viper.GetStringMap("control")["ip"], err))
	}
	log.Printf("(Carrier) Initiating connection to control tower.")
	_control = connection
}

func initializeSubscriber() {
	_carrier.FleetChan = _redis.PubSub()
	messages := make(chan *redis.Message)
	_redis.Publish("formation:fleet", _carrier.ID)
	_carrier.FleetChan.Subscribe("formation:fleet")
	go func() {
		for {
			msg, err := _carrier.FleetChan.Receive()
			if err != nil && msg != nil {
				log.Println("New server")
				messages <- msg.(*redis.Message)
			}
		}
		id := <-messages
		log.Printf("(Formation) New carrier(%s) lifted off. Commencing connection", id)
		go enterSquadron(id.Payload)
	}()
	defer _carrier.FleetChan.Close()
}

func initializeRPC() {
	userRPC := new(UserRPC)
	rpc.Register(userRPC)
	rpc.HandleHTTP()

	rpcHandler, err := net.Listen("tcp", fmt.Sprintf("%s:%d", viper.GetStringMap("rpc")["ip"], viper.GetStringMap("rpc")["port"]))
	if err != nil {
		log.Printf("(Formation) Error while initializing listener: %s", err)
	}
	go http.Serve(rpcHandler, nil)
}

func SetupSocketHandlers(socketServer *socketio.SocketIOServer) {
	socketServer.On("connect", ConnectHandler)
	socketServer.On("authorize", AuthorizationHandler)
	socketServer.On("disconnect", DisconnectionHandler)
}

func Init() {
	var config string
	flag.StringVar(&config, "c", "carrier", "Configuration file")
	flag.Parse()

	viper.SetConfigName(config)
	viper.AddConfigPath("../config")

	if viper.ReadInConfig() != nil {
		log.Fatal("(Configuration) Error while loading configuration")
	}

	initializeDatabase()
	initializeRedis()
	initializeControlConnection()
	initializeCarrier()
	initializeSubscriber()
	initializeSocketServer()

	preparationForShutdown()

	http_server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"]),
		Handler: SocketServer,
	}

	log.Printf("(Carrier) %s lifting up on %s:%d/socketIo", _carrier.ID, viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"])
	log.Printf("(Carrier) RPC Interface: %s:%d", viper.GetStringMap("rpc")["ip"], viper.GetStringMap("rpc")["port"])
	log.Fatal(http_server.ListenAndServe())
}
