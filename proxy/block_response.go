package main

type RPCResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  Block  `json:"result"`
}

type Block struct {
	BaseFeePerGas    interface{} `json:"baseFeePerGas"`
	Difficulty       interface{} `json:"difficulty"`
	ExtraData        interface{} `json:"extraData"`
	GasLimit         interface{} `json:"gasLimit"`
	GasUsed          interface{} `json:"gasUsed"`
	Hash             interface{} `json:"hash"`
	LogsBloom        interface{} `json:"logsBloom"`
	Miner            interface{} `json:"miner"`
	MixHash          interface{} `json:"mixHash"`
	Nonce            interface{} `json:"nonce"`
	Number           string      `json:"number"`
	ParentHash       interface{} `json:"parentHash"`
	ReceiptsRoot     interface{} `json:"receiptsRoot"`
	Sha3Uncles       interface{} `json:"sha3Uncles"`
	Size             interface{} `json:"size"`
	StateRoot        interface{} `json:"stateRoot"`
	Timestamp        interface{} `json:"timestamp"`
	TotalDifficulty  interface{} `json:"totalDifficulty"`
	Transactions     interface{} `json:"transactions"`
	TransactionsRoot interface{} `json:"transactionsRoot"`
	Uncles           interface{} `json:"uncles"`
}
