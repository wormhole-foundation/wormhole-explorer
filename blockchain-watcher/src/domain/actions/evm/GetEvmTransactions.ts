import { EvmBlock, EvmTransaction, ReceiptTransaction } from "../../entities";
import { EvmBlockRepository } from "../../repositories";
import { GetEvmOpts } from "./PollEvm";
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

    for (const filter of opts.filters!) {
      // Fetch logs from blockchain
      const logs = await this.blockRepo.getFilteredLogs(opts.chain, {
        fromBlock,
        toBlock,
        addresses: filter.addresses,
        topics: filter.topics,
      });

      if (logs.length > 0) {
        const blockNumbers = new Set(logs.map((log) => log.blockNumber));
        const blockHashes = new Set(logs.map((log) => log.blockHash));
        const txHashes = new Set(logs.map((log) => log.transactionHash));

        // Fetch blocks and transaction receipts from blockchain
        const evmBlocks = await this.blockRepo.getBlocks(opts.chain, blockNumbers, true);

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
            opts.chain,
            new Set(transactionsMap.map((tx) => tx.hash))
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
      }
    }

    this.logger.info(
      `[${chain}][exec] Got ${populatedTransactions?.length} transactions to process [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
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
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
