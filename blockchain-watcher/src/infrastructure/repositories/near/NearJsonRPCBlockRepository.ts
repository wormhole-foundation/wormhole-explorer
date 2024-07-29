import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { NearTransaction } from "../../../domain/entities/near";
import { FinalExecutionOutcome } from "near-api-js/lib/providers/provider";
import { NearRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import winston from "winston";
import { HttpClientError } from "../../errors/HttpClientError";

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;

let STATUS_ENDPOINT = "/v2/status";

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

  async getTransactions(contract: string, fromBlock: bigint, toBlock: bigint): Promise<any[]> {
    const chunks: any[] = [];

    try {
      for (let block = fromBlock; block <= toBlock; block++) {
        const responseBlock = await this.getBlockById(block);

        for (const chunk of responseBlock.result.chunks) {
          const responseTx = await this.getChunk(chunk.chunk_hash);
          chunks.push(responseTx.result.transactions);
        }

        const transactions = chunks.flatMap((transactions) => transactions);
        for (const tx of transactions) {
          // Remove this if is oly to test
          if (tx.hash === "DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj") {
            // Test with this tx: https://nearblocks.io/txns/DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj?tab=execution
            console.log(tx);
          }

          const outcome = await this.getTxStatus(contract, tx.hash);

          const logs = outcome.receipts_outcome
            .filter(({ outcome }) => {
              return (
                (outcome as any).executor_id === contract
                //(outcome.status as ExecutionStatus).SuccessValue
              );
            })
            .flatMap(({ outcome }) => outcome.logs)
            //.filter((log) => log.startsWith("EVENT_JSON:")) // https://nomicon.io/Standards/EventsFormat
            .map((log) => JSON.parse(log.slice(11)) as EventLog)
            .filter(this.isWormholePublishEventLog);

          console.log(logs);
        }
      }
      console.log(chunks);
    } catch (e: HttpClientError | any) {
      this.handleError(e, "getBlockHeight");
      throw e;
    }

    throw new Error("Method not implemented.");
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
    let responseTx: { result: any };
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

  private isWormholePublishEventLog = (log: EventLog): log is any => {
    return log.standard === "wormhole";
  };
}

// https://nomicon.io/Standards/EventsFormat
type EventLog = {
  event: string;
  standard: string;
  data?: unknown;
  version?: string; // this is supposed to exist but is missing in WH logs
};

interface BlockResult {
  header: BlockHeader;
  chunks: Chunk[];
}

interface BlockHeader {
  height: number;
}

interface Chunk {
  chunk_hash: string;
}
