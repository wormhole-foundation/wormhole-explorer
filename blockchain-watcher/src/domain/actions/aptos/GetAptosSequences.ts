import { GetAptosOpts, PreviousRange, Range } from "./PollAptos";
import { AptosTransaction } from "../../entities/aptos";
import { AptosRepository } from "../../repositories";
import winston from "winston";

export class GetAptosSequences {
  protected readonly logger: winston.Logger;
  private readonly repo: AptosRepository;

  private previousFrom?: bigint;
  private lastFrom?: bigint;

  constructor(repo: AptosRepository) {
    this.logger = winston.child({ module: "GetAptosSequences" });
    this.repo = repo;
  }

  async execute(range: Range | undefined, opts: GetAptosOpts): Promise<AptosTransaction[]> {
    let populatedTransactions: AptosTransaction[] = [];

    this.logger.info(
      `[aptos][exec] Processing blocks [previousFrom: ${opts.previousFrom} - lastFrom: ${opts.lastFrom}]`
    );

    const incrementBatchIndex = 100;
    const limitBatch = opts.previousFrom
      ? Number(opts.previousFrom) - Number(opts.lastFrom) + 1
      : 100;
    let limit = limitBatch < 100 ? 1 : 100;

    while (limit <= limitBatch) {
      const fromBatch = this.lastFrom ? Number(this.lastFrom) : range?.from;

      const events = await this.repo.getEventsByEventHandle(
        {
          from: fromBatch,
          limit: limitBatch,
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

      limit += incrementBatchIndex;
    }

    this.logger.info(
      `[aptos][exec] Got ${populatedTransactions?.length} transactions to process for [addresses:${opts.addresses}][from: ${range?.from}]`
    );
    return populatedTransactions;
  }

  getBlockRange(
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
}
