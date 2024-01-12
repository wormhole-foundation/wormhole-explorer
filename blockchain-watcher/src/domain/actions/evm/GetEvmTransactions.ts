import { EvmBlock, EvmTransaction, ReceiptTransaction } from "../../entities";
import { EvmBlockRepository } from "../../repositories";
import { GetEvmOpts } from "./GetEvmLogs";
import winston from "winston";

const REDEEMED_TOPIC = "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e";

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

    let populateTransactions: EvmTransaction[] = [];
    const isTransactionsPresent = true;
    const chain = opts.chain;

    for (let block = fromBlock; block <= toBlock; block++) {
      const evmBlock = await this.blockRepo.getBlock(chain, block, isTransactionsPresent);
      const transactions = evmBlock.transactions ?? [];

      // Only process transactions to the contract address configured
      const transactionsByaddressConfigured = transactions.filter(
        (transaction) =>
          opts.addresses?.includes(String(transaction.to).toLowerCase()) ||
          opts.addresses?.includes(String(transaction.from).toLowerCase())
      );

      if (transactionsByaddressConfigured.length > 0) {
        const hashNumbers = new Set(
          transactionsByaddressConfigured.map((transaction) => transaction.hash)
        );
        const receiptTransaction = await this.blockRepo.getTransactionReceipt(chain, hashNumbers);

        const filterTransactions = this.filterTransactions(
          opts,
          transactionsByaddressConfigured,
          receiptTransaction
        );

        populateTransactions = await this.populateTransaction(
          opts,
          evmBlock,
          receiptTransaction,
          filterTransactions
        );
      }
    }

    this.logger.info(
      `[${chain}][exec] Got ${
        populateTransactions?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
    return populateTransactions;
  }

  private async populateTransaction(
    opts: GetEvmOpts,
    evmBlock: EvmBlock,
    receiptTransaction: Record<string, ReceiptTransaction>,
    filterTransactions: EvmTransaction[]
  ): Promise<EvmTransaction[]> {
    filterTransactions.forEach((transaction) => {
      const redeemedTopic = opts.topics?.find((topic) => topic === REDEEMED_TOPIC);
      const logs = receiptTransaction[transaction.hash].logs;

      logs
        .filter((log) => redeemedTopic && log.topics.includes(redeemedTopic))
        .map((log) => {
          transaction.emitterChain = Number(log.topics[1]);
          transaction.emitterAddress = BigInt(log.topics[2])
            .toString(16)
            .toUpperCase()
            .padStart(64, "0");
          transaction.sequence = Number(log.topics[3]);
        });

      transaction.status = receiptTransaction[transaction.hash].status;
      transaction.timestamp = evmBlock.timestamp;
      transaction.environment = opts.environment;
      transaction.chainId = opts.chainId;
      transaction.chain = opts.chain;
      transaction.logs = logs;
    });

    return filterTransactions;
  }

  /**
   * This method filter the transactions in base your logs with the topic and address configured in the job
   * For example: Redeemed or MintAndWithdraw transactions
   */
  private filterTransactions(
    opts: GetEvmOpts,
    transactionsByaddressConfigured: EvmTransaction[],
    receiptTransaction: Record<string, ReceiptTransaction>
  ): EvmTransaction[] {
    return transactionsByaddressConfigured.filter((transaction) => {
      const optsAddresses = opts.addresses;
      const optsTopics = opts.topics;
      const logs = receiptTransaction[transaction.hash].logs;

      return logs.filter((log) => {
        optsAddresses?.includes(log.address) ||
          optsTopics?.every((topic) => log.topics?.includes(topic));
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
