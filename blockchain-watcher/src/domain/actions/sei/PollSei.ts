import { MetadataRepository, SeiRepository, StatRepository } from "../../repositories";
import { RunPollingJob } from "../RunPollingJob";
import { GetSeiRedeems } from "./GetSeiRedeems";
import winston from "winston";

const ID = "watch-sei-logs";

export class PollSei extends RunPollingJob {
  protected readonly logger: winston.Logger;
  private readonly metadataRepo: MetadataRepository<PollSeiLogsMetadata>;
  private readonly getSeiRedeems: GetSeiRedeems;
  private readonly blockRepo: SeiRepository;
  private readonly statsRepo: StatRepository;

  private previousFrom?: bigint;
  private lastFrom?: bigint;
  private latestBlockHeight?: bigint;
  private blockHeightCursor?: bigint;
  private lastRange?: { fromBlock: bigint; toBlock: bigint };
  private cfg: PollSeiLogsConfig;

  constructor(
    blockRepo: SeiRepository,
    metadataRepo: MetadataRepository<PollSeiLogsMetadata>,
    statsRepo: StatRepository,
    cfg: PollSeiLogsConfig
  ) {
    super(cfg.id, statsRepo, cfg.interval);
    this.blockRepo = blockRepo;
    this.metadataRepo = metadataRepo;
    this.statsRepo = statsRepo;
    this.cfg = cfg;
    this.logger = winston.child({ module: "PollSei", label: this.cfg.id });
    this.getSeiRedeems = new GetSeiRedeems(blockRepo);
  }

  protected async preHook(): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.previousFrom = metadata.previousFrom;
      this.lastFrom = metadata.lastFrom;
    }
  }

  protected async hasNext(): Promise<boolean> {
    const hasFinished = this.cfg.hasFinished(this.blockHeightCursor);
    if (hasFinished) {
      this.logger.info(
        `[hasNext] PollSei: (${this.cfg.id}) Finished processing all blocks from ${this.cfg.fromBlock} to ${this.cfg.toBlock}`
      );
    }
    return !hasFinished;
  }

  protected async get(): Promise<any[]> {
    const seiRedeems = await this.getSeiRedeems.execute({
      addresses: this.cfg.addresses,
      previousFrom: this.previousFrom,
      lastFrom: this.lastFrom,
      chainId: this.cfg.chainId,
      blockBatchSize: this.cfg.getBlockBatchSize(),
    });

    this.updateRange();
    return seiRedeems;
  }

  private updateRange(): void {
    // Update the previousFrom and lastFrom based on the executed range
    const updatedRange = this.getSeiRedeems.getUpdatedRange();
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
      chain: "sei",
      commitment: "latest",
    };
    const lastFrom = this.lastFrom ?? 0n;
    const previousFrom = this.previousFrom ?? 0n;
    const diffCursor = BigInt(lastFrom) - BigInt(previousFrom);

    this.statsRepo.count("job_execution", labels);

    this.statsRepo.measure("polling_cursor", lastFrom, {
      ...labels,
      type: "max",
    });

    this.statsRepo.measure("polling_cursor", previousFrom, {
      ...labels,
      type: "current",
    });

    this.statsRepo.measure("polling_cursor", diffCursor, {
      ...labels,
      type: "diff",
    });
  }
}

export type PreviousRange = {
  previousFrom: bigint | undefined;
  lastFrom: bigint | undefined;
};

export type PollSeiLogsMetadata = {
  previousFrom?: bigint;
  lastFrom?: bigint;
};

export interface PollSeiLogsConfigProps {
  blockBatchSize?: number;
  fromBlock?: bigint;
  addresses: string[];
  interval?: number;
  toBlock?: bigint;
  chainId: number;
  chain: string;
  id?: string;
}

export type GetSeiOpts = {
  addresses: string[];
  previousFrom?: bigint | undefined;
  lastFrom?: bigint | undefined;
  chainId: number;
  blockBatchSize: number;
};

export class PollSeiLogsConfig {
  private props: PollSeiLogsConfigProps;

  constructor(props: PollSeiLogsConfigProps) {
    if (props.fromBlock && props.toBlock && props.fromBlock > props.toBlock) {
      throw new Error("fromBlock must be less than or equal to toBlock");
    }

    this.props = props;
  }
  public getBlockBatchSize() {
    return this.props.blockBatchSize ?? 100;
  }

  public hasFinished(currentFromBlock?: bigint): boolean {
    return (
      currentFromBlock != undefined &&
      this.props.toBlock != undefined &&
      currentFromBlock >= this.props.toBlock
    );
  }

  public get fromBlock() {
    return this.props.fromBlock ? BigInt(this.props.fromBlock) : undefined;
  }

  public setFromBlock(fromBlock: bigint | undefined) {
    this.props.fromBlock = fromBlock;
  }

  public get toBlock() {
    return this.props.toBlock;
  }

  public get interval() {
    return this.props.interval;
  }

  public get addresses() {
    return this.props.addresses.map((address) => address.toLowerCase());
  }

  public get id() {
    return this.props.id ?? ID;
  }

  public get chain() {
    return this.props.chain;
  }

  public get chainId() {
    return this.props.chainId;
  }
}
