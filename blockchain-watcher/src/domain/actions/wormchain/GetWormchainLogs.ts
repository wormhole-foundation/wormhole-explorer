import { WormchainRepository } from "../../repositories";
import { WormchainLog } from "../../entities/wormchain";
import winston from "winston";

export class GetWormchainLogs {
  private readonly blockRepo: WormchainRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: WormchainRepository) {
    this.logger = winston.child({ module: "GetWormchainLogs" });
    this.blockRepo = blockRepo;
  }

  async execute(range: Range, opts: GetWormchainOpts): Promise<WormchainLog[]> {
    const fromBlock = range.fromBlock;
    const toBlock = range.toBlock;

    const collectWormchainLogs: WormchainLog[] = [];

    if (fromBlock > toBlock) {
      this.logger.info(`[exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`);
      return [];
    }

    for (let blockNumber = fromBlock; blockNumber <= toBlock; blockNumber++) {
      const wormchainLogs = await this.blockRepo.getBlockLogs(blockNumber);

      if (wormchainLogs && wormchainLogs.transactions && wormchainLogs.transactions.length > 0) {
        collectWormchainLogs.push(wormchainLogs);
      }
    }

    this.logger.info(
      `[wormchain][exec] Got ${
        collectWormchainLogs?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
    return collectWormchainLogs;
  }

  private populateLog(opts: GetWormchainOpts, fromBlock: bigint, toBlock: bigint): string {
    return `[addresses:${opts.addresses}][blocks:${fromBlock} - ${toBlock}]`;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};

type TopicFilter = string | string[];

type GetWormchainOpts = {
  environment: string;
  addresses?: string[];
  topics?: TopicFilter[];
  chainId: number;
  chain: string;
};
