import { NearTransaction } from "../../entities/near";
import { NearRepository } from "../../repositories";
import { GetNearOpts } from "./PollNear";
import winston from "winston";

export class GetNearTransactions {
  private readonly blockRepo: NearRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: NearRepository) {
    this.logger = winston.child({ module: "GetNearTransactions" });
    this.blockRepo = blockRepo;
  }

  async execute(range: Range, opts: GetNearOpts): Promise<NearTransaction[]> {
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

    return [];
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
