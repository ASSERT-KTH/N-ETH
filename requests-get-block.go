package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TimePair struct {
	start int64
	end   int64
}

func (tp *TimePair) getTime() int64 {
	return tp.end - tp.start
}

type Requests []Request

type Request struct {
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
	Jsonrpc string        `json:"jsonrpc"`
}

type Response struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Res     Result `json:"result"`
}

type Result struct {
	Number string `json:"number"`
}

type IndexedResponse struct {
	Method   string
	Response string
	Index    int
}

func do_request(index int, req Request, time_pairs *[]TimePair, out chan IndexedResponse) {

	json_data, err := json.Marshal(req)

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	start := time.Now().UnixMicro()
	resp, err := client.Post("http://localhost:8545", "application/json", bytes.NewBuffer(json_data))
	end := time.Now().UnixMicro()

	measured_time := TimePair{
		start: start,
		end:   end,
	}

	(*time_pairs)[index] = measured_time

	// error on request
	if err != nil {
		err_str := fmt.Sprintf("error: %s\n", err.Error())
		fmt.Print(err_str)
		error_response := IndexedResponse{Method: req.Method, Index: index, Response: err_str}
		out <- error_response
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	// error on reading response
	if err != nil {
		err_str := fmt.Sprintf("error: %s\n", err.Error())
		fmt.Print(err_str)
		error_response := IndexedResponse{Method: req.Method, Index: index, Response: err_str}
		out <- error_response
		return
	}

	json_obj := new(Response)
	err = json.Unmarshal(body, json_obj)
	// error on parsing json
	if err != nil {
		err_str := fmt.Sprintf("error: %s\n", err.Error())
		fmt.Print(err_str)
		error_response := IndexedResponse{Method: req.Method, Index: index, Response: err_str}
		out <- error_response
		return
	}

	out_response := IndexedResponse{Method: req.Method, Index: index, Response: fmt.Sprintf("%s,%s\n", json_obj.Res.Number, etherscan_block)}
	out <- out_response
}

func load_requests() (Requests, error) {
	jsonFile, err := os.Open("./requests.json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	reqs := Requests{}
	json.Unmarshal(bytes, &reqs)

	return reqs, nil
}

var etherscan_block int64 = -1

//{"jsonrpc":"2.0","id":83,"result":"0xf4dc68"}
type EtherscanResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  string `json:"result"`
}

func updateEtherscanBlockNumber() {

	for {
		resp, err := http.Get("https://api.etherscan.io/api?module=proxy&action=eth_blockNumber")

		if err != nil {
			fmt.Println("failed to update etherscan block number")
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("failed to update etherscan block number")
			return
		}

		etherscan_resp := new(EtherscanResponse)
		json.Unmarshal(body, etherscan_resp)

		trim := etherscan_resp.Result[2:]

		block_n, err := strconv.ParseInt(trim, 16, 64)
		if err != nil {
			fmt.Println("failed to update etherscan block number")
			return
		}

		etherscan_block = block_n

		time.Sleep(10 * time.Second)

	}
}

func main() {

	n_requests := 360_000
	time_pairs := make([]TimePair, n_requests)

	// "jsonrpc": "2.0",
	// "method": "eth_getBlockByNumber",
	// "id": 1,
	// "params": [
	// 	"0xa55e27",
	// 	false
	// ]

	go updateEtherscanBlockNumber()

	req := Request{
		Jsonrpc: "2.0",
		Method:  "eth_getBlockByNumber",
		Id:      1,
		Params:  []interface{}{"latest", false},
	}

	out := make(chan IndexedResponse)
	go func() {
		for n := 0; n < n_requests; n++ {
			go do_request(n, req, &time_pairs, out)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	f, err := os.Create("/output/responses-get-block.dat")
	defer f.Close()

	if err != nil {
		println("Unable to create results file")
		os.Exit(-1)
	}

	n := 0
	for response := range out {
		fmt.Fprintf(f, "%d, %s, %s\n", response.Index, response.Method, response.Response)
		n++

		if n == n_requests {
			close(out)
		}
	}

	g, err := os.Create("/output/latencies-get-block.dat")
	defer g.Close()

	for index, lat := range time_pairs {
		fmt.Fprintf(g, "%d,%d,%d,%d\n", index, lat.start, lat.end, lat.getTime())
	}

	println("Done!")

}
