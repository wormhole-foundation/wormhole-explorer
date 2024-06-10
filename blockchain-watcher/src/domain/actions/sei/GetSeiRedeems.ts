import { GetSeiOpts, PreviousRange } from "./PollSei";
import { SeiRepository } from "../../repositories";
import { SeiRedeem } from "../../entities/sei";
import winston from "winston";

export class GetSeiRedeems {
  private readonly blockRepo: SeiRepository;
  protected readonly logger: winston.Logger;

  private previousFrom?: bigint;
  private lastFrom?: bigint;

  constructor(blockRepo: SeiRepository) {
    this.logger = winston.child({ module: "GetSeiRedeems" });
    this.blockRepo = blockRepo;
  }

  async execute(opts: GetSeiOpts): Promise<SeiRedeem[]> {
    this.logger.info(
      `[sei][exec] Processing range [previousFrom: ${opts.previousFrom} - lastFrom: ${opts.lastFrom}]`
    );

    const seiRedeems = await this.blockRepo.getRedeems(opts.chainId, opts.addresses[0]);

    const newLastFrom = BigInt(seiRedeems[seiRedeems.length - 1].height);
    if (opts.previousFrom == newLastFrom) {
      return [];
    }

    if (seiRedeems && seiRedeems.length >= 0) {
      await Promise.all(
        seiRedeems.map(async (seiRedeem) => {
          const timestamp = await this.blockRepo.getBlockTimestamp(BigInt(seiRedeem.height));
          seiRedeem.timestamp = timestamp;
        })
      );
    }

    // Update previousFrom and lastFrom with opts lastFrom
    this.previousFrom = opts.lastFrom ?? newLastFrom; // If saved lastFrom is undefined, use newLastFrom because it's the first time
    this.lastFrom = newLastFrom;

    this.logger.info(
      `[sei][exec] Got ${seiRedeems?.length} transactions to process for ${this.populateLog(
        opts,
        this.previousFrom,
        this.lastFrom
      )}`
    );
    return seiRedeems;
  }

  getUpdatedRange(): PreviousRange {
    return {
      previousFrom: this.previousFrom,
      lastFrom: this.lastFrom,
    };
  }

  private populateLog(
    opts: { addresses: string[] },
    previousFrom: bigint | undefined,
    lastFrom: bigint
  ): string {
    return `[addresses:${opts.addresses}][range:${previousFrom} - ${lastFrom}]`;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
