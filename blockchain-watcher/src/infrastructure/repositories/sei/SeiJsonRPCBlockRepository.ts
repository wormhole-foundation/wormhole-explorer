import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { SeiRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import { SeiRedeem } from "../../../domain/entities/sei";
import winston from "winston";

const TRANSACTION_SEARCH_ENDPOINT = "/tx_search";
const BLOCK_ENDPOINT = "/block";
const ACTION = "complete_transfer_with_payload";

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;

export class SeiJsonRPCBlockRepository implements SeiRepository {
  private readonly logger: winston.Logger;
  protected pool: ProviderPoolMap;

  constructor(pool: ProviderPool<InstrumentedHttpProvider>) {
    this.logger = winston.child({ module: "SeiJsonRPCBlockRepository" });
    this.pool = pool;
  }

  async getRedeems(chainId: number, address: string, blockBatchSize: number): Promise<SeiRedeem[]> {
    try {
      let resultTransactionSearch: ResultTransactionSearch | undefined;
      const query = `wasm._contract_address='${address}' AND wasm.action='${ACTION}'`;

      const perPageLimit = 20;
      const seiRedeems = [];
      let continuesFetching = true;
      let page = 1;

      while (continuesFetching) {
        try {
          resultTransactionSearch = await this.pool
            .get()
            .get<typeof resultTransactionSearch>(
              `${TRANSACTION_SEARCH_ENDPOINT}?query=${query}&page=${page}&per_page=${perPageLimit}`
            );

          if (resultTransactionSearch?.txs) {
            seiRedeems.push(...resultTransactionSearch.txs);
          }

          const totalCount = page * perPageLimit;
          if (totalCount >= blockBatchSize) {
            continuesFetching = false;
          }
          page++;
        } catch (e) {
          this.handleError(
            `[sei] Get transaction error: ${e} with query \n${query}\n`,
            "getRedeems"
          );
          continuesFetching = false;
        }
      }

      if (!seiRedeems) {
        this.logger.warn(`[getRedeems] Do not find any transaction with query \n${query}\n`);
        return [];
      }

      const sortedSeiRedeems = seiRedeems.sort((a, b) => Number(a.height) - Number(b.height));
      return sortedSeiRedeems.map((tx) => {
        return {
          chainId: chainId,
          events: tx.tx_result.events,
          height: BigInt(tx.height),
          data: tx.tx_result.data,
          hash: tx.hash,
          tx: Buffer.from(tx.tx, "base64"),
        };
      });
    } catch (e) {
      this.handleError(`Error: ${e}`, "getRedeems");
      throw e;
    }
  }

  async getBlockTimestamp(blockNumber: bigint): Promise<number | undefined> {
    try {
      const blockEndpoint = `${BLOCK_ENDPOINT}?height=${blockNumber}`;
      let resultsBlock: ResultBlock;

      resultsBlock = await this.pool.get().get<typeof resultsBlock>(blockEndpoint);

      if (
        !resultsBlock ||
        !resultsBlock.block ||
        !resultsBlock.block.header ||
        !resultsBlock.block.header.time
      ) {
        return undefined;
      }

      const dateTime: Date = new Date(resultsBlock.block.header.time);
      const timestamp: number = dateTime.getTime();

      return timestamp;
    } catch (e) {
      this.handleError(`Error: ${e}`, "getBlockTimestamp");
      throw e;
    }
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[sei] Error calling ${method}: ${e.message ?? e}`);
  }
}

type ResultTransactionSearch = {
  txs: [
    {
      height: string;
      hash: string;
      tx_result: {
        height: string;
        txhash: string;
        codespace: string;
        code: 0;
        data: string;
        raw_log: string;
        logs: [{ msg_index: number; log: string; events: EventsType }];
        info: string;
        gas_wanted: string;
        gas_used: string;
        tx: {
          "@type": "/cosmos.tx.v1beta1.Tx";
          body: {
            messages: [
              {
                "@type": "/cosmos.staking.v1beta1.MsgBeginRedelegate";
                delegator_address: string;
                validator_src_address: string;
                validator_dst_address: string;
                amount: { denom: string; amount: string };
              }
            ];
            memo: "";
            timeout_height: "0";
            extension_options: [];
            non_critical_extension_options: [];
          };
          auth_info: {
            signer_infos: [
              {
                public_key: {
                  "@type": "/cosmos.crypto.secp256k1.PubKey";
                  key: string;
                };
                mode_info: { single: { mode: string } };
                sequence: string;
              }
            ];
            fee: {
              amount: [{ denom: string; amount: string }];
              gas_limit: string;
              payer: string;
              granter: string;
            };
          };
          signatures: string[];
        };
        timestamp: string; // eg. '2023-01-03T12:12:54Z'
        events: EventsType[];
      };
      tx: string;
    }
  ];
};

type EventsType = {
  type: string;
  attributes: [
    {
      key: string;
      value: string;
      index: boolean;
    }
  ];
};

type ResultBlock = {
  block_id: {
    hash: string;
    parts: {
      total: number;
      hash: string;
    };
  };
  block: {
    header: {
      version: { block: string };
      chain_id: string;
      height: string;
      time: string; // eg. '2023-01-03T12:13:00.849094631Z'
      last_block_id: { hash: string; parts: { total: number; hash: string } };
      last_commit_hash: string;
      data_hash: string;
      validators_hash: string;
      next_validators_hash: string;
      consensus_hash: string;
      app_hash: string;
      last_results_hash: string;
      evidence_hash: string;
      proposer_address: string;
    };
    data: { txs: string[] | null };
    evidence: { evidence: null };
    last_commit: {
      height: string;
      round: number;
      block_id: { hash: string; parts: { total: number; hash: string } };
      signatures: string[];
    };
  };
};
