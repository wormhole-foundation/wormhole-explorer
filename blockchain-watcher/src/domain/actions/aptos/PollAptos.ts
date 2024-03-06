import { AptosRepository, MetadataRepository, StatRepository } from "../../repositories";
import { TransactionsByVersion } from "../../../infrastructure/repositories/aptos/AptosJsonRPCBlockRepository";
import { GetAptosTransactions } from "./GetAptosTransactions";
import { GetAptosSequences } from "./GetAptosSequences";
import winston, { Logger } from "winston";
import { RunPollingJob } from "../RunPollingJob";

export class PollAptos extends RunPollingJob {
  protected readonly logger: Logger;
  private readonly getAptos: GetAptosSequences;

  private lastSequence?: bigint;
  private sequenceHeightCursor?: bigint;
  private previousSequence?: bigint;
  private getAptosRecords: { [key: string]: any } = {
    GetAptosSequences,
    GetAptosTransactions,
  };

  constructor(
    private readonly cfg: PollAptosTransactionsConfig,
    private readonly statsRepo: StatRepository,
    private readonly metadataRepo: MetadataRepository<PollAptosTransactionsMetadata>,
    private readonly repo: AptosRepository,
    getAptos: string
  ) {
    super(cfg.id, statsRepo, cfg.interval);
    this.logger = winston.child({ module: "PollAptos", label: this.cfg.id });
    this.getAptos = new this.getAptosRecords[getAptos ?? "GetAptosSequences"](repo);
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
        `[aptos][PollAptos] Finished processing all transactions from sequence ${this.cfg.fromSequence} to ${this.cfg.toSequence}`
      );
      return false;
    }

    return true;
  }

  protected async get(): Promise<TransactionsByVersion[]> {
    const range = this.getAptos.getBlockRange(
      this.cfg.getBlockBatchSize(),
      this.cfg.fromSequence,
      this.previousSequence,
      this.lastSequence
    );

    const records = await this.getAptos.execute(range, {
      addresses: this.cfg.addresses,
      filter: this.cfg.filter,
      previousSequence: this.previousSequence,
      lastSequence: this.lastSequence,
    });

    const updatedRange = this.getAptos.updatedRange();
    this.previousSequence = updatedRange?.previousSequence;
    this.lastSequence = updatedRange?.lastSequence;

    return records;
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
    this.statsRepo.measure("polling_cursor", this.sequenceHeightCursor ?? 0n, {
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

  public get addresses(): string[] {
    return this.props.addresses;
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

export type PollAptosTransactionsMetadata = {
  previousSequence?: bigint;
  lastSequence?: bigint;
};

export type TransactionFilter = {
  fieldName: string;
  address: string;
  event: string;
  type: string;
};

export type Sequence = {
  fromSequence?: number;
  toSequence?: number;
};

export type Filter = {
  fieldName: string;
  address: string;
  event: string;
  type: string;
};
