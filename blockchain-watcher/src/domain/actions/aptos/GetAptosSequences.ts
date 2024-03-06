import { TransactionsByVersion } from "../../../infrastructure/repositories/aptos/AptosJsonRPCBlockRepository";
import { AptosRepository } from "../../repositories";
import winston from "winston";
import { Filter, Sequence } from "./PollAptos";

export class GetAptosSequences {
  private readonly repo: AptosRepository;
  protected readonly logger: winston.Logger;

  private lastSequence?: bigint;
  private previousSequence?: bigint;

  constructor(repo: AptosRepository) {
    this.logger = winston.child({ module: "GetAptosSequences" });
    this.repo = repo;
  }

  async execute(range: Sequence | undefined, opts: GetAptosOpts): Promise<TransactionsByVersion[]> {
    let populatedTransactions: TransactionsByVersion[] = [];

    const batches = this.createBatches(range);

    for (const batch of batches) {
      const events = await this.repo.getSequenceNumber(
        {
          fromSequence: range?.fromSequence,
          toSequence: batch,
        },
        opts.filter
      );

      // update last sequence with the new sequence
      this.lastSequence = BigInt(events[events.length - 1].sequence_number);

      if (opts.previousSequence == this.lastSequence) {
        return [];
      }

      // save previous sequence with last sequence
      this.previousSequence = opts.lastSequence;

      const transactions = await this.repo.getTransactionsForVersions(events, opts.filter);
      transactions.forEach((tx) => {
        populatedTransactions.push(tx);
      });
    }

    this.logger.info(
      `[aptos][exec] Got ${populatedTransactions?.length} transactions to process for [addresses:${opts.addresses}][sequence: ${range?.fromSequence}]`
    );
    return populatedTransactions;
  }

  getBlockRange(
    cfgBlockBarchSize: number,
    cfgFromSequence: bigint | undefined,
    savedPreviousSequence: bigint | undefined,
    savedLastSequence: bigint | undefined
  ): Sequence | undefined {
    // if [set up a from sequence for cfg], return the from sequence and the to sequence equal the block batch size
    if (cfgFromSequence) {
      return {
        fromSequence: Number(cfgFromSequence),
        toSequence: cfgBlockBarchSize,
      };
    }

    if (savedPreviousSequence && savedLastSequence) {
      // if process the [same sequence], return the same last sequence and the to sequence equal the block batch size
      if (savedPreviousSequence === savedLastSequence) {
        return {
          fromSequence: Number(savedLastSequence),
          toSequence: cfgBlockBarchSize,
        };
      }

      // if process [different sequences], return the difference between the last sequence and the previous sequence plus 1
      if (savedPreviousSequence !== savedLastSequence) {
        return {
          fromSequence: Number(savedLastSequence),
          toSequence: Number(savedLastSequence) - Number(savedPreviousSequence) + 1,
        };
      }
    }

    if (savedLastSequence) {
      // if there is [no previous sequence], return the last sequence and the to sequence equal the block batch size
      if (!cfgFromSequence || BigInt(cfgFromSequence) < savedLastSequence) {
        return {
          fromSequence: Number(savedLastSequence),
          toSequence: cfgBlockBarchSize,
        };
      }
    }
  }

  updatedRange() {
    return {
      previousSequence: this.previousSequence,
      lastSequence: this.lastSequence,
    };
  }

  private createBatches(range: Sequence | undefined): number[] {
    let batchSize = 100;
    let total = 1;

    if (range && range.toSequence) {
      batchSize = range.toSequence < batchSize ? range.toSequence : batchSize;
      total = range.toSequence ?? total;
    }

    const numBatches = Math.ceil(total / batchSize);
    const batches: number[] = [];

    for (let i = 0; i < numBatches; i++) {
      batches.push(batchSize);
    }

    return batches;
  }
}

export type GetAptosOpts = {
  addresses: string[];
  filter: Filter;
  previousSequence?: bigint | undefined;
  lastSequence?: bigint | undefined;
};
