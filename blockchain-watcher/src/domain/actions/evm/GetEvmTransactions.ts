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

    if (fromBlock > toBlock) {
      this.logger.info(`[exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`);
      return [];
    }

    let populatedTransactions: EvmTransaction[] = [];
    const isTransactionsPresent = true;
    const chain = opts.chain;

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
        const receiptTransaction = await this.blockRepo.getTransactionReceipt(chain, hashNumbers);

        const filterTransactions = this.filterTransactions(
          opts,
          transactionsByAddressConfigured,
          receiptTransaction
        );

        await this.populateTransaction(
          opts,
          evmBlock,
          receiptTransaction,
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
    receiptTransaction: Record<string, ReceiptTransaction>,
    filterTransactions: EvmTransaction[],
    populatedTransactions: EvmTransaction[]
  ) {
    filterTransactions.forEach((transaction) => {
      transaction.status = receiptTransaction[transaction.hash].status;
      transaction.timestamp = evmBlock.timestamp;
      transaction.environment = opts.environment;
      transaction.chainId = opts.chainId;
      transaction.chain = opts.chain;
      transaction.logs = receiptTransaction[transaction.hash].logs;
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
    receiptTransaction: Record<string, ReceiptTransaction>
  ): EvmTransaction[] {
    return transactionsByAddressConfigured.filter((transaction) => {
      const optsTopics = opts.topics;
      const logs = receiptTransaction[transaction.hash]?.logs || [];

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
