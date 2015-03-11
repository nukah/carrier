package carrier

import (
	"flag"
	"fmt"
	"github.com/Intelity/go-socket.io"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/twinj/uuid"
	"gopkg.in/redis.v2"
	"log"
	"net/http"
	"time"
)

type CarrierInstance struct {
	ID string
}

func NewCarrier() *CarrierInstance {
	return &CarrierInstance{
		ID: uuid.Formatter(uuid.NewV4(), uuid.Clean),
	}
}

var (
	Redis        *redis.Client
	DB           *gorm.DB
	Environment  string
	SocketsMap   map[*socketio.NameSpace]int
	UsersMap     map[int]map[*socketio.NameSpace]bool
	SocketServer *socketio.SocketIOServer
	Carrier      = NewCarrier()
)

func init() {

	// Environment parsing from CLI
	//

	flag.StringVar(&Environment, "e", "development", "Configuration environment")
	flag.Parse()

	viper.SetConfigName(Environment)
	viper.AddConfigPath("../config")

	if viper.ReadInConfig() != nil {
		log.Fatal("Error while loading configuration")
	}

	SocketsMap = make(map[*socketio.NameSpace]int)
	UsersMap = make(map[int]map[*socketio.NameSpace]bool)

	if !viper.IsSet("redis") {
		log.Fatal("Redis configuration is missing")
	}

	if !viper.IsSet("db") {
		log.Fatal("Database configuration is missing")
	}

	// Database initialization
	//

	dbConfig := viper.GetStringMap("db")
	dbConnection, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d dbname=%s password=%s user=%s connect_timeout=5",
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
		log.Fatal(fmt.Sprintf("Error connecting to database."))
	}

	// Redis initialization
	//
	redisConfig := viper.GetStringMap("redis")
	redisOptions := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig["host"], redisConfig["port"]),
		Password: cast.ToString(redisConfig["password"]),
		DB:       int64(cast.ToInt(redisConfig["database"])),
	}

	Redis = redis.NewTCPClient(redisOptions)

	if _, err = Redis.Ping().Result(); err != nil {
		log.Fatal(fmt.Sprintf("Error connecting to redis (%s).", err))
	}

	// Transport server initialization
	//

	transports := socketio.NewTransportManager()

	for _, v := range cast.ToStringSlice(viper.GetStringMap("sockets")["transports"]) {
		transports.RegisterTransport(v)
	}

	sioConfig := &socketio.Config{}
	sioConfig.Transports = transports
	sioConfig.ClosingTimeout = 10000
	sioConfig.HeartbeatTimeout = 10000

	SocketServer = socketio.NewSocketIOServer(sioConfig)

	SocketServer.On("authorize", AuthorizationHandler)
	SocketServer.On("disconnect", DisconnectionHandler)
}

func Init() {
	http_server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"]),
		Handler: SocketServer,
	}

	// http.Handle("/socketIo", socket_server)
	log.Println("Carrier Id:", Carrier.ID)
	defer DB.Close()
	defer Redis.Close()
	HubSubscribe()
	log.Printf("Starting up Carrier on %s:%d/socketIo", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"])
	log.Fatal(http_server.ListenAndServe())

}

func HubSubscribe() {
	subtimeout := time.NewTicker(time.Millisecond * 500)
	go func() {
		for range subtimeout.C {
			//instruction := Redis.LPop("formation:hub:commands")
			log.Println("Tick")
			//ProcessInstructions(instruction)
		}
	}()
}
