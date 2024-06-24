import { AlgorandRepository } from "../../repositories";
import { GetAlgorandOpts } from "./PollAlgorand";
import winston from "winston";

export class GetAlgorandTransactions {
  private readonly blockRepo: AlgorandRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: AlgorandRepository) {
    this.logger = winston.child({ module: "GetAlgorandTransactions" });
    this.blockRepo = blockRepo;
  }

  async execute(range: Range, opts: GetAlgorandOpts): Promise<any[]> {
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

    const txs = await this.blockRepo.getTransactions(opts.applicationIds[0], fromBlock, toBlock);

    this.logger.info(
      `[${chain}][exec] Got ${txs?.length} transactions to process [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );
    return txs;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
