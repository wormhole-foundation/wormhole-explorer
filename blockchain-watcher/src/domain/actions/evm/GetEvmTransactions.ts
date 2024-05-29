import { EvmBlock, EvmTransaction, ReceiptTransaction } from "../../entities";
import { EvmBlockRepository } from "../../repositories";
import { DefaultProcess } from "./strategy/DefaultProcess";
import { NFTProcess } from "./strategy/NFTProcess";
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
    const { fromBlock, toBlock } = range;
    const chain = opts.chain;

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

    const processes = [
      new DefaultProcess(this.blockRepo, fromBlock, toBlock, chain, opts),
      new NFTProcess(this.blockRepo, fromBlock, toBlock, chain, opts),
    ];

    await Promise.all(
      opts.filters.map(async (filter) => {
        await Promise.all(
          processes.map(async (process) => {
            if (process.apply(filter.topics)) {
              const result = await process.execute(filter);
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

// Interface for strategy pattern
export interface GetTransactions {
  apply(topics: string[]): boolean;
  execute(filter: Filter): Promise<EvmTransaction[]>;
}

export type Filter = {
  addresses: string[];
  topics: string[];
};

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
