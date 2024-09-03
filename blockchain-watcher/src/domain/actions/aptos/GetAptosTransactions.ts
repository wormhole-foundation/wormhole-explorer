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
    const { previousFrom, lastFrom, cfgFrom, filters } = opts;
    let populatedTransactions: AptosTransaction[] = [];

    this.logger.info(
      `[aptos][exec] Processing range [previousFrom: ${previousFrom} - lastFrom: ${lastFrom}]`
    );

    const transactions = await this.repo.getTransactions({
      from: range?.from,
      limit: range?.limit,
    });

    // Validate if the transactions are the same or there is a delay
    const newLastFrom = BigInt(transactions[transactions.length - 1].version!);
    if (!cfgFrom && lastFrom && lastFrom >= newLastFrom) {
      this.logger.warn(
        `[aptos][exec] Processing the same block or encountering a delay [lastFrom: ${lastFrom} - newLastFrom: ${newLastFrom}]`
      );
      return [];
    }

    // Only process transactions to the contract address configured
    const transactionsByAddressConfigured = transactions.filter((transaction) =>
      filters.some((filter) =>
        filter.type?.includes(String(transaction.payload?.function).toLowerCase())
      )
    );

    if (transactionsByAddressConfigured.length > 0) {
      const txsByVersion = await this.repo.getTransactionsByVersion(
        transactionsByAddressConfigured
      );
      txsByVersion.forEach((tx) => {
        populatedTransactions.push(tx);
      });
    }

    this.logger.info(
      `[aptos][exec] Got ${populatedTransactions?.length} transactions to process for [addresses:${opts.addresses}][from: ${range?.from} - limit: ${range?.limit}]`
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
      // If process [equal or different from], return the same lastFrom and limit equal the from batch size
      if (savedPreviousFrom === savedlastFrom || savedPreviousFrom <= savedlastFrom) {
        return {
          from: Number(savedlastFrom),
          limit: cfgLimitBatchSize,
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
