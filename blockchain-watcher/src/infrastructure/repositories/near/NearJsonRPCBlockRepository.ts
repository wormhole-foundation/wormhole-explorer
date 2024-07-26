import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { NearTransaction } from "../../../domain/entities/near";
import { NearRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import winston from "winston";

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;

let STATUS_ENDPOINT = "/v2/status";

export class NearJsonRPCBlockRepository implements NearRepository {
  private readonly logger: winston.Logger;
  protected pool: ProviderPoolMap;

  constructor(pool: ProviderPool<InstrumentedHttpProvider>) {
    this.logger = winston.child({ module: "NearJsonRPCBlockRepository" });
    this.pool = pool;
  }

  async getBlockHeight(): Promise<bigint | undefined> {
    let result: BlockResult;
    result = await this.pool.get().get<typeof result>(STATUS_ENDPOINT);
    return BigInt(result.header.height);
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[Near] Error calling ${method}: ${e.message ?? e}`);
  }
}

export interface BlockResult {
  header: BlockHeader;
}

export interface BlockHeader {
  height: number;
}
