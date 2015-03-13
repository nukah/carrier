package control

import (
	"log"
)

type UserCallRPCArguments struct {
	User User
	Call Call
}

type UserRPCArguments struct {
	User User
}

func userOnline(user User) bool {
	var result bool
	online := make(chan bool)
	carrierId := _redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) Carrier ID not retrieved")
		return false
	}
	_control.formation[carrierId].Go("isUserOnline", &UserRPCArguments{user}, result, nil)
	result = <-online
	return result
}

func userInCall(user User) bool {
	var result bool
	inCall := make(chan bool)
	carrierId := _redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) Carrier ID not retrieved")
		return false
	}
	_control.formation[carrierId].Go("isUserInCall", &UserRPCArguments{user}, result, nil)
	result = <-inCall
	return result
}

func makeCallConnect(user User, call Call) {
	carrierId := _redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) Carrier ID not retrieved")
	}
	_control.formation[carrierId].Call("CallConnect", &UserCallRPCArguments{user, call}, nil)
}

type ControlRPC int
