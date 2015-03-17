package carrier

import "log"

type ControlCallAcceptRPC struct {
	User     User
	CallId   string
	Decision bool
}

func controlCallAccept(user User, call_id string, decision bool) {
	err := control.Call("AcceptCall", &ControlCallAcceptRPC{user, call_id, decision}, nil)
	if err != nil {
		log.Printf("(CallAccept) Error with user %d in call %s", user.ID, call_id)
	}
}
