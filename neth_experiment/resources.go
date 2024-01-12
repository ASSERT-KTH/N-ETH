package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Clients     []ClientInfo
	Experiments []Experiment
}

type Experiment struct {
	Clients   []string
	Workloads []string
}

type ClientInfo struct {
	Name       string
	Port       string
	Disk_name  string
	Image_name string
	Nvme_index int
}

func RunClient(target ClientInfo, script string, tag string, wg *sync.WaitGroup, stop chan os.Signal, extra_args ...string) {

	fmt.Printf("init sync for %s\n", target.Name)

	nvme_dir := fmt.Sprintf("%s/%s", os.Getenv("HOME"), target.Disk_name)

	output_dir := fmt.Sprintf("%s/%s-%s", os.Getenv("OUTPUT_DIR"), target.Name, tag)
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
		fmt.Sprintf("%s:8545", target.Port),
		fmt.Sprintf(target.Image_name),
		script,
		target.Name,
	}

	args = append(args, extra_args...)
	cmd := exec.Command("docker", args...)

	fmt.Printf("Begin sync %s in path %s\n", target.Name, nvme_dir)
	fmt.Println(cmd.String())

	outfile, err := os.Create(fmt.Sprintf("%s/docker-sync-%s.log", output_dir, target.Name))
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile

	errfile, err := os.Create(fmt.Sprintf("%s/docker-sync-err-%s.log", output_dir, target.Name))
	if err != nil {
		panic(err)
	}
	defer errfile.Close()

	cmd.Stderr = errfile
	cmd.Start()

	sig := <-stop
	fmt.Printf("%s: %s sync script\n", target.Disk_name, sig)
	cmd.Process.Signal(sig)

	fmt.Printf("%s: check for stop\n", target.Disk_name)
	cmd.Wait()

	fmt.Printf("%s: done stopping\n", target.Disk_name)
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
					client.Name,
					exp_tag,
					client.Name,
				),
			)

			if err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("Client %s not yet started...\n", client.Name)

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

func SyncSourceClients(source_clients []ClientInfo, stop_chan chan struct {
	os.Signal
	bool
}, restart_chan chan int) {

	wg := new(sync.WaitGroup)

	for {
		wg.Add(len(source_clients))

		channel_slice := make([]chan os.Signal, 0, len(source_clients))

		for _, client := range source_clients {
			stop_clients_chan := make(chan os.Signal)
			channel_slice = append(channel_slice, stop_clients_chan)
			go RunClient(client, "./synchronize-resources.sh", "resources", wg, stop_clients_chan)
		}

		sig := <-stop_chan

		for _, ch := range channel_slice {
			ch <- sig.Signal
		}

		wg.Wait()

		stop_chan <- sig
		if sig.bool {
			<-restart_chan
		} else {
			return
		}
	}
}

func StartExperimentClients(experiment_clients []ClientInfo, script string, tag string, stop_chan chan os.Signal, extra_args ...string) {

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
	stop_chan <- sig
}

func CheckEnvs() {
	_, ok := os.LookupEnv("OUTPUT_DIR")

	if !ok {
		fmt.Println("Env variable OUTPUT_DIR not set!")
		os.Exit(-1)
	}
}

