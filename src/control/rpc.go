package control

import (
	"log"
)

type UserCallRPCArguments struct {
	User User
	Call Call
}

type UserCallStopArguments struct {
	User           User
	Call           Call
	CallStopReason string
}

type UserRPCArguments struct {
	User User
}

type ControlRPC int

func userOnline(user User) bool {
	var result bool
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (User(%d) userOnline) Carrier ID not retrieved", user.ID)
		return false
	} else {
		call := this.fleet[carrierId].Go("isUserOnline", &UserRPCArguments{user}, result, nil)
		<-call.Done
		return result
	}
	return false
}

func userInCall(user User) bool {
	var result bool
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (User(%d) userInCall) Carrier ID not retrieved", user.ID)
		return false
	} else {
		call := this.fleet[carrierId].Go("isUserInCall", &UserRPCArguments{user}, result, nil)
		<-call.Done
		return result
	}
	return false
}

func makeCallStop(user User, call Call, reason string) {
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callStop)  Carrier ID not retrieved", call.ID, user.ID)
		return
	} else {
		this.fleet[carrierId].Call("CallStop", &UserCallStopArguments{user, call, reason}, nil)
		return
	}
}

func makeCallConnect(user User, call Call) {
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callConnect)  Carrier ID not retrieved", call.ID, user.ID)
		return
	} else {
		this.fleet[carrierId].Call("CallConnect", &UserCallRPCArguments{user, call}, nil)
		return
	}
}
