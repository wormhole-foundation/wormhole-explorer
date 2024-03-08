import { TransactionsByVersion } from "../../../infrastructure/repositories/aptos/AptosJsonRPCBlockRepository";
import { AptosRepository } from "../../repositories";
import winston from "winston";
import { Block, TransactionFilter } from "./PollAptos";

export class GetAptosTransactions {
  private readonly repo: AptosRepository;
  protected readonly logger: winston.Logger;

  private lastBlock?: bigint;
  private previousBlock?: bigint;

  constructor(repo: AptosRepository) {
    this.logger = winston.child({ module: "GetAptosTransactions" });
    this.repo = repo;
  }

  async execute(range: Block | undefined, opts: GetAptosOpts): Promise<TransactionsByVersion[]> {
    let populatedTransactions: TransactionsByVersion[] = [];

    this.logger.info(
      `[aptos][exec] Processing blocks [previousBlock: ${opts.previousBlock} - latestBlock: ${opts.lastBlock}]`
    );

    const batches = this.createBatches(range);

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

      // update last block with the new block
      this.lastBlock = BigInt(transaction[transaction.length - 1].version);

      if (opts.previousBlock == this.lastBlock) {
        return [];
      }

      // save previous block with last block
      this.previousBlock = opts.lastBlock;

      if (transactionsByAddressConfigured.length > 0) {
        const transactions = await this.repo.getTransactionsByVersionsForRedeemedEvent(
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
    // if [set up a from block for cfg], return the from block and the to block equal the block batch size
    if (cfgFromBlock) {
      return {
        fromBlock: Number(cfgFromBlock),
        toBlock: cfgBlockBarchSize,
      };
    }

    if (savedPreviousBlock && savedLastBlock) {
      // if process [equal or different blocks], return the same last block and the to block equal the block batch size
      if (savedPreviousBlock === savedLastBlock || savedPreviousBlock !== savedLastBlock) {
        return {
          fromBlock: Number(savedLastBlock),
          toBlock: cfgBlockBarchSize,
        };
      }
    }

    if (savedLastBlock) {
      // if there is [no previous block], return the last block and the to block equal the block batch size
      if (!cfgFromBlock || BigInt(cfgFromBlock) < savedLastBlock) {
        return {
          fromBlock: Number(savedLastBlock),
          toBlock: cfgBlockBarchSize,
        };
      }
    }
  }

  getUpdatedRange() {
    return {
      previousBlock: this.previousBlock,
      lastBlock: this.lastBlock,
    };
  }

  private createBatches(range: Block | undefined): number[] {
    let batchSize = 100;
    let total = 1;

    if (range && range.toBlock) {
      batchSize = range.toBlock < batchSize ? range.toBlock : batchSize;
      total = range.toBlock ?? total;
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
  filter: TransactionFilter;
  previousBlock?: bigint | undefined;
  lastBlock?: bigint | undefined;
};
