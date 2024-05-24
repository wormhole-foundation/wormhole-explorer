import { EvmBlock, EvmTransaction, ReceiptTransaction } from "../../entities";
import { EvmBlockRepository } from "../../repositories";
import { GetEvmOpts } from "./GetEvmLogs";
import winston from "winston";

export class GetEvmTransactions {
  private readonly blockRepo: EvmBlockRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: EvmBlockRepository) {
    this.logger = winston.child({ module: "GetEvmTransactions" });
    this.blockRepo = blockRepo;
  }

  async execute(range: Range, opts: GetEvmOpts): Promise<EvmTransaction[]> {
    const fromBlock = range.fromBlock;
    const toBlock = range.toBlock;
    const chain = opts.chain;

    if (fromBlock > toBlock) {
      this.logger.info(
        `[${chain}][exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
      );
      return [];
    }

    let populatedTransactions: EvmTransaction[] = [];

    this.logger.info(
      `[${chain}][exec] Processing blocks [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );

    // Fetch logs from blockchain
    const logs = await this.blockRepo.getFilteredLogs(opts.chain, {
      fromBlock,
      toBlock,
      addresses: [],
      topics: [],
    });

    // Filter logs by topics
    const filterLogsByTopics = [];
    for (const log of logs) {
      if (opts.topics?.includes(log.topics[0])) {
        filterLogsByTopics.push(log);
      }
    }

    if (filterLogsByTopics.length > 0) {
      try {
        const blockNumbers = new Set(filterLogsByTopics.map((log) => log.blockNumber));
        const blockHash = new Set(filterLogsByTopics.map((log) => log.blockHash));

        // Fetch blocks and transaction receipts from blockchain
        const evmBlocks = await this.blockRepo.getBlocks(opts.chain, blockNumbers, true);

        if (evmBlocks) {
          const transactionsMap: EvmTransaction[] = [];

          for (const hash of blockHash) {
            const transactions = evmBlocks[hash]?.transactions || [];

            // Collect transactions
            transactions?.forEach((transaction) => {
              transactionsMap.push(transaction);
            });
          }

          // Fetch transaction receipts from blockchain
          const hashNumbers = new Set(transactionsMap.map((tx) => tx.hash));
          const receiptTransactions = await this.blockRepo.getTransactionReceipt(
            opts.chain,
            hashNumbers
          );

          // Populate transactions
          this.populateTransaction(
            opts,
            evmBlocks,
            receiptTransactions,
            transactionsMap,
            populatedTransactions
          );
        }
      } catch (e) {
        // Handle errors
        console.error("3- TEST error:", e);
      }
    }

    this.logger.info(
      `[${chain}][exec] Got ${
        populatedTransactions?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
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

  private populateLog(opts: GetEvmOpts, fromBlock: bigint, toBlock: bigint): string {
    return `[addresses:${opts.addresses}][topics:${opts.topics}][blocks:${fromBlock} - ${toBlock}]`;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
