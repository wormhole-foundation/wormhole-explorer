import { GetCosmosOpts, PreviousRange } from "./PollCosmos";
import { CosmosRepository } from "../../repositories";
import { CosmosRedeem } from "../../entities/wormchain";
import winston from "winston";

export class GetCosmosRedeems {
  private readonly blockRepo: CosmosRepository;
  protected readonly logger: winston.Logger;

  private previousFrom?: bigint;
  private lastFrom?: bigint;

  constructor(blockRepo: CosmosRepository) {
    this.logger = winston.child({ module: "GetCosmosRedeems" });
    this.blockRepo = blockRepo;
  }

  async execute(opts: GetCosmosOpts): Promise<CosmosRedeem[]> {
    const { chainId, filter, blockBatchSize, previousFrom, chain } = opts;
    this.logger.info(
      `[${chain}][exec] Processing range [previousFrom: ${opts.previousFrom} - lastFrom: ${opts.lastFrom}]`
    );

    const cosmosRedeems = await this.blockRepo.getRedeems(chainId, filter, blockBatchSize, chain);
    if (cosmosRedeems.length === 0) {
      return [];
    }

    const newLastFrom = BigInt(cosmosRedeems[cosmosRedeems.length - 1].height);
    if (previousFrom == newLastFrom) {
      return [];
    }

    const filteredCosmosRedeems =
      previousFrom && newLastFrom
        ? cosmosRedeems.filter(
            (cosmosRedeem) =>
              cosmosRedeem.height >= previousFrom && cosmosRedeem.height <= newLastFrom
          )
        : cosmosRedeems;

    await Promise.all(
      filteredCosmosRedeems.map(async (cosmosRedeem) => {
        const timestamp = await this.blockRepo.getBlockTimestamp(
          cosmosRedeem.height,
          chainId,
          chain
        );
        cosmosRedeem.timestamp = String(timestamp);
      })
    );

    // Update previousFrom and lastFrom with opts lastFrom
    this.previousFrom = BigInt(cosmosRedeems[cosmosRedeems.length - 1].height);
    this.lastFrom = newLastFrom;

    this.logger.info(
      `[${chain}][exec] Got ${
        filteredCosmosRedeems?.length
      } transactions to process for ${this.populateLog(
        filter.addresses,
        this.previousFrom,
        this.lastFrom
      )}`
    );
    return filteredCosmosRedeems;
  }

  getUpdatedRange(): PreviousRange {
    return {
      previousFrom: this.previousFrom,
      lastFrom: this.lastFrom,
    };
  }

  private populateLog(
    addresses: string[],
    previousFrom: bigint | undefined,
    lastFrom: bigint
  ): string {
    return `[addresses:${addresses}][previousFrom: ${previousFrom} - lastFrom: ${lastFrom}]`;
  }
}
