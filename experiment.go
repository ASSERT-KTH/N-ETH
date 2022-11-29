package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var error_models = []string{
	"error_models_processed_1_1.005.json",
	// "error_models_processed_1_1.015.json",
	// "error_models_processed_1_1.01.json",
	// "error_models_processed_1_1.025.json",
	// "error_models_processed_1_1.05.json",
	// "error_models_processed_1_1.1.json",
	"error_models_processed_2_1.005.json",
	// "error_models_processed_2_1.015.json",
	// "error_models_processed_2_1.01.json",
	// "error_models_processed_2_1.025.json",
	// "error_models_processed_2_1.05.json",
	// "error_models_processed_2_1.1.json",
	"error_models_processed_3_1.005.json",
	// "error_models_processed_3_1.015.json",
	// "error_models_processed_3_1.01.json",
	// "error_models_processed_3_1.025.json",
	// "error_models_processed_3_1.05.json",
	// "error_models_processed_3_1.1.json",
	"error_models_processed_4_1.005.json",
	// "error_models_processed_4_1.015.json",
	// "error_models_processed_4_1.01.json",
	// "error_models_processed_4_1.025.json",
	// "error_models_processed_4_1.05.json",
	// "error_models_processed_4_1.1.json",
	"error_models_processed_5_1.005.json",
	// "error_models_processed_5_1.015.json",
	// "error_models_processed_5_1.01.json",
	// "error_models_processed_5_1.025.json",
	// "error_models_processed_5_1.05.json",
	// "error_models_processed_5_1.1.json",
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

	mkdir := exec.Command(
		"mkdir",
		"-p",
		nvme_dir,
	)

	error_models_prefix := "https://raw.githubusercontent.com/javierron/royal-chaos/error-model-extraction/chaoseth/experiments/common-error-models"

	output_dir := fmt.Sprintf("./output-%d", index)

	mkdir = exec.Command(
		"mkdir",
		"-p",
		output_dir,
	)

	mkdir.Start()
	mkdir.Wait()

	cmd := exec.Command(
		"docker",
		"run",
		"--privilieged",
		"--rm",
		"--pid=host",
		fmt.Sprintf("-v %s:/root/nvme", nvme_dir),
		fmt.Sprintf("-v %s:/output", output_dir),
		"./single-version-controller.sh", //command
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

	source_partition := "/dev/nvme0n1p1"
	source_partition_mount_pouint := fmt.Sprintf("%s/nvme", os.Getenv("HOME"))

	target_partition := fmt.Sprintf("/dev/nvme%dn1p1", index)
	target_partition_mount_pouint := fmt.Sprintf("%s/docker-nvme-%d", os.Getenv("HOME"), index)

	umount_source := exec.Command(
		"sudo",
		"umount",
		source_partition_mount_pouint,
	)

	umount_source.Run()

	umount_target := exec.Command(
		"sudo",
		"umount",
		target_partition_mount_pouint,
	)

	umount_target.Run()

	umount_source.Wait()
	umount_target.Wait()

	cmd := exec.Command(
		"sudo",
		"dd",
		fmt.Sprintf("if=%s", source_partition),
		fmt.Sprintf("of=%s", target_partition),
		"bs=3000M",
		"status=progress",
	)

	outfile, err := os.Create(fmt.Sprintf("%s/dd_progress-%s.log", os.Getenv("HOME"), index))
	if err != nil {
		panic(err)
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	cmd.Run()
	cmd.Wait()

	mount_source := exec.Command(
		"sudo",
		"mount",
		source_partition,
		source_partition_mount_pouint,
	)

	mount_source.Run()

	mount_target := exec.Command(
		"sudo",
		"mount",
		target_partition,
		target_partition_mount_pouint,
	)

	mount_target.Run()

	mount_source.Wait()
	mount_target.Wait()
}

type State string

const (
	Syncing State = "SYNC"
	Copying State = "COPY"
)

func source_loop(start chan int, copy chan CopyInfo, err_chan chan error) {
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

func error_handler(err_chan chan error) {
	err := <-err_chan
	panic(err)
}

func print_usage() {
	fmt.Printf("Usage: go run experiment.go <target> <n_disks> \n")
}

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("Argument error\n")
		print_usage()
		os.Exit(-1)
	}

	target := os.Args[1]
	experiments := len(error_models)

	parsed, err := strconv.ParseInt(os.Args[2], 10, 32)
	if err != nil {
		fmt.Printf("cannot parse number of disks argument: %s", os.Args[2])
		return
	}
	disks := int(parsed)

	start_chan := make(chan int)
	copy_chan := make(chan CopyInfo)
	err_chan := make(chan error)
	go source_loop(start_chan, copy_chan, err_chan)
	go error_handler(err_chan)
	<-start_chan

	mstack := new(MutexStack)
	mstack.Init(disks)

	go func() {
		for exp := 1; exp <= experiments; exp++ {
			go new_run(mstack, exp, target, copy_chan)
			time.Sleep(time.Duration(rand.Intn(3)) * time.Second)
		}
	}()

	for {
		mstack.Print()
		time.Sleep(2 * time.Second)
	}

}
