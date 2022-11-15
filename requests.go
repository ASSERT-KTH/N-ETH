package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {

	// {"method":"eth_getBlockByNumber","params":["0xa55e27", false],"id":1,"jsonrpc":"2.0"}
	values := map[string]string{
		"method":  "eth_getBlockByNumber",
		"params":  "[\"0xa55e27\", false]",
		"id":      "1",
		"jsonrpc": "2.0",
	}
	json_data, err := json.Marshal(values)

	resp, err := http.Post("http://localhost:4545", "application/json", bytes.NewBuffer(json_data))
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
