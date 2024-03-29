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
    const isTransactionsPresent = true;

    this.logger.info(
      `[${chain}][exec] Processing blocks [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );

    const blockNumbers: Set<bigint> = new Set();
    for (let block = fromBlock; block <= toBlock; block++) {
      blockNumbers.add(block);
    }
    const evmBlocks = await this.blockRepo.getBlocks(chain, blockNumbers, isTransactionsPresent);

    for (const blockKey in evmBlocks) {
      const evmBlock = evmBlocks[blockKey];
      const transactions = evmBlock.transactions ?? [];

      // Only process transactions to the contract address configured
      const transactionsByAddressConfigured = transactions.filter(
        (transaction) =>
          opts.addresses?.includes(String(transaction.from).toLowerCase()) ||
          opts.addresses?.includes(String(transaction.to).toLowerCase())
      );

      if (transactionsByAddressConfigured.length > 0) {
        const hashNumbers = new Set(transactionsByAddressConfigured.map((tx) => tx.hash));
        const receiptTransactions = await this.blockRepo.getTransactionReceipt(chain, hashNumbers);

        const filterTransactions = this.filterTransactions(
          opts,
          transactionsByAddressConfigured,
          receiptTransactions
        );

        await this.populateTransaction(
          opts,
          evmBlock,
          receiptTransactions,
          filterTransactions,
          populatedTransactions
        );
      }
    }

    this.logger.info(
      `[${chain}][exec] Got ${
        populatedTransactions?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
    return populatedTransactions;
  }

  private async populateTransaction(
    opts: GetEvmOpts,
    evmBlock: EvmBlock,
    receiptTransactions: Record<string, ReceiptTransaction>,
    filterTransactions: EvmTransaction[],
    populatedTransactions: EvmTransaction[]
  ) {
    filterTransactions.forEach((transaction) => {
      transaction.status = receiptTransactions[transaction.hash].status;
      transaction.timestamp = evmBlock.timestamp;
      transaction.environment = opts.environment;
      transaction.chainId = opts.chainId;
      transaction.chain = opts.chain;
      transaction.logs = receiptTransactions[transaction.hash].logs;
      populatedTransactions.push(transaction);
    });
  }

  /**
   * This method filter the transactions in base your logs with the topic and address configured in the job
   * For example: Redeemed or MintAndWithdraw transactions
   */
  private filterTransactions(
    opts: GetEvmOpts,
    transactionsByAddressConfigured: EvmTransaction[],
    receiptTransactions: Record<string, ReceiptTransaction>
  ): EvmTransaction[] {
    return transactionsByAddressConfigured.filter((transaction) => {
      const optsTopics = opts.topics || [];
      const logs = receiptTransactions[transaction.hash]?.logs || [];

      return optsTopics.some((topicsFilter) => {
        // if the filter is an array, we need to check if all desired topics are present in the logs
        if (Array.isArray(topicsFilter)) {
          return topicsFilter.every((tf) => logs.some((log) => log.topics.some((t) => t === tf)));
        }

        // if the filter is a string, we need to check if it's present in any of the logs
        return logs.some((log) => log.topics.some((t) => t === topicsFilter));
      });
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
