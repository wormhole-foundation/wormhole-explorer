import { solana } from "../../entities";

/**
 * Handling means mapping and forward to a given target if present.
 */
export class HandleSolanaTransaction<T> {
  cfg: HandleSolanaTxConfig;
  mapper: (txs: solana.Transaction) => T;
  target?: (parsed: T[]) => Promise<void>;

  constructor(
    cfg: HandleSolanaTxConfig,
    mapper: (txs: solana.Transaction) => T,
    target?: (parsed: T[]) => Promise<void>
  ) {
    this.cfg = cfg;
    this.mapper = mapper;
    this.target = target;
  }

  public async handle(txs: solana.Transaction[]): Promise<T[]> {
    const mappedItems = txs.filter((tx) => !tx.meta?.err).map((tx) => this.mapper(tx));

    if (this.target) await this.target(mappedItems);

    return mappedItems;
  }
}

export type HandleSolanaTxConfig = {
  programId: string;
};
