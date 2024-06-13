import { MetadataRepository, StatRepository, WormchainRepository } from "../../repositories";
import { GetWormchainRedeems } from "./GetWormchainRedeems";
import { GetWormchainLogs } from "./GetWormchainLogs";
import { RunPollingJob } from "../RunPollingJob";
import winston from "winston";

const ID = "watch-wormchain-logs";

export class PollWormchain extends RunPollingJob {
  protected readonly logger: winston.Logger;
  private readonly metadataRepo: MetadataRepository<PollWormchainLogsMetadata>;
  private readonly getWormchain: GetWormchainLogs;
  private readonly blockRepo: WormchainRepository;
  private readonly statsRepo: StatRepository;
  private getWormchainRecords: { [key: string]: any } = {
    GetWormchainRedeems,
    GetWormchainLogs,
  };

  private previousFrom?: bigint;
  private lastFrom?: bigint;
  private latestBlockHeight?: bigint;
  private blockHeightCursor?: bigint;
  private lastRange?: { fromBlock: bigint; toBlock: bigint };
  private cfg: PollWormchainConfig;

  constructor(
    blockRepo: WormchainRepository,
    metadataRepo: MetadataRepository<PollWormchainLogsMetadata>,
    statsRepo: StatRepository,
    cfg: PollWormchainConfig,
    getWormchain: string
  ) {
    super(cfg.id, statsRepo, cfg.interval);
    this.blockRepo = blockRepo;
    this.metadataRepo = metadataRepo;
    this.statsRepo = statsRepo;
    this.cfg = cfg;
    this.logger = winston.child({ module: "PollWormchain", label: this.cfg.id });
    this.getWormchain = new this.getWormchainRecords[getWormchain ?? "GetWormchainLogs"](blockRepo);
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
        `[hasNext] PollWormchain: (${this.cfg.id}) Finished processing all blocks from ${this.cfg.fromBlock} to ${this.cfg.toBlock}`
      );
    }
    return !hasFinished;
  }

  protected async get(): Promise<any[]> {
    const wormchainTxs = await this.getWormchain.execute({
      addresses: this.cfg.addresses,
      previousFrom: this.previousFrom,
      lastFrom: this.lastFrom,
      chainId: this.cfg.chainId,
      blockBatchSize: this.cfg.getBlockBatchSize(),
    });

    this.updateRange();
    return wormchainTxs;
  }

  private updateRange(): void {
    // Update the previousFrom and lastFrom based on the executed range
    const updatedRange = this.getWormchain.getUpdatedRange();
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
      chain: "wormchain",
      commitment: "immediate",
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

export type PollWormchainLogsMetadata = {
  previousFrom?: bigint;
  lastFrom?: bigint;
};

export interface PollWormchainConfigProps {
  blockBatchSize?: number;
  fromBlock?: bigint;
  addresses: string[];
  interval?: number;
  toBlock?: bigint;
  chainId: number;
  chain: string;
  id?: string;
}

export type GetWormchainOpts = {
  addresses: string[];
  previousFrom?: bigint | undefined;
  lastFrom?: bigint | undefined;
  chainId: number;
  blockBatchSize: number;
};

export class PollWormchainConfig {
  private props: PollWormchainConfigProps;

  constructor(props: PollWormchainConfigProps) {
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
