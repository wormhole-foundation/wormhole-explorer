import { EvmLog } from "../entities";
import { EvmBlockRepository, MetadataRepository, StatRepository } from "../repositories";
import { setTimeout } from "timers/promises";
import winston from "winston";

const ID = "watch-evm-logs";
let ref: any;

/**
 * PollEvmLogs is an action that watches for new blocks and extracts logs from them.
 */
export class PollEvmLogs {
  private readonly logger: winston.Logger = winston.child({ module: "PollEvmLogs" });

  private readonly blockRepo: EvmBlockRepository;
  private readonly metadataRepo: MetadataRepository<PollEvmLogsMetadata>;
  private readonly statsRepository: StatRepository;
  private cfg: PollEvmLogsConfig;

  private latestBlockHeight?: bigint;
  private blockHeightCursor?: bigint;
  private started: boolean = false;

  constructor(
    blockRepo: EvmBlockRepository,
    metadataRepo: MetadataRepository<PollEvmLogsMetadata>,
    statsRepository: StatRepository,
    cfg: PollEvmLogsConfig
  ) {
    this.blockRepo = blockRepo;
    this.metadataRepo = metadataRepo;
    this.statsRepository = statsRepository;
    this.cfg = cfg;
  }

  public async start(handlers: ((logs: EvmLog[]) => Promise<any>)[]): Promise<void> {
    const metadata = await this.metadataRepo.get(this.cfg.id);
    if (metadata) {
      this.blockHeightCursor = BigInt(metadata.lastBlock);
    }

    this.started = true;
    this.watch(handlers);
  }

  private async watch(handlers: ((logs: EvmLog[]) => Promise<void>)[]): Promise<void> {
    while (this.started) {
      this.report();
      if (this.cfg.hasFinished(this.blockHeightCursor)) {
        this.logger.info(
          `PollEvmLogs: (${this.cfg.id}) Finished processing all blocks from ${this.cfg.fromBlock} to ${this.cfg.toBlock}`
        );
        await this.stop();
        break;
      }

      this.latestBlockHeight = await this.blockRepo.getBlockHeight(this.cfg.getCommitment());

      const range = this.getBlockRange(this.latestBlockHeight);

      if (range.fromBlock > this.latestBlockHeight) {
        this.logger.info(`Next range is after latest block height, waiting...`);
        ref = await setTimeout(this.cfg.interval ?? 1_000, undefined);
        continue;
      }

      const logs = await this.blockRepo.getFilteredLogs({
        fromBlock: range.fromBlock,
        toBlock: range.toBlock,
        addresses: this.cfg.addresses, // Works when sending multiple addresses, but not multiple topics.
        topics: [], // this.cfg.topics => will be applied by handlers
      });

      const blockNumbers = new Set(logs.map((log) => log.blockNumber));
      const blocks = await this.blockRepo.getBlocks(blockNumbers);
      logs.forEach((log) => {
        const block = blocks[log.blockHash];
        log.blockTime = block.timestamp;
      });

      // TODO: add error handling.
      await Promise.all(handlers.map((handler) => handler(logs)));

      await this.metadataRepo.save(this.cfg.id, { lastBlock: range.toBlock });
      this.blockHeightCursor = range.toBlock;

      ref = await setTimeout(this.cfg.interval ?? 1_000, undefined);
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

    let toBlock = fromBlock + BigInt(this.cfg.getBlockBatchSize());
    // limit toBlock to obtained block height
    if (toBlock > fromBlock && toBlock > latestBlockHeight) {
      toBlock = latestBlockHeight;
    }
    // limit toBlock to configured toBlock
    if (this.cfg.toBlock && toBlock > this.cfg.toBlock) {
      toBlock = this.cfg.toBlock;
    }

    return { fromBlock, toBlock };
  }

  private report(): void {
    const labels = {
      job: this.cfg.id,
      chain: this.cfg.chain ?? "",
      commitment: this.cfg.getCommitment(),
    };
    this.statsRepository.count("job_execution", labels);
    this.statsRepository.measure("block_height", this.latestBlockHeight ?? 0n, labels);
    this.statsRepository.measure("block_cursor", this.blockHeightCursor ?? 0n, labels);
  }

  public async stop(): Promise<void> {
    clearTimeout(ref);
    this.started = false;
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
  chain?: string;
}

export class PollEvmLogsConfig {
  private props: PollEvmLogsConfigProps;

  constructor(props: PollEvmLogsConfigProps = { addresses: [], topics: [] }) {
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

  public hasFinished(currentFromBlock?: bigint) {
    return currentFromBlock && this.props.toBlock && currentFromBlock >= this.props.toBlock;
  }

  public get fromBlock() {
    return this.props.fromBlock;
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

  static fromBlock(fromBlock: bigint) {
    const cfg = new PollEvmLogsConfig();
    cfg.props.fromBlock = fromBlock;
    return cfg;
  }
}
