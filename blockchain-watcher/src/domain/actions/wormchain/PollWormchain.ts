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

  private latestBlockHeight?: bigint;
  private blockHeightCursor?: bigint;
  private lastRange?: { fromBlock: bigint; toBlock: bigint };
  private cfg: PollWormchainLogsConfig;
  private getWormchainRecords: { [key: string]: any } = {
    GetWormchainRedeems,
    GetWormchainLogs,
  };

  constructor(
    blockRepo: WormchainRepository,
    metadataRepo: MetadataRepository<PollWormchainLogsMetadata>,
    statsRepo: StatRepository,
    cfg: PollWormchainLogsConfig,
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
      this.blockHeightCursor = BigInt(metadata.lastBlock);
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
    this.latestBlockHeight = await this.blockRepo.getBlockHeight(this.cfg.chainId);

    if (!this.latestBlockHeight) {
      throw new Error(`Could not obtain latest block height: ${this.latestBlockHeight}`);
    }

    const range = this.getBlockRange(this.latestBlockHeight);

    const records = await this.getWormchain.execute(range, {
      addresses: this.cfg.addresses,
      chainId: this.cfg.chainId,
    });

    this.lastRange = range;

    return records;
  }

  private getBlockRange(latestBlockHeight: bigint): {
    fromBlock: bigint;
    toBlock: bigint;
  } {
    let fromBlock = this.blockHeightCursor
      ? this.blockHeightCursor + 1n
      : this.cfg.fromBlock ?? latestBlockHeight;
    // fromBlock is configured and is greater than current block height, then we allow to skip blocks
    if (
      this.blockHeightCursor &&
      this.cfg.fromBlock &&
      this.cfg.fromBlock > this.blockHeightCursor
    ) {
      fromBlock = this.cfg.fromBlock;
    }

    let toBlock = BigInt(fromBlock) + BigInt(this.cfg.getBlockBatchSize());
    // limit toBlock to obtained block height
    if (toBlock > fromBlock && toBlock > latestBlockHeight) {
      toBlock = latestBlockHeight;
    }
    // limit toBlock to configured toBlock
    if (this.cfg.toBlock && toBlock > this.cfg.toBlock) {
      toBlock = this.cfg.toBlock;
    }

    return { fromBlock: BigInt(fromBlock), toBlock: BigInt(toBlock) };
  }

  protected async persist(): Promise<void> {
    this.blockHeightCursor = this.lastRange?.toBlock ?? this.blockHeightCursor;
    if (this.blockHeightCursor) {
      await this.metadataRepo.save(this.cfg.id, { lastBlock: this.blockHeightCursor });
    }
  }

  protected report(): void {
    const labels = {
      job: this.cfg.id,
      chain: this.cfg.chain ?? "",
      commitment: this.cfg.getCommitment(),
    };
    const latestBlockHeight = this.latestBlockHeight ?? 0n;
    const blockHeightCursor = this.blockHeightCursor ?? 0n;
    const diffCursor = latestBlockHeight - blockHeightCursor;

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

export type PollWormchainLogsMetadata = {
  lastBlock: bigint;
};

export interface PollWormchainLogsConfigProps {
  blockBatchSize?: number;
  commitment?: string;
  fromBlock?: bigint;
  addresses: string[];
  interval?: number;
  toBlock?: bigint;
  chainId: number;
  chain: string;
  id?: string;
}

export class PollWormchainLogsConfig {
  private props: PollWormchainLogsConfigProps;

  constructor(props: PollWormchainLogsConfigProps) {
    if (props.fromBlock && props.toBlock && props.fromBlock > props.toBlock) {
      throw new Error("fromBlock must be less than or equal to toBlock");
    }

    this.props = props;
  }

  public getBlockBatchSize() {
    return this.props.blockBatchSize ?? 100;
  }

  public getCommitment() {
    return this.props.commitment ?? "latest";
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
