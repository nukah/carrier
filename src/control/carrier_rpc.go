package control

import (
	"log"
	"net/rpc"
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

type CarrierCallAftermathRPC struct {
	User   User
	Call   Call
	Action string
	Type   string
}

type CarrierUserRPC struct {
	User User
}

type CarrierMessageRPC struct {
	Destination User
	Message     Message
}

func userOnline(user User) bool {
	var result bool
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (User(%d) userOnline) Carrier ID not retrieved", user.ID)
		return false
	}
	this.fleet[carrierId].Go("CarrierRPC.IsUserOnline", &CarrierUserRPC{user}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) UserOnline: ", err.Error)
	}
	return result
}

func userInCall(user User) bool {
	var result bool
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (User(%d) userInCall) Carrier ID not retrieved", user.ID)
		return false
	}
	this.fleet[carrierId].Go("CarrierRPC.IsUserInCall", &CarrierUserRPC{user}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) UserInCall: ", err.Error)
	}
	return result
}

func makeCallStop(user User, call Call, reason string) {
	var result error
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callStop)  Carrier ID not retrieved", call.ID, user.ID)
		return
	}
	this.fleet[carrierId].Go("CarrierRPC.CallStop", &CarrierCallStringRPC{user, call, reason}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) CallStop: ", err.Error)
	}
	return
}

func makeCallConnect(user User, call Call) {
	var result error
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callConnect)  Carrier ID not retrieved", call.ID, user.ID)
		return
	}
	this.fleet[carrierId].Go("CarrierRPC.CallConnect", &CarrierCallRPC{user, call}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) CallConnect: ", err.Error)
	}
	return
}

func makeCallStart(user User, call Call) {
	var result error
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callStart)  Carrier ID not retrieved", call.ID, user.ID)
		return
	}
	this.fleet[carrierId].Go("CarrierRPC.CallStart", &CarrierCallRPC{user, call}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) CallStart: ", err.Error)
	}
	return
}

func makeCallFinish(user User, call Call) {
	var result error
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callFinish)  Carrier ID not retrieved", call.ID, user.ID)
		return
	}
	this.fleet[carrierId].Go("CarrierRPC.CallFinish", &CarrierCallRPC{user, call}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) CallFinish: ", err.Error)
	}
	return
}

func makeCallAftermath(user User, call Call, event_type string, event_action string) {
	var result error
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) startCallAftermath)  Carrier ID not retrieved", call.ID, user.ID)
		return
	}
	this.fleet[carrierId].Go("CarrierRPC.CallAftermath", &CarrierCallAftermathRPC{user, call, event_action, event_type}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) StartCallAftermath: ", err.Error)
	}
	return
}

func makeCallAnswer(user User, call Call, decision bool) {
	var result error
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callAnswer)  Carrier ID not retrieved", call.ID, user.ID)
		return
	}
	this.fleet[carrierId].Go("CarrierRPC.CallAnswer", &CarrierCallBoolRPC{user, call, decision}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) CallAnswer: ", err.Error)
	}
	return
}

func makeCallReveal(user User, call Call, decision bool) {
	var result error
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (Call(%d), User(%d) callReveal)  Carrier ID not retrieved", call.ID, user.ID)
		return
	}
	this.fleet[carrierId].Go("CarrierRPC.CallReveal", &CarrierCallBoolRPC{user, call, decision}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) CallReveal: ", err.Error)
	}
	return
}

func makeSendMessage(user User, message Message) {
	var result error
	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(user.ID)).Val()
	if carrierId == "" {
		log.Printf("(RPC) (User(%d) Message)  Carrier ID not retrieved", user.ID)
		return
	}
	this.fleet[carrierId].Go("CarrierRPC.SendMessage", &CarrierMessageRPC{user, message}, &result, callChan)
	err := <-callChan
	if err.Error != nil {
		log.Println("(RPC) SendMessage: ", err.Error)
	}
	return
}
