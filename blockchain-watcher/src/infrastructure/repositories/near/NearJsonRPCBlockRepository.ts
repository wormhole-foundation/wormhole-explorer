import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { FinalExecutionOutcome } from "near-api-js/lib/providers/provider";
import { HttpClientError } from "../../errors/HttpClientError";
import { NearTransaction } from "../../../domain/entities/near";
import { NearRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import winston from "winston";

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;

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
        id: "",
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
    const nearTransactions: NearTransaction[] = [];

    try {
      const chunksTransactions: ChunkTransaction[] = [];

      for (let block = fromBlock; block <= toBlock; block++) {
        const responseBlock = await this.getBlockById(block);

        for (const chunk of responseBlock.result.chunks) {
          const responseTx = await this.getChunk(chunk.chunk_hash);
          chunksTransactions.push(responseTx.result.transactions);
        }

        const transactions = chunksTransactions.flatMap((transactions) => transactions);
        for (const tx of transactions) {
          if (tx.receiver_id === contract) {
            const outcome = await this.getTxStatus(contract, tx.hash);

            const logs = outcome.receipts_outcome.filter(({ outcome }) => {
              return (outcome as any).executor_id === contract;
            });
            nearTransactions.push({
              receiverId: tx.receiver_id,
              timestamp: Math.floor(responseBlock.result.header.timestamp / 1_000_000),
              actions: tx.actions.map((action: any) => {
                return {
                  functionCall: {
                    method: action.FunctionCall.method_name,
                    args: action.FunctionCall.args,
                  },
                };
              }),
              height: BigInt(responseBlock.result.header.height),
              hash: tx.hash,
              logs,
            });
          }
        }
      }
    } catch (e: HttpClientError | any) {
      this.handleError(e, "getTransactions");
      throw e;
    }
    return nearTransactions; // TODP: Duplicated maybe
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
  actions: any[];
  nonce: number;
  hash: string;
}[];

type Chunk = {
  chunk_hash: string;
};
