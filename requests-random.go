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

type IndexedResponse struct {
	Method   string
	Response string
	Index    int
}

func do_request(index int, req Request, time_pairs *[]TimePair, out chan IndexedResponse) {

	json_data, err := json.Marshal(req)

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	start := time.Now().UnixMicro()
	resp, err := client.Post("http://localhost:8080", "application/json", bytes.NewBuffer(json_data))
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

func main() {

	// n_requests := 360_000
	n_requests := 10
	requests, err := load_requests()
	time_pairs := make([]TimePair, n_requests)

	if err != nil {
		fmt.Println("Unable to load file with JSON requests")
		os.Exit(-1)
	}

	out := make(chan IndexedResponse)
	go func() {
		for n := 0; n < n_requests; n++ {
			req := requests[rand.Intn(len(requests))]
			go do_request(n, req, &time_pairs, out)
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
