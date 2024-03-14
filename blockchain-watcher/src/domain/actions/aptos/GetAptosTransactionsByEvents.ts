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
    let populatedTransactions: AptosTransaction[] = [];

    this.logger.info(
      `[aptos][exec] Processing range [previousFrom: ${opts.previousFrom} - lastFrom: ${opts.lastFrom}]`
    );

    const events = await this.repo.getEventsByEventHandle(
      {
        from: range?.from,
        limit: range?.limit,
      },
      opts.filter
    );

    const newLastFrom = BigInt(events[events.length - 1].sequence_number);

    if (opts.previousFrom == newLastFrom) {
      return [];
    }

    const transactions = await this.repo.getTransactionsByVersion(events, opts.filter);

    transactions.forEach((tx) => {
      populatedTransactions.push(tx);
    });

    this.logger.info(
      `[aptos][exec] Got ${populatedTransactions?.length} transactions to process for [addresses:${opts.addresses}][from: ${range?.from} - limit: ${range?.limit}]`
    );

    // Update lastFrom and previousFrom with opts lastFrom
    this.lastFrom = BigInt(events[events.length - 1].sequence_number);
    this.previousFrom = opts.lastFrom;

    return populatedTransactions;
  }

  getRange(
    cfgLimitBatchSize: number,
    cfgFrom: bigint | undefined,
    savedPreviousSequence: bigint | undefined,
    savedlastFrom: bigint | undefined
  ): Range | undefined {
    // If [set up a from for cfg], return the from and limit equal the from batch size
    if (cfgFrom) {
      return {
        from: Number(cfgFrom),
        limit: cfgLimitBatchSize,
      };
    }

    if (savedPreviousSequence && savedlastFrom) {
      // If process the [same from], return the same lastFrom and limit equal the from batch size
      if (savedPreviousSequence === savedlastFrom) {
        return {
          from: Number(savedlastFrom),
          limit: cfgLimitBatchSize,
        };
      } else {
        // If process [different sequences], return the difference between the lastFrom and the previousFrom plus 1
        return {
          from: Number(savedlastFrom),
          limit: Number(savedlastFrom) - Number(savedPreviousSequence) + 1,
        };
      }
    }

    if (savedlastFrom) {
      // If there is [no previous from], return the lastFrom and limit equal the from batch size
      if (!cfgFrom || BigInt(cfgFrom) < savedlastFrom) {
        return {
          from: Number(savedlastFrom),
          limit: cfgLimitBatchSize,
        };
      }
    }
  }

  getUpdatedRange(): PreviousRange {
    return {
      previousFrom: this.previousFrom,
      lastFrom: this.lastFrom,
    };
  }
}
