import winston from "winston";
import { solana } from "../../entities";

/**
 * Handling means mapping and forward to a given target if present.
 */
export class HandleSolanaTransactions<T> {
  cfg: HandleSolanaTxConfig;
  mapper: (txs: solana.Transaction, args: { programId: string }) => Promise<T>;
  target?: (parsed: T[]) => Promise<void>;
  logger: winston.Logger = winston.child({ module: "HandleSolanaTransaction" });

  constructor(
    cfg: HandleSolanaTxConfig,
    mapper: (txs: solana.Transaction) => Promise<T>,
    target?: (parsed: T[]) => Promise<void>
  ) {
    this.cfg = cfg;
    this.mapper = mapper;
    this.target = target;
  }

  public async handle(txs: solana.Transaction[]): Promise<T[]> {
    const filteredItems = txs.filter((tx) => {
      const hasError = tx.meta?.err;
      if (hasError)
        this.logger.warn(
          `Ignoring tx for program ${this.cfg.programId} in ${tx.slot} has error: ${tx.meta?.err}`
        );
      return !hasError;
    });

    let mappedItems: T[] = [];
    for (const tx of filteredItems) {
      const result = await this.mapper(tx, { programId: this.cfg.programId });
      if (result) {
        mappedItems = mappedItems.concat(result);
      }
    }

    if (this.target) {
      await this.target(mappedItems);
    } else {
      this.logger.warn(`No target for ${this.cfg.programId} txs`);
    }

    return mappedItems;
  }
}

export type HandleSolanaTxConfig = {
  programId: string;
};
