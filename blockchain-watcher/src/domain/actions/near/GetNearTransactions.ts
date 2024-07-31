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
    const contract = opts.contracts[0];
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

    const nearTransactions = await this.blockRepo.getTransactions(contract, fromBlock, toBlock);

    this.logger.info(
      `[${chain}][exec] Got ${
        nearTransactions?.length
      } transactions to process for ${this.populateLog(contract, fromBlock, toBlock)}`
    );
    return nearTransactions;
  }

  private populateLog(
    addresses: string,
    previousFrom: bigint | undefined,
    lastFrom: bigint
  ): string {
    return `[contract:${addresses}][previousFrom: ${previousFrom} - lastFrom: ${lastFrom}]`;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
