import { EvmLog } from "../../entities";
import { EvmBlockRepository } from "../../repositories";
import winston from "winston";

export class GetEvmLogs {
  private readonly blockRepo: EvmBlockRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: EvmBlockRepository) {
    this.blockRepo = blockRepo;
    this.logger = winston.child({ module: "GetEvmLogs" });
  }

  async execute(range: Range, opts: GetEvmLogsOpts): Promise<EvmLog[]> {
    if (range.fromBlock > range.toBlock) {
      this.logger.info(
        `[exec] Invalid range [fromBlock: ${range.fromBlock} - toBlock: ${range.toBlock}]`
      );
      return [];
    }

    const logs = await this.blockRepo.getFilteredLogs(opts.chain, {
      fromBlock: range.fromBlock,
      toBlock: range.toBlock,
      addresses: opts.addresses ?? [], // Works when sending multiple addresses, but not multiple topics.
      topics: opts.topics ?? [],
    });

    const blockNumbers = new Set(logs.map((log) => log.blockNumber));
    const blocks = await this.blockRepo.getBlocks(opts.chain, blockNumbers);
    logs.forEach((log) => {
      const block = blocks[log.blockHash];
      log.blockTime = block.timestamp;
    });

    return logs;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};

type GetEvmLogsOpts = {
  addresses?: string[];
  topics?: string[];
  chain: string;
  environment: string;
};
