import { GetCosmosOpts, PreviousRange } from "./PollCosmos";
import { CosmosTransaction } from "../../entities/cosmos";
import { CosmosRepository } from "../../repositories";
import winston from "winston";

export class GetCosmosTransactions {
  private readonly blockRepo: CosmosRepository;
  protected readonly logger: winston.Logger;

  private previousFrom?: bigint;
  private lastFrom?: bigint;

  constructor(blockRepo: CosmosRepository) {
    this.logger = winston.child({ module: "GetCosmosTransactions" });
    this.blockRepo = blockRepo;
  }

  async execute(opts: GetCosmosOpts): Promise<CosmosTransaction[]> {
    const { chainId, filter, blockBatchSize, previousFrom, chain } = opts;
    this.logger.info(
      `[${chain}][exec] Processing range [previousFrom: ${opts.previousFrom} - lastFrom: ${opts.lastFrom}]`
    );

    const cosmosTransactions = await this.blockRepo.getTransactions(filter, blockBatchSize, chain);

    if (cosmosTransactions.length === 0) {
      return [];
    }

    const newLastFrom = BigInt(cosmosTransactions[cosmosTransactions.length - 1].height);
    if (previousFrom && (previousFrom == newLastFrom || previousFrom > newLastFrom)) {
      return [];
    }

    const filteredCosmosTransactions =
      previousFrom && newLastFrom
        ? cosmosTransactions.filter(
            (cosmosTransaction) =>
              BigInt(cosmosTransaction.height) >= previousFrom &&
              BigInt(cosmosTransaction.height) <= newLastFrom
          )
        : cosmosTransactions;

    for (const cosmosTransaction of filteredCosmosTransactions) {
      const timestamp = await this.blockRepo.getBlockTimestamp(cosmosTransaction.height, chain);
      // Populate the transaction with the timestamp and chainId
      cosmosTransaction.timestamp = timestamp;
      cosmosTransaction.chainId = chainId;
    }

    // Update previousFrom and lastFrom with opts lastFrom
    this.previousFrom = BigInt(cosmosTransactions[cosmosTransactions.length - 1].height);
    this.lastFrom = newLastFrom;

    this.logger.info(
      `[${chain}][exec] Got ${
        filteredCosmosTransactions?.length
      } transactions to process for ${this.populateLog(
        filter.addresses,
        opts.previousFrom,
        this.lastFrom
      )}`
    );
    return filteredCosmosTransactions;
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
