import { Range, PreviousRange, GetAptosOpts } from "./PollAptos";
import { AptosTransaction } from "../../entities/aptos";
import { AptosRepository } from "../../repositories";
import winston from "winston";

export class GetAptosTransactions {
  protected readonly logger: winston.Logger;
  private readonly repo: AptosRepository;

  private previousFrom?: bigint;
  private lastFrom?: bigint;

  constructor(repo: AptosRepository) {
    this.logger = winston.child({ module: "GetAptosTransactions" });
    this.repo = repo;
  }

  async execute(range: Range | undefined, opts: GetAptosOpts): Promise<AptosTransaction[]> {
    let populatedTransactions: AptosTransaction[] = [];

    this.logger.info(
      `[aptos][exec] Processing blocks [previousFrom: ${opts.previousFrom} - lastFrom: ${opts.lastFrom}]`
    );

    const from = this.lastFrom ? Number(this.lastFrom) : range?.from;

    const transactions = await this.repo.getTransactions({
      from: from,
      limit: range?.limit,
    });

    // Only process transactions to the contract address configured
    const transactionsByAddressConfigured = transactions.filter((transaction) =>
      opts.filter?.type?.includes(String(transaction.payload?.function).toLowerCase())
    );

    // Update lastFrom with the new lastFrom
    this.lastFrom = BigInt(transactions[transactions.length - 1].version!);

    if (opts.previousFrom == this.lastFrom) {
      return [];
    }

    // Update previousFrom with opts lastFrom
    this.previousFrom = opts.lastFrom;

    if (transactionsByAddressConfigured.length > 0) {
      const transactions = await this.repo.getTransactionsByVersion(
        transactionsByAddressConfigured,
        opts.filter
      );

      transactions.forEach((tx) => {
        populatedTransactions.push(tx);
      });
    }

    return populatedTransactions;
  }

  getRange(
    cfgBlockBarchSize: number,
    cfgFrom: bigint | undefined,
    savedpreviousFrom: bigint | undefined,
    savedlastFrom: bigint | undefined
  ): Range | undefined {
    // If [set up a from for cfg], return the from and limit equal the from batch size
    if (cfgFrom) {
      return {
        from: Number(cfgFrom),
        limit: cfgBlockBarchSize,
      };
    }

    if (savedpreviousFrom && savedlastFrom) {
      // If process [equal or different blocks], return the same lastFrom and limit equal the from batch size
      if (savedpreviousFrom === savedlastFrom || savedpreviousFrom !== savedlastFrom) {
        return {
          from: Number(savedlastFrom),
          limit: cfgBlockBarchSize,
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

  private createBatch(range: Range | undefined) {
    const batchSize = 100;
    const totalBatchLimit = range?.limit ?? batchSize;
    let limitBatch = 100;

    return {
      batchSize,
      totalBatchLimit,
      limitBatch,
    };
  }
}
