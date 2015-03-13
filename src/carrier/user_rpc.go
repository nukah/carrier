package carrier

type UserCallRPCArguments struct {
	User User
	Call Call
}

type UserRPCArguments struct {
	User User
}

type UserRPC int

func (rpc *UserRPC) CallConnect(args *UserCallRPCArguments, err *error) error {
	return args.User.SendCallConnect(args.Call)
}

func (rpc *UserRPC) isUserOnline(args *UserRPCArguments, result *bool) error {
	if args.User.Online() {
		*result = true
	} else {
		*result = false
	}
	return nil
}

func (rpc *UserRPC) isUserInCall(args *UserRPCArguments, result *bool) error {
	if args.User.Online() {
		*result = true
	} else {
		*result = false
	}
	return nil
}

// func userOnline(user User) (bool, error) {
// 	var result bool
// 	online := make(chan bool)
// 	carrierId := _redis.HGet("formation:users", string(user.ID)).Val()
// 	if carrierId == "" {
// 		return false, errors.New("(RPC) Carrier ID not retrieved")
// 	}
// 	_carrier.Formation[carrierId].Go("isUserOnline", &UserRPCArguments{user}, result, nil)
// 	result = <-online
// 	return result, nil
// }

// func userInCall(user User) (bool, error) {
// 	var result bool
// 	inCall := make(chan bool)
// 	carrierId := _redis.HGet("formation:users", string(user.ID)).Val()
// 	if carrierId == "" {
// 		return false, errors.New("(RPC) Carrier ID not retrieved")
// 	}
// 	_carrier.Formation[carrierId].Go("isUserInCall", &UserRPCArguments{user}, result, nil)
// 	result = <-inCall
// 	return result, nil
// }

// func makeCallConnect(user User, call Call) error {
// 	carrierId := _redis.HGet("formation:users", string(user.ID)).Val()
// 	if carrierId == "" {
// 		return errors.New("(RPC) Carrier ID not retrieved")
// 	}
// 	_carrier.Formation[carrierId].Call("CallConnect", &UserCallRPCArguments{user, call}, nil)

// 	return nil
// }
