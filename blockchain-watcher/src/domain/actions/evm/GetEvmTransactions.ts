import { methodNameByAddressMapper } from "./mappers/methodNameByAddressMapper";
import { EvmBlock, EvmTransactions } from "../../entities";
import { EvmBlockRepository } from "../../repositories";
import { GetEvmOpts } from "./GetEvmLogs";
import winston from "winston";

export class GetEvmTransactions {
  private readonly blockRepo: EvmBlockRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: EvmBlockRepository) {
    this.logger = winston.child({ module: "GetEvmTransactions" });
    this.blockRepo = blockRepo;
  }

  async execute(range: Range, opts: GetEvmOpts): Promise<EvmTransactions[]> {
    const fromBlock = range.fromBlock;
    const toBlock = range.toBlock;

    if (fromBlock > toBlock) {
      this.logger.info(`[exec] Invalid range [fromBlock: ${fromBlock} - toBlock: ${toBlock}]`);
      return [];
    }

    const transactionsUpdated: EvmTransactions[] = [];
    const environment = opts.environment;
    const isTransactionsPresent = true;
    const chain = opts.chain;

    for (let block = fromBlock; block <= toBlock; block++) {
      const evmBlock = await this.blockRepo.getBlock(chain, block, isTransactionsPresent);
      const transactions = evmBlock.transactions ?? [];

      // Only process transactions to the contract address
      const transactionsFilter = transactions.filter((transaction) =>
        opts.addresses?.includes(String(transaction.to).toLowerCase())
      );

      if (transactionsFilter.length > 0) {
        await this.populateTransaction(
          chain,
          environment,
          evmBlock,
          transactionsFilter,
          transactionsUpdated
        );
      }
    }

    this.logger.info(
      `[${chain}][exec] Got ${
        transactionsUpdated?.length
      } transactions to process for ${this.populateLog(opts, fromBlock, toBlock)}`
    );
    return transactionsUpdated;
  }

  private async populateTransaction(
    chain: string,
    environment: string,
    evmBlock: EvmBlock,
    transactionsFilter: EvmTransactions[],
    transactionsUpdated: EvmTransactions[]
  ): Promise<void> {
    await Promise.all(
      transactionsFilter.map(async (transaction) => {
        const status = await this.blockRepo.getTransactionReceipt(chain, transaction.hash);
        const methodsByAddress = methodNameByAddressMapper(chain, environment, transaction);

        transactionsUpdated.push({
          ...transaction,
          chainId: Number(transaction.chainId),
          timestamp: evmBlock.timestamp,
          methodsByAddress,
          status,
        });
      })
    );
  }

  private populateLog(opts: GetEvmOpts, fromBlock: bigint, toBlock: bigint): string {
    return `[addresses:${opts.addresses}][topics:${opts.topics}][blocks:${fromBlock} - ${toBlock}]`;
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};
