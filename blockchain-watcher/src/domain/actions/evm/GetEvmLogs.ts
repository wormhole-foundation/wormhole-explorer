import { EvmBlockRepository } from "../../repositories";
import { GetEvmOpts } from "./PollEvm";
import { EvmLog } from "../../entities";
import winston from "winston";

export class GetEvmLogs {
  private readonly blockRepo: EvmBlockRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: EvmBlockRepository) {
    this.blockRepo = blockRepo;
    this.logger = winston.child({ module: "GetEvmLogs" });
  }

  async execute(range: Range, opts: GetEvmOpts): Promise<EvmLog[]> {
    const { fromBlock, toBlock } = range;
    const chain = opts.chain;

    if (fromBlock > toBlock) {
      this.logger.info(
        `[${chain}][exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
      );
      return [];
    }

    this.logger.info(
      `[${chain}][exec] Processing blocks [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );

    const logs = await this.blockRepo.getFilteredLogs(chain, {
      fromBlock,
      toBlock,
      addresses: opts.filters[0].addresses ?? [], // At the moment, we only support one core contract per chain
      topics: opts.filters[0].topics?.flat() ?? [], // At the moment, we only support one topic per chain linked to the core contract
    });

    const blockNumbers = new Set(logs.map((log) => log.blockNumber));
    const blocks = await this.blockRepo.getBlocks(chain, blockNumbers, false);
    logs.forEach((log) => {
      const block = blocks[log.blockHash];
      if (!block) {
        this.logger.warn(`[${chain}][exec] Block not found for log [blockHash: ${log.blockHash}]`);
        throw new Error("Block not found");
      }
      log.blockTime = block.timestamp;
    });

    this.logger.info(
      `[${chain}][exec] Got ${logs.length} logs to process [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );
    return logs;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
