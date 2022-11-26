package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"
)

const disks = 9
const experiments = 30

var error_models = [experiments]string{
	"error_models_processed_1_1.005.json",
	"error_models_processed_1_1.015.json",
	"error_models_processed_1_1.01.json",
	"error_models_processed_1_1.025.json",
	"error_models_processed_1_1.05.json",
	"error_models_processed_1_1.1.json",
	"error_models_processed_2_1.005.json",
	"error_models_processed_2_1.015.json",
	"error_models_processed_2_1.01.json",
	"error_models_processed_2_1.025.json",
	"error_models_processed_2_1.05.json",
	"error_models_processed_2_1.1.json",
	"error_models_processed_3_1.005.json",
	"error_models_processed_3_1.015.json",
	"error_models_processed_3_1.01.json",
	"error_models_processed_3_1.025.json",
	"error_models_processed_3_1.05.json",
	"error_models_processed_3_1.1.json",
	"error_models_processed_4_1.005.json",
	"error_models_processed_4_1.015.json",
	"error_models_processed_4_1.01.json",
	"error_models_processed_4_1.025.json",
	"error_models_processed_4_1.05.json",
	"error_models_processed_4_1.1.json",
	"error_models_processed_5_1.005.json",
	"error_models_processed_5_1.015.json",
	"error_models_processed_5_1.01.json",
	"error_models_processed_5_1.025.json",
	"error_models_processed_5_1.05.json",
	"error_models_processed_5_1.1.json",
}

type MutexStack struct {
	container   []int
	stack_index int
	semaphore   chan int
}

type CopyInfo struct {
	ret_chan chan int
	index    int
}

func (mstack *MutexStack) push(n int) {
	mstack.stack_index++
	mstack.container[mstack.stack_index] = n
}

func (mstack *MutexStack) pop() int {
	ret := mstack.container[mstack.stack_index]
	mstack.stack_index--
	return ret
}

func (mstack *MutexStack) Init(n int) {
	mstack.container = make([]int, n)
	mstack.semaphore = make(chan int, n)
	mstack.stack_index = -1

	for x := 0; x < n; x++ {
		mstack.push(x)
	}
}

func (mstack *MutexStack) Print() {
	fmt.Print("[ ")

	for x := 0; x <= mstack.stack_index; x++ {
		fmt.Print(mstack.container[x])
		fmt.Print(", ")
	}
	fmt.Printf(" ] index: %d \n", mstack.stack_index)
}

func (mstack *MutexStack) Done(n int) {
	<-mstack.semaphore
	mstack.push(n)
}

func (mstack *MutexStack) Request() int {
	mstack.semaphore <- 1
	return mstack.pop()
}

func new_run(mstack *MutexStack, exp_number int, target string, copy chan CopyInfo) {
	index := mstack.Request()
	fmt.Printf("Start task %d with index %d\n", exp_number, index)

	request_copy(index, copy)

	time.Sleep(time.Duration(45+rand.Intn(30)) * time.Second)
	mstack.Done(index)

	nvme_dir := fmt.Sprintf("%s/docker-nvme-%d", os.Getenv("HOME"), index)
	error_models_prefix := "https://raw.githubusercontent.com/javierron/royal-chaos/error-model-extraction/chaoseth/experiments/common-error-models"

	cmd := exec.Command(
		"docker",
		"run",
		"--privilieged",
		"--rm",
		"--pid=host",
		fmt.Sprintf("-v %s:/root/nvme", nvme_dir),
		"./single-version-fault-injection.sh", //command
		target,
		fmt.Sprintf("%s/%s", error_models_prefix, error_models[exp_number]),
	)
	cmd.Wait()
	fmt.Printf("Exit experiment %d with index %d\n", exp_number, index)
}

func request_copy(index int, copy chan CopyInfo) {
	w := make(chan int)

	copy <- CopyInfo{
		ret_chan: w,
		index:    index,
	}
	fmt.Printf("index %d sent copy request\n", index)

	<-w // done!
}

func first_sync() error {
	cmd := exec.Command("./synchronize-stop.sh", "geth")
	fmt.Println("init first sync")
	err := cmd.Run()
	if err != nil {
		println("failed to start first sync")
		return err
	}
	cmd.Wait()
	fmt.Println("done first sync")
	return nil
}

func sync(stop chan int) {
	fmt.Println("init sync sript")
	cmd := exec.Command("./synchronize.sh", "geth")
	<-stop
	fmt.Println("kill sync script")
	cmd.Process.Signal(os.Interrupt) //TODO: check if kills sub-processes
	fmt.Println("check for stop")
	//check for target + teku to stop
	fmt.Println("done stopping")
	stop <- 1
}

func copy_state(index int) {
	target_nvme := fmt.Sprintf("if=/dev/nvme0n1p%d", index)
	cmd := exec.Command(
		"sudo",
		"dd",
		"if=/dev/nvme0n1p1",
		fmt.Sprintf("of=%s", target_nvme),
		"bs=3000M",
		"status=progress",
	)
	cmd.Run()
	cmd.Wait()
}

type State string

const (
	Syncing State = "SYNC"
	Copying State = "COPY"
)

func sourceLoop(start chan int, copy chan CopyInfo, err_chan chan error) {
	err := first_sync()
	if err != nil {
		err_chan <- err
		return
	}
	start <- 1

	no_activity_counter := 0
	state := Copying
	stop := make(chan int)
	// go sync(stop)

	for {
		fmt.Printf("STATE: %s!\n", state)
		switch state {
		case Syncing:
			select {
			case data := <-copy:
				fmt.Println("received copy request")
				state = Copying
				no_activity_counter = 0
				stop <- 1
				<-stop

				fmt.Printf("copy to index %d!\n", data.index)
				copy_state(data.index)
				fmt.Println("done copying")
				data.ret_chan <- 1
			}
		case Copying:
			select {
			case data := <-copy:
				fmt.Println("received copy request")
				no_activity_counter = 0
				fmt.Printf("copy to index %d!\n", data.index)
				copy_state(data.index)
				fmt.Println("done copying")
				data.ret_chan <- 1
			default:
				fmt.Println("no activity!")
				no_activity_counter++
				time.Sleep(1 * time.Second)
				if no_activity_counter > 10 {
					fmt.Println("no activity, restarting sync!")
					state = Syncing
					go sync(stop)
				}
				fmt.Println("exit default")
			}
		}
	}
}

func errorHandler(err_chan chan error) {
	err := <-err_chan
	panic(err)
}

func main() {

	target := os.Args[1]
	if len(os.Args) < 2 {
		fmt.Printf("No target specified")
		os.Exit(-1)
	}
	if target != "geth" {
		fmt.Printf("Target %s not supported", target)
		os.Exit(-1)
	}

	start_chan := make(chan int)
	copy_chan := make(chan CopyInfo)
	err_chan := make(chan error)
	go sourceLoop(start_chan, copy_chan, err_chan)
	<-start_chan

	mstack := new(MutexStack)
	mstack.Init(disks)

	go func() {
		for exp := 0; exp < experiments; exp++ {
			go new_run(mstack, exp, target, copy_chan)
			time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
		}
	}()

	for {
		mstack.Print()
		time.Sleep(2 * time.Second)
	}

}
