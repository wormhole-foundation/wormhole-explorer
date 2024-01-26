import winston, { Logger } from "winston";
import { RunPollingJob } from "../RunPollingJob";
import { MetadataRepository, StatRepository, SuiRepository } from "../../repositories";
import { Range } from "../../entities";
import { GetSuiTransactions } from "./GetSuiTransactions";
import { SuiClient, getFullnodeUrl } from "@mysten/sui.js/client";

const DEFAULT_BATCH_SIZE = 10;

export class PollSui extends RunPollingJob {
  protected readonly logger: Logger;

  private action: GetSuiTransactions;
  private checkpointCursor?: bigint;
  private lastCheckpoint?: bigint;
  private lastRange?: Range;

  constructor(
    private readonly cfg: PollSuiConfig,
    private readonly statsRepo: StatRepository,
    private readonly metadataRepo: MetadataRepository<PollSuiMetadata>,
    private readonly repo: SuiRepository
  ) {
    super(cfg.id, statsRepo, cfg.interval);
    this.logger = winston.child({ module: "PollSui", label: this.cfg.id });
    this.action = new GetSuiTransactions(repo);
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
    let from = this.checkpointCursor ? this.checkpointCursor + 1n : BigInt(this.cfg.from ?? latest);

    // from higher than current cursor
    if (this.cfg.from && this.checkpointCursor && this.cfg.from > this.checkpointCursor) {
      from = BigInt(this.cfg.from);
    }

    let to = from + BigInt(this.cfg.batchSize || DEFAULT_BATCH_SIZE - 1);
    if (to > from && to > latest) {
      to = latest;
    }

    if (this.cfg.to && to > BigInt(this.cfg.to)) {
      to = BigInt(this.cfg.to);
    }

    return { from, to };
  }
}

export interface PollSuiConfig {
  id: string;
  interval?: number;
  batchSize?: number;

  // TODO: make these bigint
  from?: number;
  to?: number;
}

export type PollSuiMetadata = {
  lastCheckpoint: bigint;
};
