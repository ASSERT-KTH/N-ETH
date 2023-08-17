package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
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
