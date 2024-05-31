import { Filter, GetTransactions, populateTransaction } from "../GetEvmTransactions";
import { EvmBlockRepository } from "../../../repositories";
import { EvmTransaction } from "../../../entities";
import { GetEvmOpts } from "../PollEvm";

export class GetTransactionsByFiltersStrategy implements GetTransactions {
  private readonly blockRepo: EvmBlockRepository;
  private readonly fromBlock: bigint;
  private readonly toBlock: bigint;
  private readonly chain: string;
  private readonly opts: GetEvmOpts;

  constructor(
    blockRepo: EvmBlockRepository,
    fromBlock: bigint,
    toBlock: bigint,
    chain: string,
    opts: GetEvmOpts
  ) {
    this.blockRepo = blockRepo;
    this.fromBlock = fromBlock;
    this.toBlock = toBlock;
    this.chain = chain;
    this.opts = opts;
  }

  appliesTo(strategy: string): boolean {
    return strategy == GetTransactionsByFiltersStrategy.name;
  }

  async execute(filter: Filter): Promise<EvmTransaction[]> {
    const populatedTransactions: EvmTransaction[] = [];

    const logs = await this.blockRepo.getFilteredLogs(this.chain, {
      fromBlock: this.fromBlock,
      toBlock: this.toBlock,
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
      const evmBlocks = await this.blockRepo.getBlocks(this.chain, blockNumbers, true);

      if (evmBlocks) {
        const filterTransactions: EvmTransaction[] = [];

        for (const blockHash of blockHashes) {
          const transactions = evmBlocks[blockHash]?.transactions || [];

          // Collect complete transactions from the block by hash
          const filtered = transactions.filter((transaction) => txHashes.has(transaction.hash));
          filterTransactions.push(...filtered);
        }

        // Get transaction details
        const transactionReceipts = await this.blockRepo.getTransactionReceipt(
          this.chain,
          txHashes
        );

        populateTransaction(
          this.opts,
          evmBlocks,
          transactionReceipts,
          filterTransactions,
          populatedTransactions
        );
      }
    }
    return populatedTransactions;
  }
}
