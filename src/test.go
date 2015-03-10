package test

import (
	"encoding/base64"
	"flag"
	"fmt"
	"gopkg.in/vmihailenco/msgpack.v2"
	"net/url"
	"reflect"
	"strings"
)

var (
	input, trimmed string
	data           []byte
	err            error
	output         map[string]interface{}
)

func main() {
	flag.StringVar(&input, "i", "", "String header")
	flag.Parse()

	trimmed, _ = url.QueryUnescape(input)

	data, err = base64.StdEncoding.DecodeString(strings.Join(strings.Split(trimmed, "\n"), ""))
	if err != nil {
		fmt.Println("Error on decoding", err)
		return
	}

	err = msgpack.Unmarshal(data, &output)
	if err != nil {
		fmt.Println("Error on unpacking", err)
		return
	}

	fmt.Println(reflect.TypeOf(output["warden.user.api.key"]))
}
