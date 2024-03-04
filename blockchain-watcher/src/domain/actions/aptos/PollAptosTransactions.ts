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

  private lastSequence?: bigint;
  private sequenceHeightCursor?: bigint;
  private previousSequence?: bigint;

  constructor(
    private readonly cfg: PollAptosTransactionsConfig,
    private readonly statsRepo: StatRepository,
    private readonly metadataRepo: MetadataRepository<PollAptosTransactionsMetadata>,
    private readonly repo: AptosRepository
  ) {
    super(cfg.id, statsRepo, cfg.interval);
    this.logger = winston.child({ module: "PollAptos", label: this.cfg.id });
  }

  protected async preHook(): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.sequenceHeightCursor = metadata.lastSequence;
      this.previousSequence = metadata.previousSequence;
      this.lastSequence = metadata.lastSequence;
    }
  }

  protected async hasNext(): Promise<boolean> {
    if (
      this.cfg.toSequence &&
      this.sequenceHeightCursor &&
      this.sequenceHeightCursor >= BigInt(this.cfg.toSequence)
    ) {
      this.logger.info(
        `[aptos][PollAptosTransactions] Finished processing all transactions from sequence ${this.cfg.fromSequence} to ${this.cfg.toSequence}`
      );
      return false;
    }

    return true;
  }

  protected async get(): Promise<any[]> {
    const range = this.getBlockRange();

    const events = await this.repo.getSequenceNumber(range);

    // save preveous sequence with last sequence and update last sequence with the new sequence
    this.previousSequence = this.lastSequence;
    this.lastSequence = BigInt(events[events.length - 1].sequence_number);

    if (this.previousSequence && this.lastSequence && this.previousSequence === this.lastSequence) {
      return [];
    }

    const transactions = await this.repo.getTransactionsForVersions(events);

    return transactions;
  }

  private getBlockRange(): Sequence | undefined {
    if (this.previousSequence && this.lastSequence && this.previousSequence === this.lastSequence) {
      return {
        fromSequence: Number(this.lastSequence),
        toSequence: Number(this.lastSequence) - Number(this.previousSequence) + 1,
      };
    }

    if (this.previousSequence && this.lastSequence && this.previousSequence !== this.lastSequence) {
      return {
        fromSequence: Number(this.lastSequence),
        toSequence: Number(this.lastSequence) - Number(this.previousSequence),
      };
    }

    if (this.lastSequence) {
      // check that it's not prior to the range start
      if (!this.cfg.fromSequence || BigInt(this.cfg.fromSequence) < this.lastSequence) {
        return {
          fromSequence: Number(this.lastSequence),
          toSequence:
            Number(this.lastSequence) + this.cfg.getBlockBatchSize() - Number(this.lastSequence),
        };
      }
    }
  }

  protected async persist(): Promise<void> {
    this.lastSequence = this.lastSequence;
    if (this.lastSequence) {
      await this.metadataRepo.save(this.cfg.id, {
        lastSequence: this.lastSequence,
        previousSequence: this.previousSequence,
      });
    }
  }

  protected report(): void {
    const labels = {
      job: this.cfg.id,
      chain: "aptos",
      commitment: this.cfg.getCommitment(),
    };
    this.statsRepo.count("job_execution", labels);
    this.statsRepo.measure("polling_cursor", this.lastSequence ?? 0n, {
      ...labels,
      type: "max",
    });
    this.statsRepo.measure("polling_cursor", this.lastSequence ?? 0n, {
      ...labels,
      type: "current",
    });
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
  lastSequence?: bigint;
  previousSequence?: bigint;
};

export type Sequence = {
  fromSequence: number;
  toSequence: number;
};
