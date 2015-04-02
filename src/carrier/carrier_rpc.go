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

type CarrierCallAftermathRPC struct {
	User   User
	Call   Call
	Action string
	Type   string
}

type CarrierMessageRPC struct {
	Destination User
	Message     Message
}

type CarrierUserRPC struct {
	User User
}

type CarrierRPC int

func (rpc *CarrierRPC) CallConnect(args *CarrierCallRPC, err *error) error {
	*err = args.User.SendCallConnect(args.Call)
	return nil
}

func (rpc *CarrierRPC) IsUserOnline(args *CarrierUserRPC, result *bool) error {
	if args.User.Online() {
		*result = true
	} else {
		*result = false
	}
	return nil
}

func (rpc *CarrierRPC) IsUserInCall(args *CarrierUserRPC, result *bool) error {
	if args.User.InCall() {
		*result = true
	} else {
		*result = false
	}
	return nil
}

func (rpc *CarrierRPC) CallStop(args *CarrierCallStringRPC, err *error) error {
	*err = args.User.SendCallStop(args.Call, args.Value)
	return nil
}

func (rpc *CarrierRPC) CallStart(args *CarrierCallRPC, err *error) error {
	*err = args.User.SendCallStart(args.Call)
	return nil
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

func (rpc *CarrierRPC) CallAftermath(args *CarrierCallAftermathRPC, result *error) error {
	*result = args.User.SendCallAftermath(args.Call, args.Type, args.Action)
	res := *result
	return res
}

func (rpc *CarrierRPC) SendMessage(args *CarrierMessageRPC, result *error) error {
	*result = args.Destination.SendMessage(args.Message)
	res := *result
	return res
}
