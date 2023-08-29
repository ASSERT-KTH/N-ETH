package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

type ContextEntry string

var latest_block int64 = 0

//TODO: move this to config file
var targets = []string{
	"http://172.17.0.1:8645", // nethermind
	"http://172.17.0.1:8646", // geth
	"http://172.17.0.1:8647", // erigon
	"http://172.17.0.1:8648", // besu
}

var id = 0

func Process(w http.ResponseWriter, r *http.Request) {
	id++
	// fmt.Println("received request!")

	response := new(ComparableResponse)

	if r.Method != http.MethodPost {
		//set error status
		return
	}

	//TODO handle timeouts!
	client := http.Client{
		Timeout: 250 * time.Millisecond,
	}

	body, err := ioutil.ReadAll(r.Body)

	if err != nil {
		return
	}

	strategy := GetNewRetryStrategy()

	success := false
	tries := 0

	for !success && tries < len(targets) {

		//select target!
		// targetIdx := r.Context().Value("request-id").(int32)
		target := strategy.GetNext()

		// fmt.Printf("sending req to target %d!\n", target)

		resp, err := client.Post(
			targets[target],
			"application/json",
			bytes.NewBuffer(body),
		)

		// fmt.Println("received res!")

		if err != nil {
			// possible request failure on request
			fmt.Println(err.Error())
			strategy.Failure(target)

			response.Update(Unavailable, 500, nil) // 500 as internal server error
			tries += 1
			continue
		}

		resp_body, err := ioutil.ReadAll(resp.Body)

		// handle possible response failures
		if err != nil {
			// possible response failure on reading request
			fmt.Println(err.Error())
			strategy.Failure(target)
			response.Update(Degraded_http, resp.StatusCode, resp_body)
			tries += 1
			continue
		}

		json_obj := new(RPCResponse)
		err = json.Unmarshal(resp_body, json_obj)

		if err != nil {
			// possible response failure on parsing json
			fmt.Println(err.Error())
			strategy.Failure(target)
			response.Update(Degraded_json, resp.StatusCode, resp_body)
			tries += 1
			continue
		}

		block_number, err := HexStringToInt(json_obj.Result.Number)

		if err != nil {
			fmt.Println("error parsing block number!")
			fmt.Println(err.Error())
			strategy.Failure(target)
			response.Update(Degraded_json, resp.StatusCode, resp_body)
			tries += 1
			continue
		}

		distance := atomic.LoadInt64(&latest_block) - block_number

		if distance > 5 {
			fmt.Printf("outdated response! block_number: %d latest: %d\n", block_number, latest_block)
			strategy.Failure(target)
			response.Update(Degraded_freshness, resp.StatusCode, resp_body)
			tries += 1
			continue
		}

		if distance < 0 {
			atomic.SwapInt64(&latest_block, block_number)
		}

		success = true
		tries += 1
		// relay response
		// fmt.Println(string(resp_body))
		strategy.Success(target)
		response.Update(Available, resp.StatusCode, resp_body)
	}

	w.WriteHeader(response.statusCode)
	w.Write(response.body)

	fmt.Println(id)
	strategy.LogStatus()
}

func handleRequests() {
	mux := http.NewServeMux()

	handler := http.HandlerFunc(Process)

	// mux.Handle("/", addRequestID(handler))
	mux.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func UpdateLatestBlock() {
	for {
		time.Sleep(12 * time.Second)
		atomic.AddInt64(&latest_block, 1)
	}
}

func main() {

	//read target from args
	if len(os.Args) < 3 {
		fmt.Println("usage: ./proxy <strategy> <N>")
		os.Exit(-1)
	}
	n, err := strconv.Atoi(os.Args[2])

	if err != nil || n < 1 || n > 4 {
		fmt.Println("invalid N")
		os.Exit(-1)
	}

	go UpdateLatestBlock()
	SelectStrategy(os.Args[1])
	handleRequests()
}
