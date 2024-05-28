import { EvmBlock, EvmTransaction, ReceiptTransaction } from "../../../entities";
import { GetTransactions, Filter } from "../GetEvmTransactions";
import { EvmBlockRepository } from "../../../repositories";
import { GetEvmOpts } from "../PollEvm";

const TOPICS_APPLY = ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"];

export class NFTTransactions implements GetTransactions {
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

  apply(topics: string[]): boolean {
    return topics.some((topic) => TOPICS_APPLY.includes(topic));
  }

  async execute(filter: Filter): Promise<EvmTransaction[]> {
    let populatedTransactions: EvmTransaction[] = [];
    const isTransactionsPresent = true;
    const blockNumbers: Set<bigint> = new Set();

    for (let block = this.fromBlock; block <= this.toBlock; block++) {
      blockNumbers.add(block);
    }
    const evmBlocks = await this.blockRepo.getBlocks(
      this.chain,
      blockNumbers,
      isTransactionsPresent
    );

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
        const hashNumbers = new Set(transactionsByAddressConfigured.map((tx) => tx.hash));
        const receiptTransactions = await this.blockRepo.getTransactionReceipt(
          this.chain,
          hashNumbers
        );

        await this.populateTransaction(
          this.opts,
          evmBlock,
          receiptTransactions,
          transactionsByAddressConfigured,
          populatedTransactions
        );
      }
    }

    return populatedTransactions;
  }

  private async populateTransaction(
    opts: GetEvmOpts,
    evmBlock: EvmBlock,
    receiptTransactions: Record<string, ReceiptTransaction>,
    filterTransactions: EvmTransaction[],
    populatedTransactions: EvmTransaction[]
  ) {
    filterTransactions.forEach((transaction) => {
      transaction.status = receiptTransactions[transaction.hash].status;
      transaction.timestamp = evmBlock.timestamp;
      transaction.environment = opts.environment;
      transaction.chainId = opts.chainId;
      transaction.chain = opts.chain;
      transaction.logs = receiptTransactions[transaction.hash].logs;
      populatedTransactions.push(transaction);
    });
  }
}
