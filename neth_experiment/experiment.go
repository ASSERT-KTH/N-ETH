package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

var clients = []string{
	"geth",
	"besu",
	"erigon",
	"nethermind",
}

func RunClient(target string, wg *sync.WaitGroup, stop chan int) {

	fmt.Printf("init sync for %s\n", target)

	nvme_dir := fmt.Sprintf("%s/docker-nvme-%s", os.Getenv("HOME"), target)
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
		fmt.Sprintf("javierron/neth:%s", target),
		"./synchronize-ready.sh", //command
		target,
	)

	fmt.Printf("Begin sync %s in path %s\n", target, nvme_dir)
	fmt.Println(cmd.String())

	outfile, err := os.Create(fmt.Sprintf("%s/docker-sync-%s.log", output_dir, target))
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile

	errfile, err := os.Create(fmt.Sprintf("%s/docker-sync-err-%s.log", output_dir, target))
	if err != nil {
		panic(err)
	}
	defer errfile.Close()

	cmd.Stderr = errfile
	cmd.Start()

	<-stop
	fmt.Printf("%s: interrupt sync script\n", target)
	cmd.Process.Signal(os.Interrupt)

	fmt.Printf("%s: check for stop\n", target)
	cmd.Wait()

	fmt.Printf("%s: done stopping\n", target)
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

	SyncSourceClients()
	fmt.Println("All clients synchronized and stopped!")
	// sync targets

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
