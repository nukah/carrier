package carrier

import (
	"errors"
	"log"
	"net/rpc"
)

type ControlCallStopRPC struct {
	User   User
	CallId int
}

type ControlCallAcceptRPC struct {
	User     User
	CallId   int
	Decision bool
}

type ControlCallMessageRPC struct {
	Message Message
}

func controlCallAccept(user User, call_id int, decision bool) error {
	var result bool
	var callChan = make(chan *rpc.Call, 1)

	control.Go("ControlRPC.AcceptCall", &ControlCallAcceptRPC{user, call_id, decision}, &result, callChan)
	err := <-callChan
	if result == false {
		return errors.New("Call Accept Failed")
	}
	if err.Error != nil {
		return err.Error
	}
	return nil
}

func controlCallCancel(user User, call_id int) error {
	var result bool
	var callChan = make(chan *rpc.Call, 1)

	control.Go("ControlRPC.StopCall", &ControlCallStopRPC{user, call_id}, &result, callChan)

	err := <-callChan
	if result == false {
		return errors.New("Call Stop Failed")
	} else {
		if err.Error != nil {
			return err.Error
		}
	}
	return nil
}

func controlCallMessage(message Message) error {
	var result bool
	var callChan = make(chan *rpc.Call, 1)

	control.Go("ControlRPC.CallMessage", &ControlCallMessageRPC{message}, &result, callChan)

	err := <-callChan
	if err.Error != nil {
		return err.Error
	}

	message.Action = "sent"
	message.Source.SendMessage(message)

	return nil
}

func carrierContactMessage(message Message) error {
	var result error
	var destination User

	this.db.Find(&destination, message.DestinationId)

	var callChan = make(chan *rpc.Call, 1)
	carrierId := this.redis.HGet("formation:users", string(destination.ID)).Val()
	if carrierId != "" && carrierId != this.id {
		this.carrierFleet[carrierId].Go("CarrierRPC.SendMessage", &CarrierMessageRPC{destination, message}, &result, callChan)
		err := <-callChan
		if err.Error != nil {
			log.Println(err.Error)
			return err.Error
		}
	} else {
		destination.SendMessage(message)
	}

	message.Action = "sent"
	message.Source.SendMessage(message)

	return nil
}
