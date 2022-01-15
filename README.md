# n-version-ethereum
N-Version Implementation of Ethereum Nodes

## Clone


`git clone --recurse-submodules git@github.com:KTH/n-version-ethereum.git && git submodule update --remote`

## Components

### P2P proxy

Code at go-eth submodule: [go-ethereum/p2p-server](https://github.com/javierron/go-ethereum/tree/p2p-server)
Docker image: [javierron/geth:p2p](https://hub.docker.com/repository/docker/javierron/geth)

### go-ethereum subnode

Code at go-eth submodule: [go-ethereum/subnode-3](https://github.com/javierron/go-ethereum/tree/subnode-3)
Docker image: [javierron/geth:subnode](https://hub.docker.com/repository/docker/javierron/geth)

## Executing

docker-compose file at ethereum-docker submodule [go-ethereum/geth-v1.10.12](https://github.com/gluckzhang/ethereum-docker/tree/geth-v1.10.12)