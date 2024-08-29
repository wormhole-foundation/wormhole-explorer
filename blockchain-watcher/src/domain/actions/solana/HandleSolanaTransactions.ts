import { InstructionFound, TransactionFoundEvent, solana } from "../../entities";
import { StatRepository } from "../../repositories";
import winston from "winston";

/**
 * Handling means mapping and forward to a given target if present.
 */
export class HandleSolanaTransactions<T> {
  cfg: HandleSolanaTxConfig;
  mapper: <M>(txs: solana.Transaction, args: M) => Promise<T[]>;
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
    let mappedItems: T[] = [];
    for (const tx of txs) {
      const result = await this.mapper(tx, this.cfg);
      if (result.length) {
        const txs = result as TransactionFoundEvent<InstructionFound>[];
        txs.forEach((tx) => {
          this.report(tx.attributes.protocol);
        });
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

  private report(protocol: string) {
    if (!this.cfg.metricName) return;

    const labels = {
      job: this.cfg.id,
      chain: this.cfg.chain ?? "",
      commitment: this.cfg.commitment,
      protocol: protocol ?? "unknown",
    };
    this.statsRepo!.count(this.cfg.metricName, labels);
  }
}

export type HandleSolanaTxConfig = {
  metricName: string;
  commitment: string;
  chainId: number;
  chain: string;
  abis: {
    topic: string;
    abi: string;
  }[];
  id: string;

  // TODO: perhaps create mapper object in the config with the params instead
  // of having them in the handler config
  programId?: string;
  programs?: Record<string, string[]>;
};
