import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { AlgorandTransaction } from "../../../domain/entities/algorand";
import { AlgorandRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import winston from "winston";

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;

let TRANSACTIONS_ENDPOINT = "/v2/transactions";
let STATUS_ENDPOINT = "/v2/status";

export class AlgorandJsonRPCBlockRepository implements AlgorandRepository {
  private readonly logger: winston.Logger;
  protected algoV2Pools: ProviderPoolMap;
  protected algoIndexerPools: ProviderPoolMap;

  constructor(
    algoV2Pools: ProviderPool<InstrumentedHttpProvider>,
    algoIndexerPools: ProviderPool<InstrumentedHttpProvider>
  ) {
    this.logger = winston.child({ module: "AlgorandJsonRPCBlockRepository" });
    this.algoV2Pools = algoV2Pools;
    this.algoIndexerPools = algoIndexerPools;
  }

  async getBlockHeight(): Promise<bigint | undefined> {
    let results: ResultStatus;
    results = await this.algoV2Pools.get().get<typeof results>(STATUS_ENDPOINT);
    return BigInt(results["last-round"]);
  }

  async getTransactions(
    applicationId: string,
    fromBlock: bigint,
    toBlock: bigint
  ): Promise<AlgorandTransaction[]> {
    try {
      let results: ResultTransactions;
      results = await this.algoIndexerPools
        .get()
        .get<typeof results>(
          `${TRANSACTIONS_ENDPOINT}?application-id=${Number(
            applicationId
          )}&min-round=${fromBlock}&max-round=${toBlock}`
        );

      return results.transactions.map((tx) => {
        return {
          payload: tx["application-transaction"]?.["application-args"][1],
          applicationId: tx["application-transaction"]["application-id"],
          blockNumber: tx["confirmed-round"],
          timestamp: tx["round-time"],
          innerTxs: tx["inner-txns"],
          sender: tx.sender,
          hash: tx.id,
        };
      });
    } catch (e) {
      this.handleError(
        `Application id: ${applicationId} and range params: ${fromBlock} - ${toBlock}, error: ${e}`,
        "getTransactions"
      );
      throw e;
    }
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[algorand] Error calling ${method}: ${e.message ?? e}`);
  }
}

type ResultStatus = {
  "last-round": number;
};

type ResultTransactions = {
  "current-round": number;
  "next-token": string;
  transactions: {
    "tx-type": string;
    "application-transaction": {
      "application-id": string;
      "application-args": string[];
    };
    id: string;
    sender: string;
    "confirmed-round": number;
    "application-args": any;
    "round-time": number;
    logs: string[];
    "inner-txns": {
      sender: string;
      logs: string[];
    }[];
  }[];
};
