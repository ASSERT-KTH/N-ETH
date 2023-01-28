package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

type ClientInfo struct {
	name string
	port string
}

var clients = []ClientInfo{
	{
		name: "geth",
		port: "8545",
	},
	{
		name: "besu",
		port: "8546",
	},
	{
		name: "erigon",
		port: "8547",
	},
	{
		name: "nethermind",
		port: "8548",
	},
}

func RunClient(target ClientInfo, wg *sync.WaitGroup, stop chan int) {

	fmt.Printf("init sync for %s\n", target.name)

	nvme_dir := fmt.Sprintf("%s/docker-nvme-%s", os.Getenv("HOME"), target.name)
	output_dir := os.Getenv("OUTPUT_DIR")

	cmd := exec.Command(
		"docker",
		"run",
		"--privileged",
		"--rm",
		"--pid=host",
		"-v",
		fmt.Sprintf("%s:/root/nvme", nvme_dir),
		"-v",
		fmt.Sprintf("%s:/output", output_dir),
		"-e",
		"ETHERSCAN_API_KEY",
		"-p",
		fmt.Sprintf("%s:8545", target.port),
		fmt.Sprintf("javierron/neth:%s", target.name),
		"./synchronize-ready.sh", //command
		target.name,
	)

	fmt.Printf("Begin sync %s in path %s\n", target.name, nvme_dir)
	fmt.Println(cmd.String())

	outfile, err := os.Create(fmt.Sprintf("%s/docker-sync-%s.log", output_dir, target.name))
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile

	errfile, err := os.Create(fmt.Sprintf("%s/docker-sync-err-%s.log", output_dir, target.name))
	if err != nil {
		panic(err)
	}
	defer errfile.Close()

	cmd.Stderr = errfile
	cmd.Start()

	<-stop
	fmt.Printf("%s: interrupt sync script\n", target.name)
	cmd.Process.Signal(os.Interrupt)

	fmt.Printf("%s: check for stop\n", target.name)
	cmd.Wait()

	fmt.Printf("%s: done stopping\n", target.name)
	wg.Done()
}

func PollInitialSync() {
	all_clients_ready := false
	for !all_clients_ready {
		all_clients_ready = true
		for _, client := range clients {
			dat, err := os.ReadFile(
				fmt.Sprintf("%s/ipc-%s.dat", os.Getenv("OUTPUT_DIR"), client),
			)

			if err != nil {
				panic(err)
			}

			all_clients_ready = all_clients_ready && (string(dat) == "READY")
		}

		if !all_clients_ready {
			time.Sleep(10 * time.Second)
		}
	}
}

func SyncSourceClients() {

	stop_chan := make(chan int)
	wg := new(sync.WaitGroup)
	wg.Add(len(clients))
	for _, client := range clients {
		go RunClient(client, wg, stop_chan)
	}

	time.Sleep(10 * time.Second)

	PollInitialSync()

	for range clients {
		stop_chan <- 1
	}

	wg.Wait()

}

func CheckEnvs() {
	_, ok := os.LookupEnv("OUTPUT_DIR")

	if !ok {
		fmt.Println("Env variable OUTPUT_DIR not set!")
		os.Exit(-1)
	}
}

func main() {

	CheckEnvs()

	// sync targets
	SyncSourceClients()
	fmt.Println("All clients synchronized and stopped!")

	//foreach error model
	//	 copy targets
	//	 start sync again

	//	run copies
	//	  run docker images with pre-sync script (synchronize.sh)

	//	when pre-sync of all clients is done
	//	  terminate sync docker containers

	//	  run docker images directly with error injection script (TODO!)
	//	  run workload

}
