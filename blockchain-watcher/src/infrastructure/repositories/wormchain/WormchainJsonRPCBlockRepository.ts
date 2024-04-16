import { divideIntoBatches, hexToHash } from "../common/utils";
import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { WormchainRepository } from "../../../domain/repositories";
import { WormchainBlockLogs } from "../../../domain/entities/wormchain";
import { ProviderPool } from "@xlabs/rpc-pool";
import winston from "winston";

let BLOCK_HEIGHT_ENDPOINT = "/abci_info";
let TRANSACTION_ENDPOINT = "/tx";
let BLOCK_ENDPOINT = "/block";

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;

export class WormchainJsonRPCBlockRepository implements WormchainRepository {
  private readonly logger: winston.Logger;
  protected pool: ProviderPoolMap;

  constructor(pool: ProviderPool<InstrumentedHttpProvider>) {
    this.logger = winston.child({ module: "WormchainJsonRPCBlockRepository" });
    this.pool = pool;
  }

  async getBlockHeight(): Promise<bigint | undefined> {
    try {
      let results: ResultBlockHeight;

      results = await this.pool.get().get<typeof results>(BLOCK_HEIGHT_ENDPOINT);

      if (
        results &&
        results.result &&
        results.result.response &&
        results.result.response.last_block_height
      ) {
        const blockHeight = results.result.response.last_block_height;
        return BigInt(blockHeight);
      }
      return undefined;
    } catch (e) {
      this.handleError(`Error: ${e}`, "getBlockHeight");
      throw e;
    }
  }

  async getBlockLogs(chainId: number, blockNumber: bigint): Promise<WormchainBlockLogs> {
    try {
      const blockEndpoint = `${BLOCK_ENDPOINT}?height=${blockNumber}`;
      let resultsBlock: ResultBlock;

      resultsBlock = await this.pool.get().get<typeof resultsBlock>(blockEndpoint);
      const txs = resultsBlock.result.block.data.txs;

      if (!txs) {
        return {
          transactions: [],
          blockHeight: BigInt(resultsBlock.result.block.header.height),
          timestamp: Number(resultsBlock.result.block.header.time),
          chainId,
        };
      }

      const cosmosTransaction: CosmosTransaction[] = [];
      const hashNumbers = new Set(txs.map((tx) => tx));
      const batches = divideIntoBatches(hashNumbers, 10);

      for (const batch of batches) {
        for (let hashBatch of batch) {
          const hash: string = hexToHash(hashBatch);
          const txEndpoint = `${TRANSACTION_ENDPOINT}?hash=0x${hash}`;

          const resultTransaction: ResultTransaction = await this.pool
            .get()
            .get<typeof resultTransaction>(txEndpoint);

          if (
            resultTransaction &&
            resultTransaction.result.tx_result &&
            resultTransaction.result.tx_result.events
          ) {
            resultTransaction.result.tx_result.events.forEach((event) => {
              if (event.type === "wasm") {
                cosmosTransaction.push({
                  hash: `0x${hash}`.toLocaleLowerCase(),
                  type: event.type,
                  attributes: event.attributes,
                });
              }
            });
          }
        }
      }

      const dateTime: Date = new Date(resultsBlock.result.block.header.time);
      const timestamp: number = dateTime.getTime();

      return {
        transactions: cosmosTransaction || [],
        blockHeight: BigInt(resultsBlock.result.block.header.height),
        timestamp: timestamp,
        chainId,
      };
    } catch (e) {
      this.handleError(`Error: ${e}`, "getBlockHeight");
      throw e;
    }
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[wormchain] Error calling ${method}: ${e.message ?? e}`);
  }
}

type ResultBlockHeight = { result: { response: { last_block_height: string } } };

type ResultBlock = {
  result: {
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
};

type ResultTransaction = {
  result: {
    tx: {
      body: {
        messages: string[];
        memo: string;
        timeout_height: string;
        extension_options: [];
        non_critical_extension_options: [];
      };
      auth_info: {
        signer_infos: string[];
        fee: {
          amount: [{ denom: string; amount: string }];
          gas_limit: string;
          payer: string;
          granter: string;
        };
      };
      signatures: string[];
    };
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
  };
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

type CosmosTransaction = {
  hash: string;
  type: string;
  attributes: {
    key: string;
    value: string;
    index: boolean;
  }[];
};
