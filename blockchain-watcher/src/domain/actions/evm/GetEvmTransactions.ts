import { EvmBlock, EvmTransaction, ReceiptTransaction } from "../../entities";
import { EvmBlockRepository } from "../../repositories";
import { GetEvmOpts } from "./PollEvm";
import winston from "winston";
import { GetTransactionsStrategy } from "./strategy/GetTransactionsStrategy";

export class GetEvmTransactions {
  private readonly blockRepo: EvmBlockRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: EvmBlockRepository) {
    this.logger = winston.child({ module: "GetEvmTransactions" });
    this.blockRepo = blockRepo;
  }

  async execute(range: Range, opts: GetEvmOpts): Promise<EvmTransaction[]> {
    const fromBlock = range.fromBlock;
    const toBlock = range.toBlock;
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

    const populatedTransactions = await new GetTransactionsStrategy(
      this.blockRepo,
      fromBlock,
      toBlock,
      chain,
      opts
    ).execute();

    this.logger.info(
      `[${chain}][exec] Got ${populatedTransactions?.length} transactions to process [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`
    );

    return populatedTransactions;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
