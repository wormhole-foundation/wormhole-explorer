package watcher

func getTotalBlocks(lastBlock, currentBlock, pageSize uint64) uint64 {
	return (lastBlock-currentBlock)/pageSize + 1
}

func getPage(currentBlock, index, pageSize, lastBlock uint64) (uint64, uint64) {
	fromBlock := currentBlock + index*pageSize
	toBlock := fromBlock + pageSize - 1
	if toBlock > lastBlock {
		toBlock = lastBlock
	}
	return fromBlock, toBlock
}
