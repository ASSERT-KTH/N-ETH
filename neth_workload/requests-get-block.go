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

func do_get_block_request(address string, index int, req Request, time_pairs *[]TimePair, out chan IndexedResponse) {

	json_data, err := json.Marshal(req)

	if err != nil {
		panic(err)
	}

	client := http.Client{
		Timeout: 1 * time.Second,
	}

	start := time.Now().UnixMicro()
	resp, err := client.Post(address, "application/json", bytes.NewBuffer(json_data))
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

	out_response := IndexedResponse{Method: req.Method, Index: index, Response: fmt.Sprintf("%s,%d\n", json_obj.Res.Number, etherscan_block)}
	out <- out_response
}

var etherscan_block int64 = -1

//{"jsonrpc":"2.0","id":83,"result":"0xf4dc68"}
type EtherscanResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  string `json:"result"`
}

func updateEtherscanBlockNumber() {
	apikey := os.Getenv("ETHERSCAN_API_KEY")
	for {
		resp, err := http.Get(fmt.Sprintf("https://api.etherscan.io/api?module=proxy&action=eth_blockNumber&apikey=%s", apikey))

		if err != nil {
			fmt.Println("failed to update etherscan block number: req")
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("failed to update etherscan block number: res")
			return
		}

		etherscan_resp := new(EtherscanResponse)
		err = json.Unmarshal(body, etherscan_resp)

		if err != nil {
			fmt.Println("failed to update etherscan block number: json")
			return
		}

		trim := etherscan_resp.Result[2:]

		block_n, err := strconv.ParseInt(trim, 16, 64)
		if err != nil {
			fmt.Println("failed to update etherscan block number: parse")
			return
		}

		etherscan_block = block_n

		time.Sleep(12 * time.Second)

	}
}

func run_get_block(config Config, experiment_tag string) {

	n_requests := config.n_requests
	time_pairs := make([]TimePair, n_requests)

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
			go do_get_block_request(config.proxy_address, n, req, &time_pairs, out)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	output_dir := fmt.Sprintf("%s/requests-%s", os.Getenv("OUTPUT_DIR"), experiment_tag)
	err := os.Mkdir(output_dir, 0775)

	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	f, err := os.Create(fmt.Sprintf("%s/responses-get-block.dat", output_dir))
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

	g, err := os.Create(fmt.Sprintf("%s/latencies-get-block.dat", output_dir))
	defer g.Close()

	for index, lat := range time_pairs {
		fmt.Fprintf(g, "%d,%d,%d,%d\n", index, lat.start, lat.end, lat.getTime())
	}

	println("Done!")

}
