package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func do_random_request(address string, index int, req Request, time_pairs *[]TimePair, out chan IndexedResponse) {

	json_data, err := json.Marshal(req)

	if err != nil {
		panic(err)
	}

	client := http.Client{
		Timeout: 10 * time.Second,
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

	json_obj := new(map[string]interface{})
	err = json.Unmarshal(body, json_obj)
	// error on parsing json
	if err != nil {
		err_str := fmt.Sprintf("error: %s\n", err.Error())
		fmt.Print(err_str)
		error_response := IndexedResponse{Method: req.Method, Index: index, Response: err_str}
		out <- error_response
		return
	}

	out_response := IndexedResponse{Method: req.Method, Index: index, Response: "success"}
	out <- out_response
}

func run_random(config Config, experiment_tag string) {

	requests, err := load_requests()
	time_pairs := make([]TimePair, config.n_requests)
	n_requests := config.n_requests

	if err != nil {
		fmt.Println("Unable to load file with JSON requests")
		os.Exit(-1)
	}

	out := make(chan IndexedResponse)
	go func() {
		for n := 0; n < n_requests; n++ {
			req := requests[rand.Intn(len(requests))]
			go do_random_request(config.proxy_address, n, req, &time_pairs, out)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	f, err := os.Create("./responses_random.dat")
	defer f.Close()

	if err != nil {
		fmt.Println("Unable to create results file")
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

	g, err := os.Create("./latencies-random.dat")
	defer g.Close()

	for index, lat := range time_pairs {
		fmt.Fprintf(g, "%d,%d,%d,%d\n", index, lat.start, lat.end, lat.getTime())
	}

	//run again w/ getBlock only workload
	// h, err := os.Create("/output/recency.dat")
	// defer h.Close()

	fmt.Println("Done!")

}
