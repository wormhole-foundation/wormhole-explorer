import { Block, Range, TransactionFilter } from "./PollAptos";
import { AptosTransaction } from "../../entities/aptos";
import { AptosRepository } from "../../repositories";
import { createBatches } from "../../../infrastructure/repositories/common/utils";
import winston from "winston";

export class GetAptosSequences {
  protected readonly logger: winston.Logger;
  private readonly repo: AptosRepository;

  private previousBlock?: bigint;
  private lastBlock?: bigint;

  constructor(repo: AptosRepository) {
    this.logger = winston.child({ module: "GetAptosSequences" });
    this.repo = repo;
  }

  async execute(range: Block | undefined, opts: GetAptosOpts): Promise<AptosTransaction[]> {
    let populatedTransactions: AptosTransaction[] = [];

    this.logger.info(
      `[aptos][exec] Processing blocks [previousBlock: ${opts.previousBlock} - lastBlock: ${opts.lastBlock}]`
    );

    const batches = createBatches(range);

    for (const toBatch of batches) {
      const fromBatch = this.lastBlock ? Number(this.lastBlock) : range?.fromBlock;

      const events = await this.repo.getSequenceNumber(
        {
          fromBlock: fromBatch,
          toBlock: toBatch,
        },
        opts.filter
      );

      // Update lastBlock with the new lastBlock
      this.lastBlock = BigInt(events[events.length - 1].sequence_number);

      if (opts.previousBlock == this.lastBlock) {
        return [];
      }

      // Update previousBlock with opts lastBlock
      this.previousBlock = opts.lastBlock;

      const transactions = await this.repo.getTransactionsByVersionForSourceEvent(
        events,
        opts.filter
      );

      transactions.forEach((tx) => {
        populatedTransactions.push(tx);
      });
    }

    this.logger.info(
      `[aptos][exec] Got ${populatedTransactions?.length} transactions to process for [addresses:${opts.addresses}][block: ${range?.fromBlock}]`
    );
    return populatedTransactions;
  }

  getBlockRange(
    cfgBlockBarchSize: number,
    cfgFromBlock: bigint | undefined,
    savedPreviousSequence: bigint | undefined,
    savedLastBlock: bigint | undefined
  ): Block | undefined {
    // If [set up a from block for cfg], return the fromBlock and toBlock equal the block batch size
    if (cfgFromBlock) {
      return {
        fromBlock: Number(cfgFromBlock),
        toBlock: cfgBlockBarchSize,
      };
    }

    if (savedPreviousSequence && savedLastBlock) {
      // If process the [same block], return the same lastBlock and toBlock equal the block batch size
      if (savedPreviousSequence === savedLastBlock) {
        return {
          fromBlock: Number(savedLastBlock),
          toBlock: cfgBlockBarchSize,
        };
      } else {
        // If process [different sequences], return the difference between the lastBlock and the previousBlock plus 1
        return {
          fromBlock: Number(savedLastBlock),
          toBlock: Number(savedLastBlock - savedPreviousSequence) + 1,
        };
      }
    }

    if (savedLastBlock) {
      // If there is [no previous block], return the lastBlock and toBlock equal the block batch size
      if (!cfgFromBlock || BigInt(cfgFromBlock) < savedLastBlock) {
        return {
          fromBlock: Number(savedLastBlock),
          toBlock: cfgBlockBarchSize,
        };
      }
    }
  }

  getUpdatedRange(): Range {
    return {
      previousBlock: this.previousBlock,
      lastBlock: this.lastBlock,
    };
  }
}

export type GetAptosOpts = {
  addresses: string[];
  filter: TransactionFilter;
  previousBlock?: bigint | undefined;
  lastBlock?: bigint | undefined;
};
