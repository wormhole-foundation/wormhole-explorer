import { DefaultTransactions } from "./strategy/DefaultTransactions";
import { EvmBlockRepository } from "../../repositories";
import { NFTTransactions } from "./strategy/NFTTransactions";
import { EvmTransaction } from "../../entities";
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

    this.logger.info(
      `[${chain}][exec] Processing blocks [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );

    let populatedTransactions: EvmTransaction[] = [];

    const processors = [
      new DefaultTransactions(this.blockRepo, fromBlock, toBlock, chain, opts),
      new NFTTransactions(this.blockRepo, fromBlock, toBlock, chain, opts),
    ];

    for (const filter of opts.filters!) {
      for (const process of processors) {
        const normalizeFilter = this.normalizeFilter(filter);

        if (process.apply(normalizeFilter.topics)) {
          const transaction = await process.execute(normalizeFilter);
          populatedTransactions.push(...transaction);
        }
      }
    }

    this.logger.info(
      `[${chain}][exec] Got ${populatedTransactions?.length} transactions to process [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );

    return populatedTransactions;
  }

  private normalizeFilter(filter: Filter): Filter {
    return {
      addresses: filter.addresses.map((address) => address.toLowerCase()),
      topics: filter.topics.map((topic) => topic.toLowerCase()),
    };
  }
}

export interface GetTransactions {
  apply(topic: string[]): boolean;
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
