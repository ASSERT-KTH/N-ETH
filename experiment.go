package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var error_models = []string{
	"error_models_1_1.05.json",
	"error_models_2_1.05.json",
	"error_models_3_1.05.json",
	"error_models_4_1.05.json",
	"error_models_5_1.05.json",
	"error_models_6_1.05.json",
	"error_models_7_1.05.json",
	"error_models_8_1.05.json",
	"error_models_9_1.05.json",
	"error_models_10_1.05.json",
	"error_models_11_1.05.json",
	"error_models_12_1.05.json",
	"error_models_13_1.05.json",
	"error_models_14_1.05.json",
	"error_models_15_1.05.json",
	"error_models_16_1.05.json",
	"error_models_17_1.05.json",
	"error_models_18_1.05.json",
	"error_models_19_1.05.json",
	"error_models_20_1.05.json",
	"error_models_21_1.05.json",
	"error_models_22_1.05.json",
	"error_models_23_1.05.json",
	"error_models_24_1.05.json",
	"error_models_25_1.05.json",
	"error_models_26_1.05.json",
	"error_models_27_1.05.json",
	"error_models_28_1.05.json",
	"error_models_29_1.05.json",
	"error_models_30_1.05.json",
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

	for x := 1; x <= n; x++ {
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

	nvme_dir := fmt.Sprintf("%s/docker-nvme-%d", os.Getenv("HOME"), index)

	mkdir := exec.Command(
		"mkdir",
		"-p",
		nvme_dir,
	)

	mkdir.Start()
	mkdir.Wait()

	request_copy(index, copy)

	error_models_prefix := "https://raw.githubusercontent.com/KTH/n-version-ethereum/neth/error_models/common"
	error_models_name := strings.Replace(error_models[exp_number], ".json", "", 1)
	path, err := os.Getwd()
	output_dir := fmt.Sprintf("%s/output-%s", path, error_models_name)

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
		"--privileged",
		"--rm",
		"--pid=host",
		"-v",
		fmt.Sprintf("%s:/root/nvme", nvme_dir),
		"-v",
		fmt.Sprintf("%s:/output", output_dir),
		"-e",
		"ETHERSCAN_API_KEY",
		fmt.Sprintf("javierron/neth:%s-kernel", target),
		"./single-version-controller.sh", //command
		target,
		fmt.Sprintf("%s/%s", error_models_prefix, error_models[exp_number]),
	)

	fmt.Printf("Begin experiment %s in disk %d\n", error_models_name, index)
	fmt.Println(cmd.String())

	outfile, err := os.Create(fmt.Sprintf("%s/docker-%s.log", output_dir, error_models_name))
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile

	errfile, err := os.Create(fmt.Sprintf("%s/docker-err-%s.log", output_dir, error_models_name))
	if err != nil {
		panic(err)
	}
	defer errfile.Close()

	cmd.Stderr = errfile

	cmd.Start()

	cmd.Wait()

	fmt.Printf("Exit experiment %s in disk %d\n", error_models_name, index)
	mstack.Done(index)
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

	outfile, err := os.Create(fmt.Sprintf("%s/sync-stop.log", os.Getenv("HOME")))
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile

	err = cmd.Run()
	if err != nil {
		println("failed to start first sync")
		return err
	}
	cmd.Wait()
	fmt.Println("done first sync")
	return nil
}

func synchronize(target string, stop chan int) {
	fmt.Println("init sync sript")
	cmd := exec.Command("./synchronize.sh", target)

	outfile, err := os.OpenFile(fmt.Sprintf("%s/sync.log", os.Getenv("HOME")), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile

	cmd.Start()

	<-stop
	fmt.Println("kill sync script")
	cmd.Process.Signal(os.Interrupt)

	fmt.Println("check for stop")
	cmd.Wait()

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

	fmt.Println(umount_source.String())
	umount_source.Run()

	umount_target := exec.Command(
		"sudo",
		"umount",
		target_partition_mount_pouint,
	)

	fmt.Println(umount_target.String())
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

	fmt.Println(cmd.String())

	outfile, err := os.Create(fmt.Sprintf("%s/dd_progress-%d.log", os.Getenv("HOME"), index))
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	//note: dd outputs to std err
	cmd.Stderr = outfile

	cmd.Run()
	cmd.Wait()

	mount_source := exec.Command(
		"sudo",
		"mount",
		source_partition,
		source_partition_mount_pouint,
	)

	fmt.Println(mount_source.String())
	mount_source.Run()

	mount_target := exec.Command(
		"sudo",
		"mount",
		target_partition,
		target_partition_mount_pouint,
	)

	fmt.Println(mount_target.String())
	mount_target.Run()

	mount_source.Wait()
	mount_target.Wait()
}

type State string

const (
	Syncing State = "SYNC"
	Copying State = "COPY"
)

func source_loop(target string, start chan int, copy chan CopyInfo, err_chan chan error) {
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
					go synchronize(target, stop)
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
		fmt.Printf("cannot parse number of disks argument: %s\n", os.Args[2])
		return
	}

	apikey := os.Getenv("ETHERSCAN_API_KEY")
	if apikey == "" {
		fmt.Println("ETHERSCAN_API_KEY env variable not set")
		return
	}

	disks := int(parsed)

	start_chan := make(chan int)
	copy_chan := make(chan CopyInfo)
	err_chan := make(chan error)
	go source_loop(target, start_chan, copy_chan, err_chan)
	go error_handler(err_chan)
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
		time.Sleep(30 * time.Second)
	}

}
