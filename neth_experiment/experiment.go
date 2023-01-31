package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var error_models_prefix = "https://raw.githubusercontent.com/KTH/n-version-ethereum/neth/error_models/common"
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

type ClientInfo struct {
	name         string
	port         string
	disk_name    string
	image_name   string
	source_index int
	target_index int
}

var clients = []ClientInfo{
	{
		name:         "geth",
		port:         "8545",
		disk_name:    "geth",
		image_name:   "javierron/neth:geth",
		source_index: 1,
		target_index: 5,
	},
	{
		name:         "besu",
		port:         "8546",
		disk_name:    "besu",
		image_name:   "javierron/neth:besu",
		source_index: 3,
		target_index: 7,
	},
	{
		name:         "erigon",
		port:         "8547",
		disk_name:    "erigon",
		image_name:   "javierron/neth:erigon",
		source_index: 2,
		target_index: 6,
	},
	{
		name:         "nethermind",
		port:         "8548",
		disk_name:    "nethermind",
		image_name:   "javierron/neth:nethermind",
		source_index: 0,
		target_index: 4,
	},
}

var experiment_clients = []ClientInfo{
	{
		name:       "geth",
		port:       "8555",
		disk_name:  "geth-copy",
		image_name: "javierron/neth:geth-kernel",
	},
	{
		name:       "besu",
		port:       "8556",
		disk_name:  "besu-copy",
		image_name: "javierron/neth:besu-kernel",
	},
	{
		name:       "erigon",
		port:       "8557",
		disk_name:  "erigon-copy",
		image_name: "javierron/neth:erigon-kernel",
	},
	{
		name:       "nethermind",
		port:       "8558",
		disk_name:  "nethermind-copy",
		image_name: "javierron/neth:nethermind-kernel",
	},
}

func RunClient(target ClientInfo, script string, tag string, wg *sync.WaitGroup, stop chan os.Signal, extra_args ...string) {

	fmt.Printf("init sync for %s\n", target.name)

	nvme_dir := fmt.Sprintf("%s/docker-nvme-%s", os.Getenv("HOME"), target.disk_name)

	output_dir := fmt.Sprintf("%s/%s-%s", os.Getenv("OUTPUT_DIR"), target.name, tag)
	err := os.Mkdir(output_dir, 0775)

	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	args := []string{
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
		fmt.Sprintf(target.image_name),
		script,
		target.name,
	}

	args = append(args, extra_args...)
	cmd := exec.Command("docker", args...)

	fmt.Printf("Begin sync %s in path %s\n", target.name, nvme_dir)
	fmt.Println(cmd.String())

	outfile, err := os.Create(fmt.Sprintf("%s/docker-sync-%s.log", output_dir, target.name))
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile

	errfile, err := os.Create(fmt.Sprintf("%s/docker-sync-err-%s.log", output_dir, target.disk_name))
	if err != nil {
		panic(err)
	}
	defer errfile.Close()

	cmd.Stderr = errfile
	cmd.Start()

	sig := <-stop
	fmt.Printf("%s: %s sync script\n", target.disk_name, sig)
	cmd.Process.Signal(sig)

	fmt.Printf("%s: check for stop\n", target.disk_name)
	cmd.Wait()

	fmt.Printf("%s: done stopping\n", target.disk_name)
	wg.Done()
}

func WaitForAllClientsSync(poll_clients []ClientInfo, exp_tag string) {
	all_clients_ready := false
	for !all_clients_ready {
		all_clients_ready = true
		for _, client := range poll_clients {
			dat, err := os.ReadFile(
				//fmt.Sprintf("%s/%s-%s", os.Getenv("OUTPUT_DIR"), target.disk_name, error_model_index)
				fmt.Sprintf(
					"%s/%s-%s/ipc-%s.dat",
					os.Getenv("OUTPUT_DIR"),
					client.disk_name,
					exp_tag,
					client.name,
				),
			)

			if err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("Client %s not yet started...\n", client.name)

					all_clients_ready = all_clients_ready && false
					continue
				} else {
					panic(err)
				}
			}

			all_clients_ready = all_clients_ready && (string(dat) == "READY\n")
		}

		if !all_clients_ready {
			fmt.Println("Clients are synchronizing...")
			time.Sleep(10 * time.Second)
		}
	}
	CleanSyncFlags(poll_clients, exp_tag)
}

func SyncSourceClients(stop_chan chan os.Signal, restart_chan chan int) {

	wg := new(sync.WaitGroup)

	for {
		wg.Add(len(clients))

		channel_slice := make([]chan os.Signal, 0, len(clients))

		for _, client := range clients {
			stop_clients_chan := make(chan os.Signal)
			channel_slice = append(channel_slice, stop_clients_chan)
			go RunClient(client, "./synchronize-ready.sh", "source", wg, stop_clients_chan)
		}

		sig := <-stop_chan

		for _, ch := range channel_slice {
			ch <- sig
		}

		wg.Wait()

		stop_chan <- nil
		<-restart_chan
	}
}

func StartExperimentClients(script string, tag string, stop_chan chan os.Signal, extra_args ...string) {

	wg := new(sync.WaitGroup)

	channel_slice := make([]chan os.Signal, 0, len(experiment_clients))
	wg.Add(len(experiment_clients))

	for _, client := range experiment_clients {
		stop_clients_chan := make(chan os.Signal)
		channel_slice = append(channel_slice, stop_clients_chan)
		go RunClient(client, script, tag, wg, stop_clients_chan, extra_args...)
	}

	sig := <-stop_chan

	for _, ch := range channel_slice {
		ch <- sig
	}

	wg.Wait()
	stop_chan <- nil
}

