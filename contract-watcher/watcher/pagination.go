package watcher

func getTotalBlocks(lastBlock, currentBlock, maxBlocks uint64) uint64 {
	return (lastBlock-currentBlock)/maxBlocks + 1
}

func getPage(currentBlock, index, maxBlocks, lastBlock uint64) (uint64, uint64) {
	fromBlock := currentBlock + index*maxBlocks
	toBlock := fromBlock + maxBlocks - 1
	if toBlock > lastBlock {
		toBlock = lastBlock
	}
	return fromBlock, toBlock
}
