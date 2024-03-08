import { AptosRepository, MetadataRepository, StatRepository } from "../../repositories";
import { TransactionsByVersion } from "../../../infrastructure/repositories/aptos/AptosJsonRPCBlockRepository";
import { GetAptosTransactions } from "./GetAptosTransactions";
import { GetAptosSequences } from "./GetAptosSequences";
import winston, { Logger } from "winston";
import { RunPollingJob } from "../RunPollingJob";

export class PollAptos extends RunPollingJob {
  protected readonly logger: Logger;
  private readonly getAptos: GetAptosSequences;

  private lastBlock?: bigint;
  private sequenceHeightCursor?: bigint;
  private previousBlock?: bigint;
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
      this.sequenceHeightCursor = metadata.lastBlock;
      this.previousBlock = metadata.previousBlock;
      this.lastBlock = metadata.lastBlock;
    }
  }

  protected async hasNext(): Promise<boolean> {
    if (
      this.cfg.toSequence &&
      this.sequenceHeightCursor &&
      this.sequenceHeightCursor >= BigInt(this.cfg.toSequence)
    ) {
      this.logger.info(
        `[aptos][PollAptos] Finished processing all transactions from sequence ${this.cfg.fromBlock} to ${this.cfg.toSequence}`
      );
      return false;
    }

    return true;
  }

  protected async get(): Promise<TransactionsByVersion[]> {
    const range = this.getAptos.getBlockRange(
      this.cfg.getBlockBatchSize(),
      this.cfg.fromBlock,
      this.previousBlock,
      this.lastBlock
    );

    const records = await this.getAptos.execute(range, {
      addresses: this.cfg.addresses,
      filter: this.cfg.filter,
      previousBlock: this.previousBlock,
      lastBlock: this.lastBlock,
    });

    this.updateBlockRange();

    return records;
  }

  private updateBlockRange(): void {
    // Update the previousBlock and lastBlock based on the executed range
    const updatedRange = this.getAptos.getUpdatedRange();
    if (updatedRange) {
      this.previousBlock = updatedRange.previousBlock;
      this.lastBlock = updatedRange.lastBlock;
    }
  }

  protected async persist(): Promise<void> {
    if (this.lastBlock) {
      await this.metadataRepo.save(this.cfg.id, {
        previousBlock: this.previousBlock,
        lastBlock: this.lastBlock,
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
    this.statsRepo.measure("polling_cursor", this.lastBlock ?? 0n, {
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
    return this.props.commitment ?? "finalized";
  }

  public get id(): string {
    return this.props.id;
  }

  public get interval(): number | undefined {
    return this.props.interval;
  }

  public get fromBlock(): bigint | undefined {
    return this.props.fromBlock ? BigInt(this.props.fromBlock) : undefined;
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
  fromBlock?: bigint;
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
  previousBlock?: bigint;
  lastBlock?: bigint;
};

export type TransactionFilter = {
  fieldName?: string;
  address: string;
  event?: string;
  type?: string;
};

export type Block = {
  fromBlock?: number;
  toBlock?: number;
};
