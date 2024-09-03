import { GetAptosOpts, PreviousRange, Range } from "./PollAptos";
import { AptosTransaction } from "../../entities/aptos";
import { AptosRepository } from "../../repositories";
import winston from "winston";

export class GetAptosTransactionsByEvents {
  protected readonly logger: winston.Logger;
  private readonly repo: AptosRepository;

  private previousFrom?: bigint;
  private lastFrom?: bigint;

  constructor(repo: AptosRepository) {
    this.logger = winston.child({ module: "GetAptosTransactionsByEvents" });
    this.repo = repo;
  }

  async execute(range: Range | undefined, opts: GetAptosOpts): Promise<AptosTransaction[]> {
    const { previousFrom, lastFrom, cfgFrom, filters, addresses } = opts;
    let populatedTransactions: AptosTransaction[] = [];

    this.logger.info(
      `[aptos][exec] Processing range [previousFrom: ${previousFrom} - lastFrom: ${lastFrom}]`
    );

    const events = await this.repo.getEventsByEventHandle(
      {
        from: range?.from,
        limit: range?.limit,
      },
      filters[0] // It use the first filter because only process the Core Contract source events
    );

    // Validate if the transactions are the same or there is a delay
    const newLastFrom = BigInt(events[events.length - 1].sequence_number);
    if (!cfgFrom && lastFrom && lastFrom >= newLastFrom) {
      this.logger.warn(
        `[aptos][exec] Processing the same block or encountering a delay [lastFrom: ${lastFrom} - newLastFrom: ${newLastFrom}]`
      );
      return [];
    }

    const transactions = await this.repo.getTransactionsByVersion(events);
    transactions.forEach((tx) => {
      populatedTransactions.push(tx);
    });

    this.logger.info(
      `[aptos][exec] Got ${populatedTransactions?.length} transactions to process for [addresses:${addresses}][from: ${range?.from} - limit: ${range?.limit}]`
    );

    this.lastFrom = newLastFrom; // Update lastFrom
    this.previousFrom = lastFrom; // Update previousFrom
    return populatedTransactions;
  }

  getRange(
    cfgLimitBatchSize: number,
    cfgFrom: bigint | undefined,
    savedPreviousFrom: bigint | undefined,
    savedlastFrom: bigint | undefined
  ): Range | undefined {
    // If [set up a from for cfg], return the from and limit equal the from batch size
    if (cfgFrom) {
      return {
        from: Number(cfgFrom),
        limit: cfgLimitBatchSize,
      };
    }

    if (savedPreviousFrom && savedlastFrom) {
      // If process the [same from], return the same lastFrom and limit equal the from batch size
      if (savedPreviousFrom === savedlastFrom) {
        return {
          from: Number(savedlastFrom),
          limit: cfgLimitBatchSize,
        };
      } else {
        // If process [different sequences], return the difference between the lastFrom and the previousFrom plus 1
        return {
          from: Number(savedlastFrom),
          limit: Number(savedlastFrom) - Number(savedPreviousFrom) + 1,
        };
      }
    }

    if (savedlastFrom) {
      // If there is [no previous from], return the lastFrom and limit equal the from batch size
      return {
        from: Number(savedlastFrom),
        limit: cfgLimitBatchSize,
      };
    }
  }

  getUpdatedRange(): PreviousRange {
    return {
      previousFrom: this.previousFrom,
      lastFrom: this.lastFrom,
    };
  }
}
