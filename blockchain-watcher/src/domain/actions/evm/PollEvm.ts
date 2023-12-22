import { EvmLog, EvmTransaction } from "../../entities";
import { RunPollingJob } from "../RunPollingJob";
import { GetEvmLogs } from "./GetEvmLogs";
import { EvmBlockRepository, MetadataRepository, StatRepository } from "../../repositories";
import winston from "winston";
import { GetEvmTransactions } from "./GetEvmTransactions";

const ID = "watch-evm-logs";

/**
 * PollEvm is an action that watches for new blocks and extracts logs from them.
 */
export class PollEvm extends RunPollingJob {
  protected readonly logger: winston.Logger;

  private readonly blockRepo: EvmBlockRepository;
  private readonly metadataRepo: MetadataRepository<PollEvmLogsMetadata>;
  private readonly statsRepository: StatRepository;
  private readonly getEvm: GetEvmLogs;

  private cfg: PollEvmLogsConfig;
  private latestBlockHeight?: bigint;
  private blockHeightCursor?: bigint;
  private lastRange?: { fromBlock: bigint; toBlock: bigint };
  private getEvmRecords: { [key: string]: any } = {
    GetEvmLogs,
    GetEvmTransactions,
  };

  constructor(
    blockRepo: EvmBlockRepository,
    metadataRepo: MetadataRepository<PollEvmLogsMetadata>,
    statsRepository: StatRepository,
    cfg: PollEvmLogsConfig,
    getEvm: string
  ) {
    super(cfg.interval ?? 1_000, cfg.id, statsRepository);
    this.blockRepo = blockRepo;
    this.metadataRepo = metadataRepo;
    this.statsRepository = statsRepository;
    this.cfg = cfg;
    this.logger = winston.child({ module: "PollEvm", label: this.cfg.id });
    this.getEvm = new this.getEvmRecords[getEvm ?? "GetEvmLogs"](blockRepo);
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
        `[hasNext] PollEvm: (${this.cfg.id}) Finished processing all blocks from ${this.cfg.fromBlock} to ${this.cfg.toBlock}`
      );
    }

    return !hasFinished;
  }

  protected async get(): Promise<EvmLog[] | EvmTransaction[]> {
    this.report();

    this.latestBlockHeight = await this.blockRepo.getBlockHeight(
      this.cfg.chain,
      this.cfg.getCommitment()
    );

    const range = this.getBlockRange(this.latestBlockHeight);

    const records = await this.getEvm.execute(range, {
      chain: this.cfg.chain,
      addresses: this.cfg.addresses,
      topics: this.cfg.topics,
      environment: this.cfg.environment,
    });

    this.lastRange = range;

    return records;
  }

  protected async persist(): Promise<void> {
    this.blockHeightCursor = this.lastRange?.toBlock ?? this.blockHeightCursor;
    if (this.blockHeightCursor) {
      await this.metadataRepo.save(this.cfg.id, { lastBlock: this.blockHeightCursor });
    }
  }

  /**
   * Get the block range to extract.
   * @param latestBlockHeight - the latest known height of the chain
   * @returns an always valid range, in the sense from is always <= to
   */
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

  private report(): void {
    const labels = {
      job: this.cfg.id,
      chain: this.cfg.chain ?? "",
      commitment: this.cfg.getCommitment(),
    };
    this.statsRepository.count("job_execution", labels);
    this.statsRepository.measure("polling_cursor", this.latestBlockHeight ?? 0n, {
      ...labels,
      type: "max",
    });
    this.statsRepository.measure("polling_cursor", this.blockHeightCursor ?? 0n, {
      ...labels,
      type: "current",
    });
  }
}

export type PollEvmLogsMetadata = {
  lastBlock: bigint;
};

export interface PollEvmLogsConfigProps {
  fromBlock?: bigint;
  toBlock?: bigint;
  blockBatchSize?: number;
  commitment?: string;
  interval?: number;
  addresses: string[];
  topics: string[];
  id?: string;
  chain: string;
  environment: string;
}

export class PollEvmLogsConfig {
  private props: PollEvmLogsConfigProps;

  constructor(props: PollEvmLogsConfigProps) {
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
    return this.props.addresses;
  }

  public get topics() {
    return this.props.topics;
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

  static fromBlock(chain: string, fromBlock: bigint) {
    return new PollEvmLogsConfig({ chain, fromBlock, addresses: [], topics: [], environment: "" });
  }
}
