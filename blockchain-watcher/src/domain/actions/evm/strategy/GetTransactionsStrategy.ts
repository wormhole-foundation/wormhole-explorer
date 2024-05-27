import { DefaultTransactions } from "./DefaultTransactions";
import { EvmBlockRepository } from "../../../repositories";
import { NFTTransactions } from "./NFTTransactions";
import { EvmTransaction } from "../../../entities";
import { GetEvmOpts } from "../PollEvm";

export class GetTransactionsStrategy {
  private readonly blockRepo: EvmBlockRepository;
  private readonly fromBlock: bigint;
  private readonly toBlock: bigint;
  private readonly chain: string;
  private readonly opts: GetEvmOpts;

  constructor(
    blockRepo: EvmBlockRepository,
    fromBlock: bigint,
    toBlock: bigint,
    chain: string,
    opts: GetEvmOpts
  ) {
    this.blockRepo = blockRepo;
    this.fromBlock = fromBlock;
    this.toBlock = toBlock;
    this.chain = chain;
    this.opts = opts;
  }

  async execute(): Promise<EvmTransaction[]> {
    let populatedTransactions: EvmTransaction[] = [];

    const processors = [
      new DefaultTransactions(this.blockRepo, this.fromBlock, this.toBlock, this.chain, this.opts),
      new NFTTransactions(this.blockRepo, this.fromBlock, this.toBlock, this.chain, this.opts),
    ];

    for (const filter of this.opts.filters!) {
      for (const process of processors) {
        const normalizeFilter = this.normalizeFilter(filter);

        if (process.apply(normalizeFilter.topics)) {
          const transaction = await process.execute(normalizeFilter);
          populatedTransactions.push(...transaction);
        }
      }
    }
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
