package control

import (
	"log"
)

type CarrierCallRPC struct {
	User User
	Call Call
}

type CarrierCallStringRPC struct {
	User  User
	Call  Call
	Value string
}

type CarrierCallBoolRPC struct {
	User  User
	Call  Call
	Value bool
}

type CarrierUserRPC struct {
	User User
}

func userOnline(user User) bool {
	var result bool
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (User(%d) userOnline) Carrier ID not retrieved", user.ID)
		return false
	} else {
		call := this.fleet[carrierId].Go("isUserOnline", &CarrierUserRPC{user}, result, nil)
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
		call := this.fleet[carrierId].Go("isUserInCall", &CarrierUserRPC{user}, result, nil)
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
		this.fleet[carrierId].Call("CallStop", &CarrierCallStringRPC{user, call, reason}, nil)
		return
	}
}

func makeCallConnect(user User, call Call) {
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callConnect)  Carrier ID not retrieved", call.ID, user.ID)
		return
	} else {
		this.fleet[carrierId].Call("CallConnect", &CarrierCallRPC{user, call}, nil)
		return
	}
}

func makeCallStart(user User, call Call) {
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callStart)  Carrier ID not retrieved", call.ID, user.ID)
		return
	} else {
		this.fleet[carrierId].Call("CallStart", &CarrierCallRPC{user, call}, nil)
		return
	}
}

func makeCallFinish(user User, call Call) {
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callFinish)  Carrier ID not retrieved", call.ID, user.ID)
		return
	} else {
		this.fleet[carrierId].Call("CallFinish", &CarrierCallRPC{user, call}, nil)
		return
	}
}

func makeCallAnswer(user User, call Call, decision bool) {
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callAnswer)  Carrier ID not retrieved", call.ID, user.ID)
		return
	} else {
		this.fleet[carrierId].Call("CallAnswer", &CarrierCallBoolRPC{user, call, decision}, nil)
		return
	}
}

func makeCallReveal(user User, call Call, decision bool) {
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callReveal)  Carrier ID not retrieved", call.ID, user.ID)
		return
	} else {
		this.fleet[carrierId].Call("CallReveal", &CarrierCallBoolRPC{user, call, decision}, nil)
		return
	}
}
