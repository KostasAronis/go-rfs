package miner

import (
	"encoding/json"
	"os"
)

//CommonMinerConfig struct describing the common configuration parameters shared by the miners
type CommonMinerConfig struct {
	//GenesisBlockHash The genesis (first) block MD5 hash for this blockchain
	GenesisBlockHash string
	//MinedCoinsPerOpBlock The number of record coins mined for an op block
	MinedCoinsPerOpBlock int
	//MinedCoinsPerNoOpBlock The number of record coins mined for a no-op block
	MinedCoinsPerNoOpBlock int
	//NumCoinsPerFileCreate The number of record coins charged for creating a file
	NumCoinsPerFileCreate int
	//GenOpBlockTimeout Time in milliseconds, the minimum time between op block mining (see diagram above)
	GenOpBlockTimeout int
	//PowPerOpBlock The op block difficulty (proof of work setting: number of zeroes)
	PowPerOpBlock int
	//PowPerNoOpBlock The no-op block difficulty (proof of work setting: number of zeroes)
	PowPerNoOpBlock int
	//ConfirmsPerFileCreate The number of confirmations for a create file operation (the number of blocks that must follow the block containing a create file operation along longest chain before the CreateFile call can return successfully)
	ConfirmsPerFileCreate int
	//ConfirmsPerFileAppend The number of confirmations for an append operation (the number of blocks that must follow the block containing an append operation along longest chain before the AppendRec call can return successfully). Note that this append confirm number will always be set to be larger than the create confirm number (above)
	ConfirmsPerFileAppend int
}

//MinerConfig struct describing the configuration for individual mienrs
type MinerConfig struct {
	//MinerID The ID of this miner (max 16 characters).
	MinerID string
	//PeerMinersAddrs An array of remote IP:port addresses, one per peer miner that this miner should connect to (using the OutgoingMinersIP below)
	PeerMinersAddrs []string
	//IncomingMinersAddr The local IP:port where the miner should expect other miners to connect to it (address it should listen on for connections from miners)
	IncomingMinersAddr string
	//OutgoingMinersIP The local IP that the miner should use to connect to peer miners
	OutgoingMinersIP string
	//IncomingClientsAddr The local IP:port where this miner should expect to receive connections from RFS clients (address it should listen on for connections from clients)
	IncomingClientsAddr string
	//CommonMinerConfig struct describing the common configuration parameters shared by the miners
	CommonMinerConfig CommonMinerConfig
}

//Load reads values from given filename
func (c MinerConfig) Load(str string) {
	f, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	err = decoder.Decode(&c)
	if err != nil {
		panic(err)
	}
}
