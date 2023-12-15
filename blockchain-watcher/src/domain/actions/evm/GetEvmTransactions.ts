import { EvmTransactions } from "../../entities";
import { EvmBlockRepository } from "../../repositories";
import winston from "winston";

export class GetEvmTransactions {
  private readonly blockRepo: EvmBlockRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: EvmBlockRepository) {
    this.blockRepo = blockRepo;
    this.logger = winston.child({ module: "GetEvmTransactions" });
  }

  async execute(range: Range, opts: GetEvmTransactionsOpts): Promise<EvmTransactions[]> {
    const transactionsUpdated: EvmTransactions[] = [];
    const fromBlock = range.fromBlock;
    const toBlock = range.toBlock;
    const chain = opts.chain;

    if (fromBlock > toBlock) {
      this.logger.info(`[exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`);
      return [];
    }

    const isTransactionsPresent = true;
    for (let block = fromBlock; block <= toBlock; block++) {
      // Get the transactions for the block
      const transactions: EvmTransactions[] | undefined = (
        await this.blockRepo.getBlock(chain, block, isTransactionsPresent)
      ).transactions;

      // Only process transactions to the contract address
      const transactionsFilter = transactions?.filter((transaction) =>
        opts.addresses?.includes(transaction.to)
      );

      transactionsFilter?.forEach(async transaction => {
        const status = await this.blockRepo.getTransactionReceipt(chain, transaction.hash);
        transaction.status = status;
        transactionsUpdated.push(transaction);
      });
    }

    return transactionsUpdated;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};

export interface GetEvmTransactionsOpts {
  addresses?: string[];
  topics?: string[];
  chain: string;
}
