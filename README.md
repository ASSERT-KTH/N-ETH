# N-ETH

A high-availability N-version Ethereum node protoype 

## Description

This repository contains necessary code to deploy N-ETH and measure its availability under unstable execution environments.

N-ETH is an N-version Ethereum Node protoype which aims to improve API availability. N-ETH builds on the existing diversity of Ethereum implementations.

To this case, a N-ETH node will execute one instance of each: geth, besu, erigon, and nethermind as _verisons_ or _sub-nodes_; and a proxy to encapsulate the sub-nodes under a single interface.

Also, this repository contains automated experiments to measure availability of Etherum clients: geth, besu, erigon, and nethermind.

### Deployment

The dockerfiles ([geth](geth.dockerfile), [besu](besu.dockerfile), [erigon](erigon.dockerfile), and [nethermind](nethermind.dockerfile)) contain the necessary files to execute them along with fault injection modules.

To build a version run e.g. `docker build geth.dockerfile -t neth/geth`. The resulting image will contain the Ethreum implementation built from source, Teku as consensus layer node, also built from source, and necessary libraries to perform fault injection.

Given that fault injection requires linux headers, these will need to be installed depending on the executing OS. [This file](kernel-headers.dockerfile) is used to install the headers in the docker images.

The docker image must be run with one of the following commands
- `synchronize.sh <node_name>` syncs the node indefinitely. 
- `synchronize_stop.sh <node_name>` syncs the node and exits.
- `single-version-fault-injection.sh <node_name> <fault_injection_strat>` starts the node and applies fault injection acordding to the requiered strategy.

The CLI parameters of the nodes are read from the [config file](config.toml)

[neth_experiment/experiment.go](`neth_experiment/experiment.go`) runs a N-ETH node with pre syncronization and fault injection. This uses the nodes' docker images, and also requieres the [proxy image](proxy/dockerfile)

### Requirements

`docker`
`go`
`jq` 
`stoml`

## N-ETH Experiment

### Workloads

- [Random RPC workload](requests-random.go)
- [Single RPC workload](requests-get-block.go)

### Fault injection strategies

FI strategies are located in [this directory](error_models)

### Experiment data

Contact @javierron for access to experiment data.

## Misc

This repository also contains SSD formatting and mount scripts for experiment automation. 
