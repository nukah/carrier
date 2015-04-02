package control

import (
	"errors"
	"fmt"
)

type ControlCallAcceptRPC struct {
	User     User
	CallId   int
	Decision bool
}

type ControlCallStopRPC struct {
	User   User
	CallId int
}

type ControlCallMessageRPC struct {
	Message Message
}

type ControlRPC int

func (control *ControlRPC) AcceptCall(args *ControlCallAcceptRPC, result *bool) error {
	call := this.calls[args.CallId]

	if call == nil {
		*result = false
	}

	if err := call.Accept(args.User.ID, args.Decision); err != nil {
		*result = false
	} else {
		*result = true
	}
	return nil
}

func (control *ControlRPC) StopCall(args *ControlCallStopRPC, result *bool) error {
	call := this.calls[args.CallId]

	if call == nil {
		*result = false
	}

	if err := call.Stop(args.User); err != nil {
		*result = false
	} else {
		*result = true
	}
	return nil
}

func (control *ControlRPC) CallMessage(args *ControlCallMessageRPC, result *error) error {
	call := this.calls[args.Message.CallId]

	if call == nil {
		*result = errors.New(fmt.Sprintf("Call does not exist %d", args.Message.CallId))
	}

	if call.Incognito {
		args.Message.Incognito = true
	}

	makeSendMessage(call.Destination, args.Message)
	return nil
}
