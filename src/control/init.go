package control

import (
	_ "carrier"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"net/rpc"
)

var (
	this = &controlInstance{
		fleet: make(map[string]*rpc.Client),
		calls: make(map[int]*Call),
	}
)

func Init() {
	var config string
	flag.StringVar(&config, "c", "control", "Configuration file")
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
	this.handleFleet()

	carriersOnline := this.redis.HGetAllMap("formation:carriers").Val()
	for id, host := range carriersOnline {
		conn, err := rpc.DialHTTP("tcp", host)
		if err != nil {
			log.Printf("(Control) Error connecting to carrier(%s): %s.", host, err)
		} else {
			log.Printf("(Formation) Communication with carrier(%s)@(%s) established.", id, host)
			this.fleet[id] = conn
		}
	}

	http.Handle("/socket.io/", &this.controlSocketServer)

	log.Printf("(Control) Control lifting up on %s:%d/socket.io/", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"])
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", viper.GetStringMap("sockets")["ip"], viper.GetStringMap("sockets")["port"]), nil))
}
