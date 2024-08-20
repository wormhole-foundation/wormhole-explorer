import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { FinalExecutionOutcome } from "near-api-js/lib/providers/provider";
import { HttpClientError } from "../../errors/HttpClientError";
import { NearTransaction } from "../../../domain/entities/near";
import { NearRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import winston from "winston";

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;
const NEAR_CHAIN_ID = 15;

export class NearJsonRPCBlockRepository implements NearRepository {
  private readonly logger: winston.Logger;
  protected pool: ProviderPoolMap;

  constructor(pool: ProviderPool<InstrumentedHttpProvider>) {
    this.logger = winston.child({ module: "NearJsonRPCBlockRepository" });
    this.pool = pool;
  }

  async getBlockHeight(commitment: string): Promise<bigint | undefined> {
    let response: { result: BlockResult };

    try {
      response = await this.pool.get().post<typeof response>({
        jsonrpc: "2.0",
        id: "", // Is not used
        method: "block",
        params: {
          finality: commitment,
        },
      });
    } catch (e: HttpClientError | any) {
      this.handleError(e, "getBlockHeight");
      throw e;
    }

    return BigInt(response.result.header.height);
  }

  async getTransactions(
    contract: string,
    fromBlock: bigint,
    toBlock: bigint
  ): Promise<NearTransaction[]> {
    const chunksTransactions: ChunkTransaction[] = [];
    const uniqueTransaction = new Set<string>();
    const nearTransactions: NearTransaction[] = [];
    const blockPromises = [];

    try {
      for (let block = fromBlock; block <= toBlock; block++) {
        blockPromises.push(this.getBlockById(block));
      }
      const blocks = await Promise.all(blockPromises);

      for (const responseBlock of blocks) {
        if (!responseBlock || !responseBlock.result || !responseBlock.result.chunks) {
          continue;
        }

        const chunkPromises = responseBlock.result.chunks.map((chunk) =>
          this.getChunk(chunk.chunk_hash)
        );
        const chunks = await Promise.all(chunkPromises);

        for (const responseTx of chunks) {
          if (responseTx.result && responseTx.result.transactions) {
            chunksTransactions.push(responseTx.result.transactions);
          }
        }

        const transactions = chunksTransactions
          .flatMap((transactions) => transactions)
          .filter((tx) => tx.receiver_id === contract && !uniqueTransaction.has(tx.hash));

        if (!transactions || transactions.length == 0) {
          continue; // Skip block process if not contain wormhole transactions
        }

        for (const tx of transactions) {
          const outcome = await this.getTxStatus(contract, tx.hash);

          const logs = outcome.receipts_outcome.filter(({ outcome }) => {
            return (outcome as any).executor_id === contract;
          });
          nearTransactions.push({
            receiverId: tx.receiver_id, // Wormhole contract
            signerId: tx.signer_id, // Sender contract
            timestamp: Math.floor(responseBlock.result.header.timestamp / 1000000000), // Convert to seconds
            blockHeight: BigInt(responseBlock.result.header.height),
            chainId: NEAR_CHAIN_ID,
            hash: tx.hash,
            logs,
            actions: tx.actions.map((action: any) => {
              return {
                functionCall: {
                  method: action.FunctionCall.method_name,
                  args: action.FunctionCall.args,
                },
              };
            }),
          });
          uniqueTransaction.add(tx.hash); // Avoid duplicated transactions
        }
      }
    } catch (e: HttpClientError | any) {
      this.handleError(e, "getTransactions");
      throw e;
    }
    return nearTransactions;
  }

  async getBlockById(block: bigint) {
    let responseBlock: { result: BlockResult };
    return await this.pool.get().post<typeof responseBlock>({
      jsonrpc: "2.0",
      id: "", // Is not used
      method: "block",
      params: {
        block_id: Number(block),
      },
    });
  }

  async getChunk(chunkHash: string) {
    let responseTx: { result: ChunkResult };
    return await this.pool.get().post<typeof responseTx>({
      jsonrpc: "2.0",
      id: "", // Is not used
      method: "chunk",
      params: {
        chunk_id: chunkHash,
      },
    });
  }

  async getTxStatus(contract: string, hash: string): Promise<FinalExecutionOutcome> {
    let responseTx: { result: FinalExecutionOutcome };
    responseTx = await this.pool.get().post<typeof responseTx>({
      jsonrpc: "2.0",
      id: "", // Is not used
      method: "tx",
      params: {
        sender_account_id: contract,
        tx_hash: hash,
      },
    });

    return responseTx.result;
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[Near] Error calling ${method}: ${e.message ?? e}`);
  }
}

type BlockResult = {
  header: BlockHeader;
  chunks: Chunk[];
};

type BlockHeader = {
  timestamp: number;
  height: number;
};

type ChunkResult = {
  header: BlockHeader;
  transactions: ChunkTransaction;
};

type ChunkTransaction = {
  receiver_id: string;
  signer_id: string;
  actions: any[];
  nonce: number;
  hash: string;
}[];

type Chunk = {
  chunk_hash: string;
};
