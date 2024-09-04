import { NearRepository, MetadataRepository, StatRepository } from "../../repositories";
import { GetNearTransactions } from "./GetNearTransactions";
import { NearTransaction } from "../../entities/near";
import { RunPollingJob } from "../RunPollingJob";
import winston from "winston";

const MAX_DIFF_BLOCK_HEIGHT = 10_000;
const ID = "watch-near-logs";

export class PollNear extends RunPollingJob {
  protected readonly logger: winston.Logger;

  private readonly blockRepo: NearRepository;
  private readonly metadataRepo: MetadataRepository<PollNearMetadata>;
  private readonly statsRepo: StatRepository;
  private readonly getNear: GetNearTransactions;

  private cfg: PollNearConfig;
  private latestBlockHeight?: bigint;
  private blockHeightCursor?: bigint;
  private lastRange?: { fromBlock: bigint; toBlock: bigint };

  constructor(
    blockRepo: NearRepository,
    metadataRepo: MetadataRepository<PollNearMetadata>,
    statsRepo: StatRepository,
    cfg: PollNearConfig
  ) {
    super(cfg.id, statsRepo, cfg.interval);
    this.blockRepo = blockRepo;
    this.metadataRepo = metadataRepo;
    this.statsRepo = statsRepo;
    this.cfg = cfg;
    this.logger = winston.child({ module: "PollNear", label: this.cfg.id });
    this.getNear = new GetNearTransactions(blockRepo);
  }

  protected async preHook(): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.blockHeightCursor = BigInt(metadata.lastBlock);
    }
  }

  protected async hasNext(): Promise<boolean> {
    const hasFinished = this.cfg.hasFinished(this.blockHeightCursor);
    if (hasFinished) {
      this.logger.info(
        `[hasNext] PollNear: (${this.cfg.id}) Finished processing all blocks from ${this.cfg.fromBlock} to ${this.cfg.toBlock}`
      );
    }

    return !hasFinished;
  }

  protected async get(): Promise<NearTransaction[]> {
    this.latestBlockHeight = await this.blockRepo.getBlockHeight(this.cfg.commitment);

    const range = this.getBlockRange(this.latestBlockHeight!);

    const nearTransactions = await this.getNear.execute(range, {
      commitment: this.cfg.commitment,
      contracts: this.cfg.contracts,
      chainId: this.cfg.chainId,
      chain: this.cfg.chain,
    });

    this.lastRange = range;

    return nearTransactions;
  }

  protected async persist(): Promise<void> {
    this.blockHeightCursor = this.lastRange?.toBlock ?? this.blockHeightCursor;
    if (this.blockHeightCursor) {
      await this.metadataRepo.save(this.cfg.id, { lastBlock: this.blockHeightCursor });
    }
  }

  private getBlockRange(latestBlockHeight: bigint): {
    fromBlock: bigint;
    toBlock: bigint;
  } {
    let fromBlock = this.blockHeightCursor
      ? this.blockHeightCursor + 1n
      : this.cfg.fromBlock ?? latestBlockHeight;
    // fromBlock is configured and is greater than current block height, then we allow to skip blocks.
    if (
      this.blockHeightCursor &&
      this.cfg.fromBlock &&
      this.cfg.fromBlock > this.blockHeightCursor
    ) {
      fromBlock = this.cfg.fromBlock;
    }

    let toBlock = BigInt(fromBlock) + BigInt(this.cfg.blockBatchSize);
    // Limit toBlock to obtained block height
    if (toBlock > fromBlock && toBlock > latestBlockHeight) {
      // Restrict toBlock update because the latestBlockHeight may be outdated
      const diffBlockHeight = toBlock - latestBlockHeight;
      if (diffBlockHeight <= MAX_DIFF_BLOCK_HEIGHT) {
        toBlock = latestBlockHeight;
      }
    }

    // limit toBlock to configured toBlock
    if (this.cfg.toBlock && toBlock > this.cfg.toBlock) {
      toBlock = this.cfg.toBlock;
    }

    return { fromBlock: BigInt(fromBlock), toBlock: BigInt(toBlock) };
  }

  protected report(): void {
    const labels = {
      job: this.cfg.id,
      chain: this.cfg.chain ?? "",
      commitment: this.cfg.commitment,
    };
    const latestBlockHeight = this.latestBlockHeight ?? 0n;
    const blockHeightCursor = this.blockHeightCursor ?? 0n;
    const diffCursor = BigInt(latestBlockHeight) - BigInt(blockHeightCursor);

    this.statsRepo.count("job_execution", labels);

    this.statsRepo.measure("polling_cursor", latestBlockHeight, {
      ...labels,
      type: "max",
    });

    this.statsRepo.measure("polling_cursor", blockHeightCursor, {
      ...labels,
      type: "current",
    });

    this.statsRepo.measure("polling_cursor", diffCursor, {
      ...labels,
      type: "diff",
    });
  }
}

export type PollNearMetadata = {
  lastBlock: bigint;
};

export interface PollNearConfigProps {
  blockBatchSize?: number;
  commitment: string;
  environment: string;
  fromBlock?: bigint;
  contracts: string[];
  interval?: number;
  toBlock?: bigint;
  chainId: number;
  chain: string;
  id?: string;
}

export class PollNearConfig {
  private props: PollNearConfigProps;

  constructor(props: PollNearConfigProps) {
    if (props.fromBlock && props.toBlock && props.fromBlock > props.toBlock) {
      throw new Error("fromBlock must be less than or equal to toBlock");
    }

    this.props = props;
  }

  public hasFinished(currentFromBlock?: bigint): boolean {
    return (
      currentFromBlock != undefined &&
      this.props.toBlock != undefined &&
      currentFromBlock >= this.props.toBlock
    );
  }

  public setFromBlock(fromBlock: bigint | undefined) {
    this.props.fromBlock = fromBlock;
  }

  public get fromBlock() {
    return this.props.fromBlock ? BigInt(this.props.fromBlock) : undefined;
  }

  public get blockBatchSize() {
    return this.props.blockBatchSize ?? 100;
  }

  public get commitment() {
    return this.props.commitment;
  }

  public get toBlock() {
    return this.props.toBlock;
  }

  public get interval() {
    return this.props.interval;
  }

  public get id() {
    return this.props.id ?? ID;
  }

  public get chain() {
    return this.props.chain;
  }

  public get environment() {
    return this.props.environment;
  }

  public get chainId() {
    return this.props.chainId;
  }

  public get contracts(): string[] {
    return this.props.contracts;
  }
}

export type GetNearOpts = {
  commitment: string;
  contracts: string[];
  chainId: number;
  chain: string;
};
