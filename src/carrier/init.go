package carrier

import (
	"flag"
	"fmt"
	"github.com/googollee/go-socket.io"
	"github.com/spf13/viper"
	"github.com/twinj/uuid"
	"gopkg.in/redis.v2"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	RPC_PORT = 25000
)

var (
	SocketsMap = make(map[socketio.Socket]int)
	UsersMap   = make(map[int]map[socketio.Socket]bool)
	this       = &carrierInstance{
		id:           uuid.Formatter(uuid.NewV4(), uuid.Clean),
		carrierFleet: make(map[string]*rpc.Client),
	}
	control *rpc.Client
)

func preparationForShutdown() {
	shutdownChannel := make(chan os.Signal, 1)
	signal.Notify(shutdownChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGCONT)
	go func() {
		<-shutdownChannel
		log.Printf("(Shutdown) Removing carrier from formation: %s", this.id)
		this.shutDown()
	}()
}

func handleFleet() {
	fleet := this.redis.PubSub()
	this.redis.Publish("fleet", this.id)
	err := fleet.Subscribe("fleet")
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
				this.interConnect(t.Payload)
			}
			time.Sleep(time.Second)
		}
	}()
}

func Start() {
	var config string
	flag.StringVar(&config, "c", "carrier", "Configuration file")
	flag.Parse()

	viper.SetConfigName(config)
	viper.AddConfigPath("../config")

	if viper.ReadInConfig() != nil {
		log.Fatal("(Configuration) Error while loading configuration")
	}

	this.initDb()
	this.initRedis()
	this.initSocket()
	this.startRPC()

	carriersOnline := this.redis.HKeys("formation:carriers").Val()
	for _, id := range carriersOnline {
		go this.interConnect(id)
	}

	this.redis.HSet("formation:carriers", this.id, fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], RPC_PORT))

	c, err := rpc.DialHTTP("tcp", fmt.Sprintf("%s:%d", viper.GetStringMap("control")["ip"], viper.GetStringMap("control")["port"]))

	control = c

	if err != nil {
		log.Fatal(fmt.Sprintf("(Carrier) Can not connect to control tower @ %s. Error: %s", viper.GetStringMap("control")["ip"], err))
	}

	handleFleet()
	preparationForShutdown()

	http.Handle("/socket.io/", &this.carrierSocketServer)

	log.Printf("(Carrier) %s lifting up on %s:%d/socket.io", this.id, viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"])
	log.Printf("(Carrier) RPC Interface: %s:%d", viper.GetStringMap("rpc")["ip"], viper.GetStringMap("rpc")["port"])
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"]), nil))
}
