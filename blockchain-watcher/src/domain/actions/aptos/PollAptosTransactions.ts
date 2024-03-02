import { AptosRepository, MetadataRepository, StatRepository } from "../../repositories";
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
export class PollAptosTransactions extends RunPollingJob {
  protected readonly logger: Logger;

  private latestSequenceNumber?: number;
  private sequenceHeightCursor?: number;

  constructor(
    private readonly cfg: PollAptosTransactionsConfig,
    private readonly statsRepo: StatRepository,
    private readonly metadataRepo: MetadataRepository<PollAptosTransactionsMetadata>,
    private readonly repo: AptosRepository
  ) {
    super(cfg.id, statsRepo, cfg.interval);
    this.logger = winston.child({ module: "PollSui", label: this.cfg.id });
  }

  protected async preHook(): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.sequenceHeightCursor = Number(metadata.lastSequence!);
    }
  }

  protected async hasNext(): Promise<boolean> {
    return true;
  }

  protected async get(): Promise<any[]> {
    const range = this.getBlockRange();

    const events = await this.repo.getSequenceNumber(range);

    this.latestSequenceNumber = Number(events[events.length - 1].sequence_number);

    const transactions = await this.repo.getTransactionsForVersions(events);

    return transactions;
  }
  protected async persist(): Promise<void> {
    this.latestSequenceNumber = this.latestSequenceNumber;
    if (this.latestSequenceNumber) {
      await this.metadataRepo.save(this.cfg.id, { lastSequence: this.latestSequenceNumber });
    }
  }

  protected report(): void {}

  /**
   * Get the block range to extract.
   * @param latestBlockHeight - the latest known height of the chain
   * @returns an always valid range, in the sense from is always <= to
   */
  private getBlockRange(): Sequence | undefined {
    if (this.latestSequenceNumber) {
      // check that it's not prior to the range start
      if (!this.cfg.fromSequence || BigInt(this.cfg.fromSequence) < this.latestSequenceNumber) {
        return {
          fromSequence: Number(this.latestSequenceNumber),
          toSequence: Number(this.latestSequenceNumber) + this.cfg.getBlockBatchSize(),
        };
      }
    }
  }
}

export class PollAptosTransactionsConfig {
  constructor(private readonly props: PollAptosTransactionsConfigProps) {}

  public getBlockBatchSize() {
    return this.props.blockBatchSize ?? 100;
  }

  public getCommitment() {
    return this.props.commitment ?? "latest";
  }

  public get id(): string {
    return this.props.id;
  }

  public get interval(): number | undefined {
    return this.props.interval;
  }

  public get fromSequence(): bigint | undefined {
    return this.props.fromSequence ? BigInt(this.props.fromSequence) : undefined;
  }

  public get toSequence(): bigint | undefined {
    return this.props.toSequence ? BigInt(this.props.toSequence) : undefined;
  }

  public get filter(): TransactionFilter | undefined {
    return this.props.filter;
  }
}

interface PollAptosTransactionsConfigProps {
  fromSequence?: bigint;
  toSequence?: bigint;
  blockBatchSize?: number;
  commitment?: string;
  interval?: number;
  addresses: string[];
  topics: string[];
  id: string;
  chain: string;
  chainId: number;
  environment: string;
  filter?: TransactionFilter;
}

type PollAptosTransactionsMetadata = {
  lastSequence?: number;
};

export type Sequence = {
  fromSequence: number;
  toSequence: number;
};
