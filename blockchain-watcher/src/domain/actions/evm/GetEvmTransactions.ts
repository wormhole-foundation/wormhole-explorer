import { EvmBlockRepository } from "../../repositories";
import { EvmTransactions } from "../../entities";
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
      const { transactions = [] } = await this.blockRepo.getBlock(chain, block, isTransactionsPresent);

      // Only process transactions to the contract address
      const transactionsFilter = transactions.filter((transaction) =>
        opts.addresses.includes(String(transaction.to).toLowerCase())
      );

      if (transactionsFilter.length > 0) {
        await this.populateTransaction(chain, transactionsFilter, transactionsUpdated);
      }
    }

    return transactionsUpdated;
  }

  private populateTransaction(chain: string, transactionsFilter: EvmTransactions[], transactionsUpdated: EvmTransactions[]): void {
    transactionsFilter.forEach(async (transaction) => {
      const status = await this.blockRepo.getTransactionReceipt(chain, transaction.hash);
      transaction.methodsByAddress = transaction.to;
      transaction.status = status;
      transactionsUpdated.push(transaction);
    });
  }

  private mappedMethodName() {
    const methodsByAddress: MethodsByAddress = {
      [String("0x3ee18B2214AFF97000D974cf647E7C347E8fa585").toLowerCase()]: [
        {
          ID: "0xc6878519",
          Name: "MethodCompleteTransfer",
        },
      ],
      [String("0xcafd2f0a35a4459fa40c0517e17e6fa2939441ca").toLowerCase()]: [
        {
          ID: "MetehodIDCompleteTransferWithRelay",
          Name: "MetehodCompleteTransferWithRelay",
        },
      ],
      [String("0xd8E1465908103eD5fd28e381920575fb09beb264").toLowerCase()]: [
        {
          ID: "MethodIDReceiveMessageAndSwap",
          Name: "MethodReceiveMessageAndSwap",
        },
      ],
    };
  }
}

type Range = {
  fromBlock: bigint;
  toBlock: bigint;
};

export interface GetEvmTransactionsOpts {
  addresses: string[];
  topics: string[];
  chain: string;
}

// ----------------------------------------------------------------------------------------------------------------------------------------------------

type BlockchainMethod = {
  ID: string;
  Name: string;
};

type MethodsByAddress = {
  [address: string]: BlockchainMethod[];
};

// Method ids for wormhole token bridge contract
export enum MethodID {
  MethodIDCompleteTransfer = "0xc6878519",
  MethodIDWrapAndTransfer = "0x9981509f",
  MethodIDTransferTokens = "0x0f5287b0",
  MethodIDAttestToken = "0xc48fa115",
  MethodIDCompleteAndUnwrapETH = "0xff200cde",
  MethodIDCreateWrapped = "0xe8059810",
  MethodIDUpdateWrapped = "0xf768441f",
}
