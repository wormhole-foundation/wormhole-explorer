import { MetadataRepository, StatRepository, SuiRepository } from "../../repositories";
import { TransactionFilter } from "@mysten/sui.js/client";
import winston, { Logger } from "winston";
import { RunPollingJob } from "../RunPollingJob";

/**
 * Instead of fetching block per block and scanning for transactions,
 * this poller uses the queryTransactions function from the sui sdk
 * to set up a filter and retrieve all transactions that match it,
 * thus avoiding the need to scan blocks which may not have any
 * valuable information.
 *
 * Each queryTransactions request accepts a cursor parameter, which
 * is the digest of the last transaction of the previous page.
 */
export class PollSuiTransactions extends RunPollingJob {
  protected readonly logger: Logger;

  private cursor?: Cursor;
  private currentCheckpoint?: bigint;

  constructor(
    private readonly cfg: PollSuiTransactionsConfig,
    private readonly statsRepo: StatRepository,
    private readonly metadataRepo: MetadataRepository<PollSuiTransactionsMetadata>,
    private readonly repo: SuiRepository
  ) {
    super(cfg.id, statsRepo, cfg.interval);
    this.logger = winston.child({ module: "PollSui", label: this.cfg.id });
  }

  protected async preHook(): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.currentCheckpoint = metadata.lastCursor?.checkpoint;
      this.cursor = metadata.lastCursor;
    }
  }

  async hasNext(): Promise<boolean> {
    if (this.cfg.to && this.cursor && this.cursor.checkpoint >= BigInt(this.cfg.to)) {
      this.logger.info(
        `[sui][PollSuiTransactions] Finished processing all transactions from checkpoint ${this.cfg.from} to ${this.cfg.to}`
      );
      return false;
    }

    return true;
  }

  protected async get(): Promise<any[]> {
    this.cursor = await this.getCursor();

    let txs = await this.repo.queryTransactions(this.cfg.filter, this.cursor.digest);

    if (txs.length === 0) {
      return [];
    }

    // clamp down to config range if present
    if (this.cfg.to) {
      const lastCheckpointIndex = txs.find(
        (tx) => tx.checkpoint === (this.cfg.to! + 1n).toString()
      );
      if (lastCheckpointIndex) {
        // take until before the tx of the checkpoint out of range
        txs = txs.slice(0, txs.indexOf(lastCheckpointIndex));
      }
    }

    const lastTx = txs[txs.length - 1];
    const newCursor = { checkpoint: BigInt(lastTx.checkpoint), digest: lastTx.digest };

    this.logger.info(
      `[sui][PollSuiTransactions] Got ${txs.length} txs from ${this.cursor.checkpoint} to ${newCursor.checkpoint}`
    );

    this.cursor = newCursor;

    return txs;
  }

  private async getCursor(): Promise<Cursor> {
    if (this.cursor) {
      // check that it's not prior to the range start
      if (!this.cfg.from || BigInt(this.cfg.from) < this.cursor.checkpoint) {
        return this.cursor;
      }
    }

    // initial cursor if not set
    const from = this.cfg.from ? this.cfg.from : await this.repo.getLastCheckpointNumber();
    const prevCheckpoint = await this.repo.getCheckpoint(from - 1n);
    return {
      checkpoint: BigInt(prevCheckpoint.sequenceNumber),
      digest: prevCheckpoint.transactions[prevCheckpoint.transactions.length - 1],
    };
  }

  protected async persist(): Promise<void> {
    if (this.cursor) {
      await this.metadataRepo.save(this.cfg.id, { lastCursor: this.cursor });
    }
  }

  protected report(): void {
    const labels = {
      job: this.cfg.id,
      chain: "sui",
      commitment: "immediate",
    };
    this.statsRepo.count("job_execution", labels);
    this.statsRepo.measure("polling_cursor", BigInt(this.cursor?.checkpoint ?? 0), {
      ...labels,
      type: "max",
    });
    this.statsRepo.measure("polling_cursor", BigInt(this.currentCheckpoint ?? 0n), {
      ...labels,
      type: "current",
    });
  }
}

export class PollSuiTransactionsConfig {
  constructor(private readonly props: PollSuiTransactionsConfigProps) {}

  public get id(): string {
    return this.props.id;
  }

  public get interval(): number | undefined {
    return this.props.interval;
  }

  public get from(): bigint | undefined {
    return this.props.from ? BigInt(this.props.from) : undefined;
  }

  public get to(): bigint | undefined {
    return this.props.to ? BigInt(this.props.to) : undefined;
  }

  public get filter(): TransactionFilter | undefined {
    return this.props.filter;
  }
}

export interface PollSuiTransactionsConfigProps {
  id: string;
  interval?: number;
  from?: bigint | string | number;
  to?: bigint | string | number;
  filter?: TransactionFilter;
}

export type PollSuiTransactionsMetadata = {
  lastCursor?: Cursor;
};

type Cursor = {
  digest: string;
  checkpoint: bigint;
};
