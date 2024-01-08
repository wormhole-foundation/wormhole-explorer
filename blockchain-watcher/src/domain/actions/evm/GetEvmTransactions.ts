import { EvmBlock, EvmTransaction } from "../../entities";
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

    let populateTransactions: EvmTransaction[] = [];
    const isTransactionsPresent = true;
    const chain = opts.chain;

    for (let block = fromBlock; block <= toBlock; block++) {
      const evmBlock = await this.blockRepo.getBlock(chain, block, isTransactionsPresent);
      const transactions = evmBlock.transactions ?? [];

      // Only process transactions to the contract address
      const transactionsFilter = transactions.filter(
        (transaction) =>
          opts.addresses?.includes(String(transaction.to).toLowerCase()) ||
          opts.addresses?.includes(String(transaction.from).toLowerCase())
      );

      if (transactionsFilter.length > 0) {
        populateTransactions = await this.populateTransaction(opts, evmBlock, transactionsFilter);
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
    transactionsFilter: EvmTransaction[]
  ): Promise<EvmTransaction[]> {
    const chain = opts.chain;
    const hashNumbers = new Set(transactionsFilter.map((transaction) => transaction.hash));
    const receiptTransaction = await this.blockRepo.getTransactionReceipt(chain, hashNumbers);

    transactionsFilter.forEach((transaction) => {
      transaction.chainId = opts.chainId;
      transaction.timestamp = evmBlock.timestamp;
      transaction.status = receiptTransaction[transaction.hash].status;
      transaction.environment = opts.environment;
      transaction.chain = chain;
    });

    return transactionsFilter;
  }

  private populateLog(opts: GetEvmOpts, fromBlock: bigint, toBlock: bigint): string {
    return `[addresses:${opts.addresses}][topics:${opts.topics}][blocks:${fromBlock} - ${toBlock}]`;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
