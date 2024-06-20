import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { AlgorandRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import winston from "winston";

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;

let APPLICATIONS_LOGS_ENDPOINT = "/v2/applications";
let STATUS_ENDPOINT = "/v2/status";

export class AlgorandJsonRPCBlockRepository implements AlgorandRepository {
  private readonly logger: winston.Logger;
  protected pool: ProviderPoolMap;

  constructor(pool: ProviderPool<InstrumentedHttpProvider>) {
    this.logger = winston.child({ module: "AlgorandJsonRPCBlockRepository" });
    this.pool = pool;
  }

  async getBlockHeight(): Promise<bigint | undefined> {
    let results: ResultStatus;

    results = await this.pool.get().get<typeof results>(STATUS_ENDPOINT);
    return BigInt(results["last-round"]);
  }

  async getApplicationsLogs(address: string, fromBlock: bigint, toBlock: bigint): Promise<any[]> {
    let results: ResultApplicationsLogs;

    results = await this.pool
      .get()
      .get<typeof results>(
        `${APPLICATIONS_LOGS_ENDPOINT}/${Number(
          address
        )}/logs?"min-round"=${fromBlock}&"max-round"=${toBlock}`
      );

    if (results) {
      results = await this.pool
        .get()
        .get<typeof results>(
          `${APPLICATIONS_LOGS_ENDPOINT}/${Number(address)}/logs?next=${results["next-token"]}`
        );
    }
    return [];
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[algorand] Error calling ${method}: ${e.message ?? e}`);
  }
}

type ResultStatus = {
  "last-round": number;
};

type ResultApplicationsLogs = {
  "application-id": string;
  "current-round": number;
  "next-token": string;
};
