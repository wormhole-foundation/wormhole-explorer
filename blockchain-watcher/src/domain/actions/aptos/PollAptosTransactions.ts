import { AptosRepository, MetadataRepository, StatRepository } from "../../repositories";
import winston, { Logger } from "winston";
import { RunPollingJob } from "../RunPollingJob";

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
    const filter = this.cfg.filter;
    const range = this.getBlockRange();

    const events = await this.repo.getSequenceNumber(range, filter);

    // save previous sequence with last sequence and update last sequence with the new sequence
    this.previousSequence = this.lastSequence;
    this.lastSequence = BigInt(events[events.length - 1].sequence_number);

    if (this.previousSequence && this.lastSequence && this.previousSequence === this.lastSequence) {
      return [];
    }

    const transactions = await this.repo.getTransactionsForVersions(events, filter);

    return transactions;
  }

  private getBlockRange(): Sequence | undefined {
    if (this.previousSequence && this.lastSequence) {
      // if process the [same sequence], return the same last sequence and the to sequence equal 1
      if (this.previousSequence === this.lastSequence) {
        return {
          fromSequence: Number(this.lastSequence),
          toSequence: Number(this.lastSequence) - Number(this.previousSequence) + 1,
        };
      }

      // if process [different sequences], return the difference between the last sequence and the previous sequence plus 1
      if (this.previousSequence !== this.lastSequence) {
        return {
          fromSequence: Number(this.lastSequence),
          toSequence: Number(this.lastSequence) - Number(this.previousSequence) + 1,
        };
      }
    }

    if (this.lastSequence) {
      // if there is [no previous sequence], return the last sequence and the to sequence equal the block batch size
      if (!this.cfg.fromSequence || BigInt(this.cfg.fromSequence) < this.lastSequence) {
        return {
          fromSequence: Number(this.lastSequence),
          toSequence:
            Number(this.lastSequence) + this.cfg.getBlockBatchSize() - Number(this.lastSequence),
        };
      }
    }

    // if [set up a from sequence for cfg], return the from sequence and the to sequence equal the block batch size
    if (this.cfg.fromSequence) {
      return {
        fromSequence: Number(this.lastSequence),
        toSequence: this.cfg.getBlockBatchSize(),
      };
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

  public get filter(): TransactionFilter {
    return this.props.filter;
  }
}

export interface PollAptosTransactionsConfigProps {
  blockBatchSize?: number;
  fromSequence?: bigint;
  toSequence?: bigint;
  environment: string;
  commitment?: string;
  addresses: string[];
  interval?: number;
  topics: string[];
  chainId: number;
  filter: TransactionFilter;
  chain: string;
  id: string;
}

type PollAptosTransactionsMetadata = {
  previousSequence?: bigint;
  lastSequence?: bigint;
};

export type TransactionFilter = {
  fieldName: string;
  address: string;
  event: string;
};

export type Sequence = {
  fromSequence: number;
  toSequence: number;
};
