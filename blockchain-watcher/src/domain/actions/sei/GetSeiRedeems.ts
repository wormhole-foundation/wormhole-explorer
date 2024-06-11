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

    const seiRedeems = await this.blockRepo.getRedeems(
      opts.chainId,
      opts.addresses[0],
      opts.blockBatchSize
    );

    const newLastFrom = BigInt(seiRedeems[seiRedeems.length - 1].height);
    if (seiRedeems.length === 0 || opts.previousFrom === newLastFrom) {
      return [];
    }

    const filteredSeiRedeems =
      opts.previousFrom && newLastFrom
        ? seiRedeems.filter(
            (seiRedeem) => seiRedeem.height >= opts.previousFrom! && seiRedeem.height <= newLastFrom
          )
        : seiRedeems;

    await Promise.all(
      filteredSeiRedeems.map(async (seiRedeem) => {
        const timestamp = await this.blockRepo.getBlockTimestamp(BigInt(seiRedeem.height));
        seiRedeem.timestamp = timestamp;
      })
    );

    // Update previousFrom and lastFrom with opts lastFrom
    this.previousFrom = BigInt(seiRedeems[seiRedeems.length - 1].height);
    this.lastFrom = newLastFrom;

    this.logger.info(
      `[sei][exec] Got ${filteredSeiRedeems?.length} transactions to process for ${this.populateLog(
        opts,
        this.previousFrom,
        this.lastFrom
      )}`
    );
    return filteredSeiRedeems;
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
