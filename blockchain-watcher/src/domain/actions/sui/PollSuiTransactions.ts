import winston, { Logger } from "winston";
import { Range } from "../../entities";
import { MetadataRepository, StatRepository, SuiRepository } from "../../repositories";
import { RunPollingJob } from "../RunPollingJob";
import { GetSuiTransactions } from "./GetSuiTransactions";

const DEFAULT_BATCH_SIZE = 10;

export class PollSuiTransactions extends RunPollingJob {
  protected readonly logger: Logger;

  private checkpointCursor?: bigint;
  private lastCheckpoint?: bigint;
  private lastRange?: Range;

  constructor(
    private readonly cfg: PollSuiConfig,
    private readonly statsRepo: StatRepository,
    private readonly metadataRepo: MetadataRepository<PollSuiMetadata>,
    private readonly repo: SuiRepository,
    private readonly action: GetSuiTransactions = new GetSuiTransactions(repo)
  ) {
    super(cfg.id, statsRepo, cfg.interval);
    this.logger = winston.child({ module: "PollSui", label: this.cfg.id });
  }

  protected async preHook(): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.checkpointCursor = BigInt(metadata.lastCheckpoint);
    }
  }

  async hasNext(): Promise<boolean> {
    if (this.cfg.to && this.checkpointCursor && this.checkpointCursor >= BigInt(this.cfg.to)) {
      this.logger.info(
        `[hasNext] Finished processing all checkpoints from ${this.cfg.from} to ${this.cfg.to}`
      );
      return false;
    }

    return true;
  }

  protected report(): void {
    const labels = {
      job: this.cfg.id,
      chain: "sui",
    };
    this.statsRepo.count("job_execution", labels);
    this.statsRepo.measure("polling_cursor", BigInt(this.lastCheckpoint ?? 0), {
      ...labels,
      type: "max",
    });
    this.statsRepo.measure("polling_cursor", BigInt(this.checkpointCursor ?? 0n), {
      ...labels,
      type: "current",
    });
  }

  protected async get(): Promise<any[]> {
    this.lastCheckpoint = await this.repo.getLastCheckpoint();

    if (this.lastCheckpoint === this.checkpointCursor) {
      this.logger.info(`No new checkpoints to process`);
      return [];
    }

    const range = this.getCheckpointRange(this.lastCheckpoint);

    this.logger.info(`Processing checkpoints from ${range.from} to ${range.to}`);
    const records = await this.action.execute(range);

    this.lastRange = range;

    return records;
  }

  protected async persist(): Promise<void> {
    this.checkpointCursor = this.lastRange?.to ?? this.checkpointCursor;
    if (this.checkpointCursor) {
      await this.metadataRepo.save(this.cfg.id, { lastCheckpoint: this.checkpointCursor });
    }
  }

  private getCheckpointRange(latest: bigint): Range {
    let from = this.checkpointCursor ? this.checkpointCursor + 1n : this.cfg.from ?? latest;

    // from higher than current cursor
    if (this.cfg.from && this.checkpointCursor && this.cfg.from > this.checkpointCursor) {
      from = this.cfg.from;
    }

    let to = from + BigInt(this.cfg.batchSize || DEFAULT_BATCH_SIZE) - 1n;

    // limit `to` to latest checkpoint
    if (to > from && to > latest) {
      to = latest;
    }

    // limit `to` to configured `to`
    if (this.cfg.to && to > this.cfg.to) {
      to = this.cfg.to;
    }

    return { from, to };
  }
}

export class PollSuiConfig {
  constructor(private readonly props: PollSuiConfigProps) {}

  public get id(): string {
    return this.props.id;
  }

  public get interval(): number | undefined {
    return this.props.interval;
  }

  public get batchSize(): number | undefined {
    return this.props.batchSize;
  }

  public get from(): bigint | undefined {
    return this.props.from ? BigInt(this.props.from) : undefined;
  }

  public get to(): bigint | undefined {
    return this.props.to ? BigInt(this.props.to) : undefined;
  }
}

export interface PollSuiConfigProps {
  id: string;
  interval?: number;
  batchSize?: number;
  from?: bigint | string | number;
  to?: bigint | string | number;
}

export type PollSuiMetadata = {
  lastCheckpoint: bigint;
};