import { AptosRepository, MetadataRepository, StatRepository } from "../../repositories";
import { GetAptosTransactionsByEvents } from "./GetAptosTransactionsByEvents";
import { GetAptosTransactions } from "./GetAptosTransactions";
import { AptosTransaction } from "../../entities/aptos";
import winston, { Logger } from "winston";
import { RunPollingJob } from "../RunPollingJob";

export class PollAptos extends RunPollingJob {
  private readonly getAptos: GetAptosTransactionsByEvents;
  protected readonly logger: Logger;

  private previousFrom?: bigint;
  private lastFrom?: bigint;
  private getAptosRecords: { [key: string]: any } = {
    GetAptosTransactionsByEvents,
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
    this.getAptos = new this.getAptosRecords[getAptos ?? "GetAptosTransactionsByEvents"](repo);
  }

  protected async preHook(): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.previousFrom = metadata.previousFrom;
      this.lastFrom = metadata.lastFrom;
    }
  }

  protected async hasNext(): Promise<boolean> {
    if (this.cfg.limit && this.previousFrom && this.previousFrom >= BigInt(this.cfg.limit)) {
      this.logger.info(
        `[aptos][PollAptos] Finished processing all transactions from sequence ${this.cfg.from} to ${this.cfg.limit}`
      );
      return false;
    }

    return true;
  }

  protected async get(): Promise<AptosTransaction[]> {
    const range = this.getAptos.getRange(
      this.cfg.getLimitBatchSize(),
      this.cfg.from,
      this.previousFrom,
      this.lastFrom
    );

    const records = await this.getAptos.execute(range, {
      addresses: this.cfg.addresses,
      filter: this.cfg.filter,
      previousFrom: this.previousFrom,
      lastFrom: this.lastFrom,
    });

    this.updateRange();

    return records;
  }

  private updateRange(): void {
    // Update the previousFrom and lastFrom based on the executed range
    const updatedRange = this.getAptos.getUpdatedRange();
    if (updatedRange) {
      this.previousFrom = updatedRange.previousFrom;
      this.lastFrom = updatedRange.lastFrom;
    }
  }

  protected async persist(): Promise<void> {
    if (this.lastFrom) {
      await this.metadataRepo.save(this.cfg.id, {
        previousFrom: this.previousFrom,
        lastFrom: this.lastFrom,
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
    this.statsRepo.measure("polling_cursor", this.lastFrom ?? 0n, {
      ...labels,
      type: "max",
    });
    this.statsRepo.measure("polling_cursor", this.previousFrom ?? 0n, {
      ...labels,
      type: "current",
    });
  }
}

export class PollAptosTransactionsConfig {
  constructor(private readonly props: PollAptosTransactionsConfigProps) {}

  public getLimitBatchSize() {
    return this.props.limitBatchSize ?? 100;
  }

  public getCommitment() {
    return this.props.commitment ?? "finalized";
  }

  public get id(): string {
    return this.props.id;
  }

  public get interval(): number | undefined {
    return this.props.interval;
  }

  public get from(): bigint | undefined {
    return this.props.from ? BigInt(this.props.from) : undefined;
  }

  public get limit(): bigint | undefined {
    return this.props.limit ? BigInt(this.props.limit) : undefined;
  }

  public get filter(): TransactionFilter {
    return this.props.filter;
  }

  public get addresses(): string[] {
    return this.props.addresses;
  }
}

export interface PollAptosTransactionsConfigProps {
  limitBatchSize?: number;
  from?: bigint;
  limit?: bigint;
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
  previousFrom?: bigint;
  lastFrom?: bigint;
};

export type TransactionFilter = {
  fieldName?: string;
  address: string;
  event?: string;
  type?: string;
};

export type Range = {
  from?: number;
  limit?: number;
};

export type PreviousRange = {
  previousFrom: bigint | undefined;
  lastFrom: bigint | undefined;
};

export type GetAptosOpts = {
  addresses: string[];
  filter: TransactionFilter;
  previousFrom?: bigint | undefined;
  lastFrom?: bigint | undefined;
};
