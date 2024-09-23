import { Filter, GetTransactions, populateTransaction } from "../GetEvmTransactions";
import { EvmBlockRepository } from "../../../repositories";
import { EvmTransaction } from "../../../entities";
import { GetEvmOpts } from "../PollEvm";

export class GetTransactionsByLogFiltersStrategy implements GetTransactions {
  private readonly blockRepo: EvmBlockRepository;

  constructor(blockRepo: EvmBlockRepository) {
    this.blockRepo = blockRepo;
  }

  appliesTo(addresses: string[], topics: string[]): boolean {
    return addresses.length == 0 && topics.length > 0; // Process only by topics
  }

  async execute(
    filter: Filter,
    fromBlock: bigint,
    toBlock: bigint,
    opts: GetEvmOpts
  ): Promise<EvmTransaction[]> {
    const populatedTransactions: EvmTransaction[] = [];
    const chain = opts.chain;

    const logs = await this.blockRepo.getFilteredLogs(chain, {
      fromBlock: fromBlock,
      toBlock: toBlock,
      addresses: filter.addresses,
      topics: filter.topics,
    });

    if (logs.length > 0) {
      const blockNumbers = new Set<bigint>();
      const blockHashes = new Set<string>();
      const txHashes = new Set<string>();

      logs.forEach((log) => {
        blockNumbers.add(log.blockNumber);
        blockHashes.add(log.blockHash);
        txHashes.add(log.transactionHash);
      });

      // Get blocks with your transactions
      const evmBlocks = await this.blockRepo.getBlocks(chain, blockNumbers, true);

      if (evmBlocks) {
        const filterTransactions: EvmTransaction[] = [];

        for (const blockHash of blockHashes) {
          const transactions = evmBlocks[blockHash]?.transactions || [];

          // Collect complete transactions from the block by hash
          const filtered = transactions.filter((transaction) => txHashes.has(transaction.hash));
          filterTransactions.push(...filtered);
        }

        // Get transaction details
        const transactionReceipts = await this.blockRepo.getTransactionReceipt(chain, txHashes);

        populateTransaction(
          opts,
          evmBlocks,
          transactionReceipts,
          filterTransactions,
          populatedTransactions,
          filter.topics
        );
      }
    }
    return populatedTransactions;
  }
}
