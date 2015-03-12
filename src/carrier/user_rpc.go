package carrier

import (
  "errors"
)

type UserRPCArguments struct {
	User User
	Call Call
}
type UserRPC int

func (rpc *UserRPC) CallConnect(args *UserRPCArguments, err *error) error {
	return args.User.SendCallConnect(args.Call)
}

func FleetCallConnect(user *User, call *Call) error {
  carrierId := Redis.HGet("formation:users", u.ID, Carrier.ID).Val()
  if carrierId == "" {
    return errors.New("(RPC) Carrier ID not retrieved")
  }
  rpcCall := Carrier.Formation[carrierId].Go('CallConnect', &UserRPCArguments{user, call}, nil, nil)
  err := <- rpcCall.Done()
  if err != nil {
    return err
  }
  return nil
}