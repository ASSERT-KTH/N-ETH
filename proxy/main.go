package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type ContextEntry string

var latest_block uint64 = 0
var targets = []string{
	"http://localhost:8545", // geth
	"http://localhost:8546", // besu
	"http://localhost:8547", // erigon
	"http://localhost:8548", // nethermind
}

func Process(w http.ResponseWriter, r *http.Request) {

	fmt.Println("received request!")

	response := new(ComparableResponse)

	if r.Method != http.MethodPost {
		//set error status
		return
	}

	//TODO handle timeouts!
	client := http.Client{
		Timeout: 1 * time.Second,
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return
	}

	strategy := new(AdaptiveStrategy)
	strategy.Init()

	success := false
	tries := 0

	for !success && tries < len(targets) {

		//select target!
		// targetIdx := r.Context().Value("request-id").(int32)
		target := strategy.GetNext()

		fmt.Printf("sending req to target %d!\n", target)

		resp, err := client.Post(
			targets[target],
			"application/json",
			bytes.NewBuffer(body),
		)

		fmt.Println("received res!")

		if err != nil {
			// possible request failure on request
			tries += 1
			fmt.Println(err.Error())
			strategy.Failure(target)

			response.Update(Unavailable, resp.StatusCode, nil)
			continue
		}

		resp_body, err := ioutil.ReadAll(resp.Body)

		// handle possible response failures
		if err != nil {
			// possible response failure on reading request
			tries += 1
			fmt.Println(err.Error())
			strategy.Failure(target)
			response.Update(Degraded_http, resp.StatusCode, resp_body)
			continue
		}

		json_obj := new(RPCResponse)
		err = json.Unmarshal(resp_body, json_obj)

		if err != nil {
			// possible response failure on parsing json
			tries += 1
			fmt.Println(err.Error())
			strategy.Failure(target)
			response.Update(Degraded_json, resp.StatusCode, resp_body)
			continue
		}

		block_number, err := HexStringToInt(json_obj.Result.Number)

		if err != nil {
			block_number = 0
		}

		if latest_block-block_number > 5 {
			fmt.Println("outdated response!")
			strategy.Failure(target)
			response.Update(Degraded_freshness, resp.StatusCode, resp_body)
			continue
		} else if latest_block < block_number {
			latest_block = block_number
		}

		success = true
		// relay response
		// fmt.Println(string(resp_body))
		strategy.Success(target)
		response.Update(Available, resp.StatusCode, resp_body)
	}

	w.WriteHeader(response.statusCode)
	w.Write(response.body)
}

func handleRequests() {
	mux := http.NewServeMux()

	handler := http.HandlerFunc(Process)

	// mux.Handle("/", addRequestID(handler))
	mux.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func main() {
	// 1. receive request
	handleRequests()

	fmt.Println("hello world")
}
