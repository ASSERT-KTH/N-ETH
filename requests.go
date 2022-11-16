package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Request struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
}

func doRequest(req Request, out chan string) {

	json_data, err := json.Marshal(req)

	resp, err := http.Post("http://localhost:8545", "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)
	out <- sb
}

func main() {
	request := Request{
		Method:  "eth_getBlockByNumber",
		Params:  []interface{}{"0xa55e27", false},
		Id:      1,
		Jsonrpc: "2.0",
	}

	out := make(chan string)
	go func() {
		for i := 0; i < 50; i++ {
			go doRequest(request, out)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	n := 0
	for response := range out {
		println(response)
		n++

		if n == 50 {
			close(out)
		}
	}

	println("Done!")

}
