package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

type Request struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
}

func main() {

	// {"method":"eth_getBlockByNumber","params":["0xa55e27", false],"id":1,"jsonrpc":"2.0"}
	request := Request{
		Method:  "eth_getBlockByNumber",
		Params:  []interface{}{"0xa55e27", false},
		Id:      1,
		Jsonrpc: "2.0",
	}
	json_data, err := json.Marshal(request)

	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	sb := string(body)
	log.Printf(sb)
}
