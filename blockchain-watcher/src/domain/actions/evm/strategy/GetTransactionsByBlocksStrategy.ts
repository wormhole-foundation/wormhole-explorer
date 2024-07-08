import { GetTransactions, Filter, populateTransaction } from "../GetEvmTransactions";
import { EvmBlockRepository } from "../../../repositories";
import { EvmTransaction } from "../../../entities";
import { GetEvmOpts } from "../PollEvm";

export class GetTransactionsByBlocksStrategy implements GetTransactions {
  private readonly blockRepo: EvmBlockRepository;

  constructor(blockRepo: EvmBlockRepository) {
    this.blockRepo = blockRepo;
  }

  appliesTo(strategy: string): boolean {
    return strategy == GetTransactionsByBlocksStrategy.name;
  }

  async execute(
    filter: Filter,
    fromBlock: bigint,
    toBlock: bigint,
    opts: GetEvmOpts
  ): Promise<EvmTransaction[]> {
    let populatedTransactions: EvmTransaction[] = [];
    const blockNumbers: Set<bigint> = new Set();
    const chain = opts.chain;

    for (let block = fromBlock; block <= toBlock; block++) {
      blockNumbers.add(block);
    }
    // Get blocks with your transactions
    const evmBlocks = await this.blockRepo.getBlocks(chain, blockNumbers, true);

    for (const blockKey in evmBlocks) {
      const evmBlock = evmBlocks[blockKey];
      const transactions = evmBlock.transactions ?? [];

      // Only process transactions to the contract address configured
      const transactionsByAddressConfigured = transactions.filter(
        (transaction) =>
          filter.addresses?.includes(String(transaction.from).toLowerCase()) ||
          filter.addresses?.includes(String(transaction.to).toLowerCase())
      );

      if (transactionsByAddressConfigured.length > 0) {
        // Get transaction details from blockchain
        const hashNumbers = new Set(transactionsByAddressConfigured.map((tx) => tx.hash));
        const transactionReceipts = await this.blockRepo.getTransactionReceipt(chain, hashNumbers);

        populateTransaction(
          opts,
          evmBlocks,
          transactionReceipts,
          transactionsByAddressConfigured,
          populatedTransactions
        );
      }
    }

    return populatedTransactions;
  }
}
