package main

import (
	"fmt"
	"os"
)

type Config struct {
	n_requests    int
	proxy_address string
}

func main() {
	// validate first argument exists
	if len(os.Args) < 3 {
		fmt.Println("usage: ./workload <random|get_block> <experiment_tag>")
		os.Exit(-1)
	}

	// validate workload name
	workload_name := os.Args[1]
	if workload_name != "random" && workload_name != "get_block" {
		fmt.Println("Second argument must be either random or get_block")
		os.Exit(-1)
	}

	//validate experiment tag
	experiment_tag := os.Args[2]
	if len(experiment_tag) > 8 {
		fmt.Println("Experiment tag must be less than 8 characters")
		os.Exit(-1)
	}

	config := Config{
		proxy_address: "http://172.17.0.1:8080",
		n_requests:    100_000,
	}

	switch workload_name {
	case "random":
		run_random(config, experiment_tag)
	case "get_block":
		run_get_block(config, experiment_tag)
	default:
		panic("unreachable")
	}
}
