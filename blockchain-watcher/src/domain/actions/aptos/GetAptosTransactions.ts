import { Block, Range, TransactionFilter } from "./PollAptos";
import { AptosTransaction } from "../../entities/aptos";
import { AptosRepository } from "../../repositories";
import { createBatches } from "../../../infrastructure/repositories/common/utils";
import winston from "winston";

export class GetAptosTransactions {
  protected readonly logger: winston.Logger;
  private readonly repo: AptosRepository;

  private previousBlock?: bigint;
  private lastBlock?: bigint;

  constructor(repo: AptosRepository) {
    this.logger = winston.child({ module: "GetAptosTransactions" });
    this.repo = repo;
  }

  async execute(range: Block | undefined, opts: GetAptosOpts): Promise<AptosTransaction[]> {
    let populatedTransactions: AptosTransaction[] = [];

    this.logger.info(
      `[aptos][exec] Processing blocks [previousBlock: ${opts.previousBlock} - latestBlock: ${opts.lastBlock}]`
    );

    const batches = createBatches(range);

    for (const toBatch of batches) {
      const fromBatch = this.lastBlock ? Number(this.lastBlock) : range?.fromBlock;

      const transaction = await this.repo.getTransactions({
        fromBlock: fromBatch,
        toBlock: toBatch,
      });

      // Only process transactions to the contract address configured
      const transactionsByAddressConfigured = transaction.filter((transaction) =>
        opts.filter?.type?.includes(String(transaction.payload?.function).toLowerCase())
      );

      // Update lastBlock with the new lastBlock
      this.lastBlock = BigInt(transaction[transaction.length - 1].version);

      if (opts.previousBlock == this.lastBlock) {
        return [];
      }

      // Update previousBlock with opts lastBlock
      this.previousBlock = opts.lastBlock;

      if (transactionsByAddressConfigured.length > 0) {
        const transactions = await this.repo.getTransactionsByVersionForRedeemedEvent(
          transactionsByAddressConfigured,
          opts.filter
        );

        transactions.forEach((tx) => {
          populatedTransactions.push(tx);
        });
      }
    }

    return populatedTransactions;
  }

  getBlockRange(
    cfgBlockBarchSize: number,
    cfgFromBlock: bigint | undefined,
    savedPreviousBlock: bigint | undefined,
    savedLastBlock: bigint | undefined
  ): Block | undefined {
    // If [set up a from block for cfg], return the fromBlock and toBlock equal the block batch size
    if (cfgFromBlock) {
      return {
        fromBlock: Number(cfgFromBlock),
        toBlock: cfgBlockBarchSize,
      };
    }

    if (savedPreviousBlock && savedLastBlock) {
      // If process [equal or different blocks], return the same lastBlock and toBlock equal the block batch size
      if (savedPreviousBlock === savedLastBlock) {
        return {
          fromBlock: Number(savedLastBlock),
          toBlock: cfgBlockBarchSize,
        };
      } else {
        // If process [different sequences], return the difference between the lastBlock and the previousBlock plus 1
        return {
          fromBlock: Number(savedLastBlock),
          toBlock: Number(savedLastBlock - savedPreviousBlock) + 1,
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
