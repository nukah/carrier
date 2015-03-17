package carrier

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

type CarrierRPC int

func (rpc *CarrierRPC) CallConnect(args *CarrierCallRPC, err *error) error {
	return args.User.SendCallConnect(args.Call)
}

func (rpc *CarrierRPC) isUserOnline(args *CarrierUserRPC, result *bool) error {
	if args.User.Online() {
		*result = true
	} else {
		*result = false
	}
	return nil
}

func (rpc *CarrierRPC) isUserInCall(args *CarrierUserRPC, result *bool) error {
	if args.User.Online() {
		*result = true
	} else {
		*result = false
	}
	return nil
}

func (rpc *CarrierRPC) CallStop(args *CarrierCallStringRPC, result *error) error {
	*result = args.User.SendCallStop(args.Call, args.Value)
	res := *result
	return res
}

func (rpc *CarrierRPC) CallFinish(args *CarrierCallRPC, result *error) error {
	*result = args.User.SendCallFinish(args.Call)
	res := *result
	return res
}

func (rpc *CarrierRPC) CallAnswer(args *CarrierCallBoolRPC, result *error) error {
	*result = args.User.SendCallAnswer(args.Call, args.Value)
	res := *result
	return res
}

func (rpc *CarrierRPC) CallReveal(args *CarrierCallBoolRPC, result *error) error {
	*result = args.User.SendCallReveal(args.Call, args.Value)
	res := *result
	return res
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