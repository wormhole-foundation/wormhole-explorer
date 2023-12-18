import { EvmBlockRepository } from "../../repositories";
import { EvmTransactions } from "../../entities";
import winston from "winston";
import { methodNameByAddressMapper } from "./mappers/methodNameByAddressMapper";

export class GetEvmTransactions {
  private readonly blockRepo: EvmBlockRepository;
  protected readonly logger: winston.Logger;

  constructor(blockRepo: EvmBlockRepository) {
    this.logger = winston.child({ module: "GetEvmTransactions" });
    this.blockRepo = blockRepo;
  }

  async execute(range: Range, opts: GetEvmTransactionsOpts): Promise<EvmTransactions[]> {
    const transactionsUpdated: EvmTransactions[] = [];
    const environment = opts.environment;
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
      const { transactions = [] } = await this.blockRepo.getBlock(
        chain,
        block,
        isTransactionsPresent
      );

      // Only process transactions to the contract address
      const transactionsFilter = transactions.filter((transaction) =>
        opts.addresses.includes(String(transaction.to).toLowerCase())
      );

      if (transactionsFilter.length > 0) {
        await this.populateTransaction(chain, environment, transactionsFilter, transactionsUpdated);
      }
    }

    return transactionsUpdated;
  }

  private async populateTransaction(
    chain: string,
    environment: string,
    transactionsFilter: EvmTransactions[],
    transactionsUpdated: EvmTransactions[]
  ): Promise<void> {
    await Promise.all(
      transactionsFilter.map(async (transaction) => {
        const status = await this.blockRepo.getTransactionReceipt(chain, transaction.hash);
        const methodsByAddress = methodNameByAddressMapper(chain, environment, transaction);

        transactionsUpdated.push({
          ...transaction,
          methodsByAddress,
          status,
        });
      })
    );
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};

type GetEvmTransactionsOpts = {
  addresses: string[];
  topics: string[];
  chain: string;
  environment: string;
};
