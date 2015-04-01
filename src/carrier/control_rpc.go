package carrier

import (
	"errors"
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
