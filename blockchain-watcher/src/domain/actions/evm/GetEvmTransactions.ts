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

    for (const topic in opts.topics) {
      const maptopic = opts.topics[Number(topic)];

      const logs = await this.blockRepo.getFilteredLogs(opts.chain, {
        fromBlock,
        toBlock,
        addresses: opts.addresses ?? [],
        topics: [String(maptopic)] ?? [],
      });

      if (logs.length > 0) {
        try {
          // Extract block numbers and transaction hashes from logs
          const blockNumbers = new Set(logs.map((log) => log.blockNumber));
          const hashNumbers = new Set(logs.map((log) => log.transactionHash));
          const blockHash = new Set(logs.map((log) => log.blockHash));

          const [evmBlocks, receiptTransactions] = await Promise.all([
            this.blockRepo.getBlocks(chain, blockNumbers, true),
            this.blockRepo.getTransactionReceipt(chain, hashNumbers),
          ]);

          const transactionsMap: EvmTransaction[] = [];
          blockHash.forEach((hash) => {
            const transactions = evmBlocks[hash].transactions;

            for (const transaction of transactions!) {
              const is = hashNumbers.has(transaction.hash);

              if (is) [transactionsMap.push(transaction)];
            }
          });

          this.populateTransaction(
            opts,
            evmBlocks,
            receiptTransactions,
            transactionsMap,
            populatedTransactions
          );
        } catch (error) {
          // Handle errors
          console.error("An error occurred while fetching blockchain data:", error);
        }
      }
    }

    const filterTransactions = this.removeDuplicates(populatedTransactions);

    this.logger.info(
      `[${chain}][exec] Got ${
        filterTransactions?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
    return filterTransactions;
  }

  private populateTransaction(
    opts: GetEvmOpts,
    evmBlocks: Record<string, EvmBlock>,
    receiptTransactions: Record<string, ReceiptTransaction>,
    filterTransactions: EvmTransaction[],
    populatedTransactions: EvmTransaction[]
  ) {
    filterTransactions.forEach((transaction) => {
      transaction.status = receiptTransactions[transaction.hash].status;
      transaction.timestamp = evmBlocks[transaction.blockHash].timestamp;
      transaction.environment = opts.environment;
      transaction.chainId = opts.chainId;
      transaction.chain = opts.chain;
      transaction.logs = receiptTransactions[transaction.hash].logs;
      populatedTransactions.push(transaction);
    });
  }

  private populateLog(opts: GetEvmOpts, fromBlock: bigint, toBlock: bigint): string {
    return `[addresses:${opts.addresses}][topics:${opts.topics}][blocks:${fromBlock} - ${toBlock}]`;
  }

  private removeDuplicates<T>(arr: T[]): T[] {
    return arr.filter(
      (item, index, self) =>
        index === self.findIndex((t) => JSON.stringify(t) === JSON.stringify(item))
    );
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
