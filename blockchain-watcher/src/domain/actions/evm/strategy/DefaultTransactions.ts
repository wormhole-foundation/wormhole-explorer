import { EvmBlock, EvmTransaction, ReceiptTransaction } from "../../../entities";
import { Filter, GetTransactions } from "./GetTransactionsStrategy";
import { EvmBlockRepository } from "../../../repositories";
import { GetEvmOpts } from "../PollEvm";

const TOPICS_APPLY = [
  "0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169",
  "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e",
  "0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e",
  "0xf6fc529540981400dc64edf649eb5e2e0eb5812a27f8c81bac2c1d317e71a5f0",
];

export class DefaultTransactions implements GetTransactions {
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

  apply(topics: string[]): boolean {
    return TOPICS_APPLY.includes(topics[0]);
  }

  async execute(filter: Filter): Promise<EvmTransaction[]> {
    let populatedTransactions: EvmTransaction[] = [];

    // Fetch logs from blockchain
    const logs = await this.blockRepo.getFilteredLogs(this.chain, {
      fromBlock: this.fromBlock,
      toBlock: this.toBlock,
      addresses: filter.addresses,
      topics: filter.topics,
    });

    if (logs.length > 0) {
      const blockNumbers = new Set(logs.map((log) => log.blockNumber));
      const blockHashes = new Set(logs.map((log) => log.blockHash));
      const txHashes = new Set(logs.map((log) => log.transactionHash));

      // Fetch blocks and transaction receipts from blockchain
      const evmBlocks = await this.blockRepo.getBlocks(this.chain, blockNumbers, true);

      if (evmBlocks) {
        const transactionsMap: EvmTransaction[] = [];

        for (const blockHash of blockHashes) {
          const transactions = evmBlocks[blockHash]?.transactions || [];

          // Collect transactions that are in the txHashes set
          transactions.forEach((transaction) => {
            if (txHashes.has(transaction.hash)) {
              transactionsMap.push(transaction);
            }
          });
        }

        // Fetch transaction receipts from blockchain
        const receiptTransactions = await this.blockRepo.getTransactionReceipt(
          this.chain,
          new Set(transactionsMap.map((tx) => tx.hash))
        );

        // Populate transactions
        this.populateTransaction(
          this.opts,
          evmBlocks,
          receiptTransactions,
          transactionsMap,
          populatedTransactions
        );
      }
    }
    return populatedTransactions;
  }

  private populateTransaction(
    opts: GetEvmOpts,
    evmBlocks: Record<string, EvmBlock>,
    receiptTransactions: Record<string, ReceiptTransaction>,
    filterTransactions: EvmTransaction[],
    populatedTransactions: EvmTransaction[]
  ) {
    filterTransactions.forEach((transaction) => {
      transaction.status = receiptTransactions[transaction.hash]?.status;
      transaction.timestamp = evmBlocks[transaction.blockHash]?.timestamp;
      transaction.environment = opts.environment;
      transaction.chainId = opts.chainId;
      transaction.chain = opts.chain;
      transaction.logs = receiptTransactions[transaction.hash]?.logs;

      if (transaction.status) {
        populatedTransactions.push(transaction);
      }
    });
  }
}
