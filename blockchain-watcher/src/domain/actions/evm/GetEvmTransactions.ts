import { EvmBlock, EvmTransaction, ReceiptTransaction } from "../../entities";
import { GetTransactionsByLogFiltersStrategy } from "./strategy/GetTransactionsByLogFiltersStrategy";
import { GetTransactionsByBlocksStrategy } from "./strategy/GetTransactionsByBlocksStrategy";
import { EvmBlockRepository } from "../../repositories";
import { GetEvmOpts } from "./PollEvm";
import winston from "winston";

export class GetEvmTransactions {
  private readonly blockRepo: EvmBlockRepository;
  protected readonly logger: winston.Logger;
  private strategies: GetTransactions[] = [];

  constructor(blockRepo: EvmBlockRepository) {
    this.logger = winston.child({ module: "GetEvmTransactions" });
    this.blockRepo = blockRepo;
    this.strategies = [
      new GetTransactionsByLogFiltersStrategy(this.blockRepo),
      new GetTransactionsByBlocksStrategy(this.blockRepo),
    ];
  }

  async execute(range: Range, opts: GetEvmOpts): Promise<EvmTransaction[]> {
    const { fromBlock, toBlock } = range;
    const { chain, filters } = opts;

    if (fromBlock > toBlock) {
      this.logger.info(
        `[${chain}][exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
      );
      return [];
    }

    this.logger.info(
      `[${chain}][exec] Processing blocks [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );

    let populatedTransactions: EvmTransaction[] = [];

    await Promise.all(
      filters.map(async (filter) => {
        await Promise.all(
          this.strategies.map(async (strategy) => {
            if (strategy.appliesTo(filter.addresses, filter.topics)) {
              const result = await strategy.execute(filter, fromBlock, toBlock, opts);
              populatedTransactions.push(...result);
            }
          })
        );
      })
    );

    this.logger.info(
      `[${chain}][exec] Got ${populatedTransactions?.length} transactions to process [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );

    return populatedTransactions;
  }
}

export function populateTransaction(
  opts: GetEvmOpts,
  evmBlocks: Record<string, EvmBlock>,
  transactionReceipts: Record<string, ReceiptTransaction>,
  filterTransactions: EvmTransaction[],
  populatedTransactions: EvmTransaction[],
  topics: string[]
) {
  filterTransactions.forEach((transaction) => {
    const txReceipt = transactionReceipts[transaction.hash];
    if (txReceipt) {
      transaction.effectiveGasPrice = txReceipt.effectiveGasPrice;
      transaction.gasUsed = txReceipt.gasUsed;
      transaction.timestamp = evmBlocks[transaction.blockHash].timestamp;
      transaction.status = txReceipt.status;
      transaction.logs = txReceipt.logs.filter((log) =>
        log.topics.some((topic) => topics.includes(topic))
      );
      transaction.environment = opts.environment;
      transaction.chainId = opts.chainId;
      transaction.chain = opts.chain;
      populatedTransactions.push(transaction);
    }
  });
}

// Interface for strategy pattern
export interface GetTransactions {
  appliesTo(addresses: string[], topics: string[]): boolean;
  execute(
    filter: Filter,
    fromBlock: bigint,
    toBlock: bigint,
    opts: GetEvmOpts
  ): Promise<EvmTransaction[]>;
}

export type Filter = {
  addresses: string[];
  topics: string[];
};

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