func CopyState(client ClientInfo, wg *sync.WaitGroup, target_index int) {

	source_partition := fmt.Sprintf("/dev/nvme%dn1p1", client.Nvme_index)
	source_partition_mount_point := fmt.Sprintf(
		"%s/docker-nvme-%s",
		os.Getenv("HOME"),
		client.Name,
	)

	target_partition := fmt.Sprintf("/dev/nvme%dn1p1", target_index)
	target_partition_mount_point := fmt.Sprintf(
		"%s-copy",
		source_partition_mount_point,
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
			client.Nvme_index,
			target_index,
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

func getFreeNvmeIndex(source_index int) int {
	//TODO: implement
	return source_index + 4
}

func CopyClients(source_clients []ClientInfo) {

	wg := new(sync.WaitGroup)
	wg.Add(len(source_clients))

	for _, client := range source_clients {
		target_index := getFreeNvmeIndex(client.Nvme_index)
		go CopyState(client, wg, target_index)
	}

	wg.Wait()
}

func RunProxy(tag string, n int, stop chan int) {
	args := []string{
		"adaptive",
		fmt.Sprintf("%d", n),
	}

	cmd := exec.Command("../proxy/proxy", args...)
	println(cmd.String())

	outfile, err := os.Create(
		fmt.Sprintf(
			"%s/proxy-%s.log",
			os.Getenv("OUTPUT_DIR"),
			tag,
		),
	)

	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile

	errfile, err := os.Create(
		fmt.Sprintf(
			"%s/proxy-err-%s.log",
			os.Getenv("OUTPUT_DIR"),
			tag,
		),
	)

	if err != nil {
		panic(err)
	}
	defer errfile.Close()
	cmd.Stderr = errfile

	cmd.Start()

	<-stop

	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()

	stop <- 0
}

func RunWorkload(experiment_tag string) {
	cmd := exec.Command(
		"../neth_workload/workload",
		"get_block",
		experiment_tag,
	)

	outfile, err := os.Create(
		fmt.Sprintf(
			"%s/workload-%s.log",
			os.Getenv("OUTPUT_DIR"),
			experiment_tag,
		),
	)

	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile

	cmd.Run()
}

func CleanSyncFlags(source_clients []ClientInfo, exp_tag string) {
	for i, client := range source_clients {
		err := os.Remove(
			fmt.Sprintf(
				"%s/%s-%s/ipc-%d.dat",
				os.Getenv("OUTPUT_DIR"),
				client.Disk_name,
				exp_tag,
				i,
			),
		)

		if err != nil && !os.IsNotExist(err) {
			panic(err)
		}
	}
}

func LoadAvailableClients(config *Config) map[string]ClientInfo {

	ret := make(map[string]ClientInfo)

	for _, config_client := range config.Clients {
		client := ClientInfo{
			Port:       "not-set",
			Name:       config_client.Name,
			Disk_name:  config_client.Disk_name,
			Image_name: config_client.Image_name,
			Nvme_index: config_client.Nvme_index,
		}
		ret[config_client.Name] = client
	}
	return ret
}

func ReadConfig() *Config {
	var ex Config // := new(Config)

	_, err := toml.DecodeFile("./config.toml", &ex)

	if err != nil {
		panic(err)
	}
	return &ex

}

func CreateExperimentClientList(
	experiment Experiment,
	avaliable_clients map[string]ClientInfo,
) []ClientInfo {

	initial_port := 8645

	experiment_clients := make([]ClientInfo, 0)

	for i, client_name := range experiment.Clients {
		client := avaliable_clients[client_name]
		client.Port = strconv.Itoa(initial_port + i)
		client.Image_name = client.Image_name + "-kernel"
		client.Disk_name = client.Disk_name + "-copy"
		experiment_clients = append(experiment_clients, client)
	}

	return experiment_clients
}

func CreateSourceClientList(
	experiment Experiment,
	avaliable_clients map[string]ClientInfo,
) []ClientInfo {

	initial_port := 8545

	//make a set of strings
	set := make(map[string]bool)

	//for each client in experiment.clients
	for _, client_name := range experiment.Clients {
		//add client to set
		set[client_name] = true
	}

	source_clients := make([]ClientInfo, 0)

	for name := range set {
		client := avaliable_clients[name]
		client.Port = strconv.Itoa(initial_port)
		source_clients = append(source_clients, client)
		initial_port++
	}

	return source_clients
}

func main() {

	CheckEnvs()

	config := ReadConfig()

	fmt.Print("read config\n")
	fmt.Printf("%v\n", config)

	available_clients := LoadAvailableClients(config)

	// for each experiment

	for _, exp := range config.Experiments {
		source_clients := CreateSourceClientList(exp, available_clients)

		fmt.Println("======================================")
		fmt.Printf("Source clients: %v\n", source_clients)
		fmt.Println("======================================")

		CleanSyncFlags(source_clients, "resources")

		stop_sync_chan := make(chan struct {
			os.Signal
			bool
		})
		restart_sync_chan := make(chan int)
		// sync targets
		go SyncSourceClients(source_clients, stop_sync_chan, restart_sync_chan)

		WaitForAllClientsSync(source_clients, "resources")
		fmt.Println("All clients synchronized!")

		proxy_stop := make(chan int)

		go RunProxy("resources", 1, proxy_stop)
		fmt.Println("Started proxy!")

		time.Sleep(180 * time.Second)

		fmt.Println("Running workload")
		//attach top to client

		RunWorkload("resources")
		fmt.Println("Workload Done!")

		fmt.Println("Shutting down proxy...")
		proxy_stop <- 0
		<-proxy_stop
		fmt.Println("Proxy shut down!")

		time.Sleep(10 * time.Second)

		fmt.Println("Closing source clients!")
		stop_sync_chan <- struct {
			os.Signal
			bool
		}{os.Interrupt, false}
		<-stop_sync_chan
		fmt.Println("Closed source clients!")
	}
}