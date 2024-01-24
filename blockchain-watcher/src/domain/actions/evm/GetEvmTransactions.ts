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
      this.logger.info(`[${chain}][exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`);
      return [];
    }

    let populatedTransactions: EvmTransaction[] = [];
    const isTransactionsPresent = true;

    this.logger.info(`[${chain}][exec] Processing blocks [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`);
    for (let block = fromBlock; block <= toBlock; block++) {
      const evmBlock = await this.blockRepo.getBlock(chain, block, isTransactionsPresent);
      const transactions = evmBlock.transactions ?? [];

      // Only process transactions to the contract address configured
      const transactionsByAddressConfigured = transactions.filter(
        (transaction) =>
          opts.addresses?.includes(String(transaction.to).toLowerCase()) ||
          opts.addresses?.includes(String(transaction.from).toLowerCase())
      );

      if (transactionsByAddressConfigured.length > 0) {
        const hashNumbers = new Set(
          transactionsByAddressConfigured.map((transaction) => transaction.hash)
        );

        const batches = this.divideIntoBatches(hashNumbers);
        let combinedReceiptTransactions = {};

        for (const batch of batches) {
          const receiptTransactionsBatch = await this.blockRepo.getTransactionReceipt(chain, batch);
          combinedReceiptTransactions = {
            ...combinedReceiptTransactions,
            ...receiptTransactionsBatch
          };
        }

        const filterTransactions = this.filterTransactions(
          opts,
          transactionsByAddressConfigured,
          combinedReceiptTransactions
        );

        await this.populateTransaction(
          opts,
          evmBlock,
          combinedReceiptTransactions,
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

  /**
  * This method divide in batches the object to send, because we have one restriction about how many object send to the endpoint
  * the maximum is 10 object per request
  */
  divideIntoBatches(set: Set<string>) {
    const batchSize = 10;
    const batches = [];
    let batch: any[] = [];

    set.forEach(item => {
      batch.push(item);
      if (batch.length === batchSize) {
        batches.push(new Set(batch));
        batch = [];
      }
    });

    if (batch.length > 0) {
      batches.push(new Set(batch));
    }
    return batches;
  }

  private async populateTransaction(
    opts: GetEvmOpts,
    evmBlock: EvmBlock,
    combinedReceiptTransactions: Record<string, ReceiptTransaction>,
    filterTransactions: EvmTransaction[],
    populatedTransactions: EvmTransaction[]
  ) {
    filterTransactions.forEach((transaction) => {
      transaction.status = combinedReceiptTransactions[transaction.hash].status;
      transaction.timestamp = evmBlock.timestamp;
      transaction.environment = opts.environment;
      transaction.chainId = opts.chainId;
      transaction.chain = opts.chain;
      transaction.logs = combinedReceiptTransactions[transaction.hash].logs;
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
    combinedReceiptTransactions: Record<string, ReceiptTransaction>
  ): EvmTransaction[] {
    return transactionsByAddressConfigured.filter((transaction) => {
      const optsTopics = opts.topics;
      const logs = combinedReceiptTransactions[transaction.hash]?.logs || [];

      return logs.some((log) => {
        return optsTopics?.find((topic) => log.topics?.includes(topic));
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
