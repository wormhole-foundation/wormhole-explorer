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
      `[aptos][exec] Processing blocks [previousFrom: ${opts.previousFrom} - lastFrom: ${opts.lastFrom}]`
    );

    const from = this.lastFrom ? Number(this.lastFrom) : range?.from;

    const events = await this.repo.getEventsByEventHandle(
      {
        from: from,
        limit: range?.limit,
      },
      opts.filter
    );

    // Update lastFrom with the new lastFrom
    this.lastFrom = BigInt(events[events.length - 1].sequence_number);

    if (opts.previousFrom == this.lastFrom) {
      return [];
    }

    // Update previousFrom with opts lastFrom
    this.previousFrom = opts.lastFrom;

    const transactions = await this.repo.getTransactionsByVersion(events, opts.filter);

    transactions.forEach((tx) => {
      populatedTransactions.push(tx);
    });

    this.logger.info(
      `[aptos][exec] Got ${populatedTransactions?.length} transactions to process for [addresses:${opts.addresses}][from: ${range?.from}]`
    );
    return populatedTransactions;
  }

  getRange(
    cfgBlockBarchSize: number,
    cfgFrom: bigint | undefined,
    savedPreviousSequence: bigint | undefined,
    savedlastFrom: bigint | undefined
  ): Range | undefined {
    // If [set up a from for cfg], return the from and limit equal the from batch size
    if (cfgFrom) {
      return {
        from: Number(cfgFrom),
        limit: cfgBlockBarchSize,
      };
    }

    if (savedPreviousSequence && savedlastFrom) {
      // If process the [same from], return the same lastFrom and limit equal the from batch size
      if (savedPreviousSequence === savedlastFrom) {
        return {
          from: Number(savedlastFrom),
          limit: cfgBlockBarchSize,
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
          limit: cfgBlockBarchSize,
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

  private createBatch(opts: GetAptosOpts) {
    const batchSize = 100;
    const totalBatchLimit =
      opts.previousFrom && opts.lastFrom
        ? Number(opts.lastFrom) - Number(opts.previousFrom) + 1
        : batchSize;
    let limitBatch = totalBatchLimit < batchSize ? 1 : batchSize;

    return {
      batchSize,
      totalBatchLimit,
      limitBatch,
    };
  }
}
