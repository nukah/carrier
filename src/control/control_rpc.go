package control

type ControlCallAcceptRPC struct {
	User     User
	CallId   string
	Decision bool
}

type ControlRPC int

func (control *ControlRPC) AcceptCall(args *ControlCallAcceptRPC, result *error) error {
	call := new(Call)

	if err := call.Find(args.CallId); err != nil {
		*result = err
	}
	if err := call.Accept(args.User.ID, args.Decision); err != nil {
		*result = err
	}
	return nil
}
