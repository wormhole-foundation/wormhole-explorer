import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { ProviderPoolDecorator } from "../../rpc/http/ProviderPoolDecorator";
import { ProviderHealthCheck } from "../../../domain/poolRpcs/PoolRpcs";
import { AlgorandTransaction } from "../../../domain/entities/algorand";
import { AlgorandRepository } from "../../../domain/repositories";
import winston from "winston";

type ProviderPoolMap = ProviderPoolDecorator<InstrumentedHttpProvider>;

let TRANSACTIONS_ENDPOINT = "/v2/transactions";
let STATUS_ENDPOINT = "/v2/status";

export class AlgorandJsonRPCBlockRepository implements AlgorandRepository {
  private readonly logger: winston.Logger;
  protected algoV2Pools: ProviderPoolMap;
  protected algoIndexerPools: ProviderPoolMap;

  constructor(
    algoV2Pools: ProviderPoolDecorator<InstrumentedHttpProvider>,
    algoIndexerPools: ProviderPoolDecorator<InstrumentedHttpProvider>
  ) {
    this.logger = winston.child({ module: "AlgorandJsonRPCBlockRepository" });
    this.algoV2Pools = algoV2Pools;
    this.algoIndexerPools = algoIndexerPools;
  }

  async healthCheck(
    chain: string,
    finality: string,
    cursor: bigint
  ): Promise<ProviderHealthCheck[]> {
    const providers = this.algoV2Pools.getProviders();
    const providersHealthCheck: ProviderHealthCheck[] = [];
    let response: ResultStatus;

    for (const provider of providers) {
      const url = provider.getUrl();
      try {
        const requestStartTime = performance.now();
        response = await provider.get<typeof response>(STATUS_ENDPOINT);
        const requestEndTime = performance.now();

        const lastRound = response["last-round"] ? BigInt(response["last-round"]) : undefined;

        providersHealthCheck.push({
          isHealthy: lastRound !== undefined,
          latency: Number(((requestEndTime - requestStartTime) / 1000).toFixed(2)),
          height: lastRound,
          url: url,
        });
      } catch (e) {
        providersHealthCheck.push({ url: url, height: undefined, isHealthy: false });
      }
    }
    this.algoV2Pools.setProviders(chain, providers, providersHealthCheck, cursor);
    return providersHealthCheck;
  }

  async getBlockHeight(): Promise<bigint | undefined> {
    let result: ResultStatus;
    result = await this.algoV2Pools.get().get<typeof result>(STATUS_ENDPOINT);
    return BigInt(result["last-round"]);
  }

  async getTransactions(
    applicationId: string,
    fromBlock: bigint,
    toBlock: bigint
  ): Promise<AlgorandTransaction[]> {
    try {
      let result: ResultTransactions;
      result = await this.algoIndexerPools
        .get()
        .get<typeof result>(
          `${TRANSACTIONS_ENDPOINT}?application-id=${Number(
            applicationId
          )}&min-round=${fromBlock}&max-round=${toBlock}`
        );

      if (!result.transactions || result.transactions.length === 0) {
        return [];
      }

      return result.transactions.map((tx) => {
        return {
          payload: tx["application-transaction"]?.["application-args"][1],
          method: tx["application-transaction"]?.["application-args"][0],
          applicationId: tx["application-transaction"]["application-id"],
          blockNumber: tx["confirmed-round"],
          timestamp: tx["round-time"],
          innerTxs: tx["inner-txns"]?.map((innerTx) => {
            // build inner transactions
            return {
              applicationId: innerTx["application-transaction"]?.["application-id"],
              payload: innerTx["application-transaction"]?.["application-args"][1],
              method: innerTx["application-transaction"]?.["application-args"][0],
              sender: innerTx.sender,
              logs: innerTx.logs,
            };
          }),
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
    "application-args": string[];
    "round-time": number;
    logs: string[];
    "inner-txns": {
      sender: string;
      logs: string[];
      "application-transaction": {
        "application-id": string;
        "application-args": string[];
      };
    }[];
  }[];
};
