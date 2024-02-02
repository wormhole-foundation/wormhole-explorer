import { StatRepository } from "../../repositories";
import { solana } from "../../entities";
import winston from "winston";

/**
 * Handling means mapping and forward to a given target if present.
 */
export class HandleSolanaTransactions<T> {
  cfg: HandleSolanaTxConfig;
  mapper: (txs: solana.Transaction, args: { programId: string }) => Promise<T[]>;
  target?: (parsed: T[]) => Promise<void>;
  logger: winston.Logger = winston.child({ module: "HandleSolanaTransaction" });
  statsRepo?: StatRepository;

  constructor(
    cfg: HandleSolanaTxConfig,
    mapper: (tx: solana.Transaction) => Promise<T[]>,
    target?: (parsed: T[]) => Promise<void>,
    statsRepo?: StatRepository
  ) {
    this.cfg = cfg;
    this.mapper = mapper;
    this.target = target;
    this.statsRepo = statsRepo;
  }

  public async handle(txs: solana.Transaction[]): Promise<T[]> {
    const filteredItems = txs.filter((tx) => {
      const hasError = tx.meta?.err;
      if (hasError)
        this.logger.warn(
          `Ignoring tx for program ${this.cfg.programId} in ${tx.slot} has error: ${JSON.stringify(
            tx.meta?.err
          )}`
        );
      return !hasError;
    });

    let mappedItems: T[] = [];
    for (const tx of filteredItems) {
      const result = await this.mapper(tx, { programId: this.cfg.programId });
      if (result.length) {
        this.report();
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

  private report() {
    const labels = {
      job: this.cfg.id,
      chain: this.cfg.chain ?? "",
      commitment: this.cfg.commitment,
    };
    this.statsRepo!.count(this.cfg.metricName, labels);
  }
}

export type HandleSolanaTxConfig = {
  metricName: string;
  programId: string;
  commitment: string;
  chainId: number;
  chain: string;
  abi: string;
  id: string;
};
