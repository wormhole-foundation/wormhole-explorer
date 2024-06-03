import { GetTransactions, Filter, populateTransaction } from "../GetEvmTransactions";
import { EvmBlockRepository } from "../../../repositories";
import { EvmTransaction } from "../../../entities";
import { GetEvmOpts } from "../PollEvm";

export class GetTransactionsByBlocksStrategy implements GetTransactions {
  private readonly blockRepo: EvmBlockRepository;
  private readonly fromBlock: bigint;
  private readonly toBlock: bigint;
  private readonly chain: string;
  private readonly opts: GetEvmOpts;

  constructor(
    blockRepo: EvmBlockRepository,
    fromBlock: bigint,
    toBlock: bigint,
    chain: string,
    opts: GetEvmOpts
  ) {
    this.blockRepo = blockRepo;
    this.fromBlock = fromBlock;
    this.toBlock = toBlock;
    this.chain = chain;
    this.opts = opts;
  }

  appliesTo(strategy: string): boolean {
    return strategy == GetTransactionsByBlocksStrategy.name;
  }

  async execute(filter: Filter): Promise<EvmTransaction[]> {
    let populatedTransactions: EvmTransaction[] = [];
    const blockNumbers: Set<bigint> = new Set();

    for (let block = this.fromBlock; block <= this.toBlock; block++) {
      blockNumbers.add(block);
    }
    // Get blocks with your transactions
    const evmBlocks = await this.blockRepo.getBlocks(this.chain, blockNumbers, true);

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
        const transactionReceipts = await this.blockRepo.getTransactionReceipt(
          this.chain,
          hashNumbers
        );

        populateTransaction(
          this.opts,
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
