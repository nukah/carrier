package carrier

import "net/http"
import "flag"
import "fmt"
import "log"
import "gopkg.in/redis.v2"
import "github.com/spf13/cast"
import "github.com/spf13/viper"
import "github.com/Intelity/go-socket.io"
import "github.com/jinzhu/gorm"
import _ "github.com/lib/pq"

var (
	Redis        *redis.Client
	DB           *gorm.DB
	Environment  string
	SocketsMap   map[*socketio.NameSpace]int
	UsersMap     map[int]map[*socketio.NameSpace]bool
	SocketServer *socketio.SocketIOServer
)

func init() {
	flag.StringVar(&Environment, "e", "development", "Configuration environment")
	flag.Parse()
	viper.SetConfigName(Environment)
	viper.AddConfigPath("../config")
	err := viper.ReadInConfig()

	SocketsMap = make(map[*socketio.NameSpace]int)
	UsersMap = make(map[int]map[*socketio.NameSpace]bool)

	if err != nil {
		log.Fatal("Error while loading configuration")
	}
	if !viper.IsSet("redis") {
		log.Fatal("Redis configuration is missing")
	}
	if !viper.IsSet("db") {
		log.Fatal("Database configuration is missing")
	}

	dbConfig := viper.GetStringMap("db")
	db, err := gorm.Open("postgres", fmt.Sprintf("host=%s port=%d dbname=%s password=%s user=%s",
		dbConfig["host"],
		dbConfig["port"],
		dbConfig["database"],
		dbConfig["password"],
		dbConfig["username"],
	))

	DB = &db
	if Environment == "development" {
		DB.LogMode(true)
	}
	err = DB.DB().Ping()

	if err != nil {
		log.Fatal("Error connecting to database (", err, ")")
	}

	redisConfig := viper.GetStringMap("redis")
	redisOptions := &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig["host"], redisConfig["port"]),
		Password: cast.ToString(redisConfig["password"]),
		DB:       int64(cast.ToInt(redisConfig["database"])),
	}

	Redis = redis.NewTCPClient(redisOptions)

	if _, err = Redis.Ping().Result(); err != nil {
		log.Fatal("Error connecting to redis: (", err, ")")
	}

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
	SocketServer.On("api_request", APIHandler)
}

func Init() {
	http_server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"]),
		Handler: SocketServer,
	}

	// http.Handle("/socketIo", socket_server)

	defer DB.Close()
	defer Redis.Close()

	log.Printf("Starting up Carrier on %s:%d/socketIo", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"])
	log.Fatal(http_server.ListenAndServe())

}