func CheckEnvs() {
	_, ok := os.LookupEnv("OUTPUT_DIR")

	if !ok {
		fmt.Println("Env variable OUTPUT_DIR not set!")
		os.Exit(-1)
	}
}

func CopyState(client ClientInfo, wg *sync.WaitGroup) {

	source_partition := fmt.Sprintf("/dev/nvme%dn1p1", client.source_index)
	source_partition_mount_point := fmt.Sprintf(
		"%s/docker-nvme-%s",
		os.Getenv("HOME"),
		client.name,
	)

	target_partition := fmt.Sprintf("/dev/nvme%dn1p1", client.target_index)
	target_partition_mount_point := fmt.Sprintf(
		"%s/docker-nvme-%s-copy",
		os.Getenv("HOME"),
		client.name,
	)

	umount_source := exec.Command(
		"sudo",
		"umount",
		source_partition_mount_point,
	)

	fmt.Println(umount_source.String())
	umount_source.Start()

	umount_target := exec.Command(
		"sudo",
		"umount",
		target_partition_mount_point,
	)

	fmt.Println(umount_target.String())
	umount_target.Start()

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

	outfile, err := os.Create(
		fmt.Sprintf(
			"%s/dd_progress-%d-%d.log",
			os.Getenv("HOME"),
			client.source_index,
			client.target_index,
		),
	)

	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	//note: dd outputs to std err
	cmd.Stderr = outfile

	cmd.Run()

	mount_source := exec.Command(
		"sudo",
		"mount",
		source_partition,
		source_partition_mount_point,
	)

	fmt.Println(mount_source.String())
	mount_source.Start()

	mount_target := exec.Command(
		"sudo",
		"mount",
		target_partition,
		target_partition_mount_point,
	)

	fmt.Println(mount_target.String())
	mount_target.Start()

	mount_source.Wait()
	mount_target.Wait()

	wg.Done()
}

func CopyClients() {

	//stop clients

	wg := new(sync.WaitGroup)
	wg.Add(len(clients))

	for _, client := range clients {
		go CopyState(client, wg)
	}

	wg.Wait()

	// start clients
}

func RunProxy() {
	args := []string{
		"run",
		"-p",
		"8080:8080",
		"javierron/neth:proxy",
		"./proxy",
		"adaptive",
	}

	outfile, err := os.Create(
		fmt.Sprintf(
			"%s/proxy.log",
			os.Getenv("OUTPUT_DIR"),
		),
	)

	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd := exec.Command("docker", args...)
	println(cmd.String())
	cmd.Stdout = outfile
	cmd.Run()
}

func RunWorkload(experiment_tag string) {
	cmd := exec.Command(
		"go",
		"run",
		"requests-get-block.go",
		experiment_tag,
	)
	cmd.Run()
}

func CleanSyncFlags(poll_clients []ClientInfo, exp_tag string) {
	for _, client := range clients {
		err := os.Remove(
			fmt.Sprintf(
				"%s/%s-%s/ipc-%s.dat",
				os.Getenv("OUTPUT_DIR"),
				client.disk_name,
				exp_tag,
				client.name,
			),
		)

		if err != nil && !os.IsNotExist(err) {
			panic(err)
		}
	}
}

func main() {

	CheckEnvs()
	CleanSyncFlags(clients, "source")
	RunProxy()

	stop_sync_chan := make(chan os.Signal)
	restart_sync_chan := make(chan int)
	// sync targets
	go SyncSourceClients(stop_sync_chan, restart_sync_chan)

	WaitForAllClientsSync(clients, "source")
	fmt.Println("All clients synchronized!")

	//foreach error model
	for error_model_index, error_model := range error_models {

		//   stop sync
		stop_sync_chan <- os.Interrupt
		<-stop_sync_chan
		fmt.Println("All clients stopped!")

		// copy targets
		CopyClients()
		fmt.Println("All clients copied!")
		restart_sync_chan <- 1
		fmt.Println("Restarting source clients sync!")

		// sync copies
		fmt.Println("Starting experiment clients sync!")

		error_model_tag := strconv.Itoa(error_model_index)
		experiments_sync_chan := make(chan os.Signal)

		go StartExperimentClients(
			"./synchronize-ready.sh",
			fmt.Sprintf("pre-sync-%s", error_model_tag),
			experiments_sync_chan,
		)

		WaitForAllClientsSync(experiment_clients, fmt.Sprintf("pre-sync-%s", error_model_tag))
		fmt.Println("Experiment clients synced!")

		experiments_sync_chan <- os.Interrupt
		<-experiments_sync_chan

		fmt.Println("Experiment clients stopped!")

		fmt.Println("Starting experiment clients with fault injection!")
		error_model_path := fmt.Sprintf("%s/%s", error_models_prefix, error_model)
		go StartExperimentClients(
			"./n-version-fault-injection.sh",
			error_model_tag,
			experiments_sync_chan,
			error_model_path,
		)

		fmt.Println("Started experiment clients with fault injection!")

		time.Sleep(30 * time.Second)

		fmt.Println("Running workload")
		RunWorkload(error_model_tag)
		fmt.Println("Workload Done!")

		fmt.Println("Closing experiment clients!")
		experiments_sync_chan <- os.Kill

		fmt.Println("Cleaning up containers...")
		<-experiments_sync_chan

		time.Sleep(10 * time.Second)
	}
}
