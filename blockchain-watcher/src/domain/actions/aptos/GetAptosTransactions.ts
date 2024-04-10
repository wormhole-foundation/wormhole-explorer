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
      `[aptos][exec] Processing range [previousFrom: ${opts.previousFrom} - lastFrom: ${opts.lastFrom}]`
    );

    const transactions = await this.repo.getTransactions({
      from: range?.from,
      limit: range?.limit,
    });

    // Only process transactions to the contract address configured
    const transactionsByAddressConfigured = transactions.filter((transaction) =>
      opts.filters.some((filter) =>
        filter.type?.includes(String(transaction.payload?.function).toLowerCase())
      )
    );

    const newLastFrom = BigInt(transactions[transactions.length - 1].version!);

    if (opts.previousFrom == newLastFrom) {
      return [];
    }

    if (transactionsByAddressConfigured.length > 0) {
      const transactions = await this.repo.getTransactionsByVersion(
        transactionsByAddressConfigured
      );

      transactions.forEach((tx) => {
        populatedTransactions.push(tx);
      });
    }

    this.logger.info(
      `[aptos][exec] Got ${populatedTransactions?.length} transactions to process for [addresses:${opts.addresses}][from: ${range?.from} - limit: ${range?.limit}]`
    );

    // Update lastFrom and previousFrom with the new lastFrom
    this.lastFrom = BigInt(transactions[transactions.length - 1].version!);
    this.previousFrom = opts.lastFrom;

    return populatedTransactions;
  }

  getRange(
    cfgLimitBatchSize: number,
    cfgFrom: bigint | undefined,
    savedpreviousFrom: bigint | undefined,
    savedlastFrom: bigint | undefined
  ): Range | undefined {
    // If [set up a from for cfg], return the from and limit equal the from batch size
    if (cfgFrom) {
      return {
        from: Number(cfgFrom),
        limit: cfgLimitBatchSize,
      };
    }

    if (savedpreviousFrom && savedlastFrom) {
      // If process [equal or different from], return the same lastFrom and limit equal the from batch size
      if (savedpreviousFrom === savedlastFrom || savedpreviousFrom !== savedlastFrom) {
        return {
          from: Number(savedlastFrom),
          limit: cfgLimitBatchSize,
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
