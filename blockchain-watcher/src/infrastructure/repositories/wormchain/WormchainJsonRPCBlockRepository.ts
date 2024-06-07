import { divideIntoBatches, hexToHash } from "../common/utils";
import { InstrumentedHttpProvider } from "../../rpc/http/InstrumentedHttpProvider";
import { WormchainRepository } from "../../../domain/repositories";
import { ProviderPool } from "@xlabs/rpc-pool";
import { setTimeout } from "timers/promises";
import winston from "winston";
import {
  WormchainBlockLogs,
  CosmosTransaction,
  IbcTransaction,
  CosmosRedeem,
} from "../../../domain/entities/wormchain";

let TRANSACTION_SEARCH_ENDPOINT = "/tx_search";
let BLOCK_HEIGHT_ENDPOINT = "/abci_info";
let TRANSACTION_ENDPOINT = "/tx";
let BLOCK_ENDPOINT = "/block";

const GROW_SLEEP_TIME = 350;
const MAX_ATTEMPTS = 10;

type ProviderPoolMap = ProviderPool<InstrumentedHttpProvider>;

export class WormchainJsonRPCBlockRepository implements WormchainRepository {
  private readonly logger: winston.Logger;
  protected cosmosPools: Map<number, ProviderPoolMap>;

  constructor(cosmosPools: Map<number, ProviderPoolMap>) {
    this.logger = winston.child({ module: "WormchainJsonRPCBlockRepository" });
    this.cosmosPools = cosmosPools;
  }

  async getBlockHeight(chainId: number): Promise<bigint | undefined> {
    try {
      const cosmosClient = this.cosmosPools.get(chainId);

      if (!cosmosClient) {
        this.logger.warn(`[wormchain] No cosmos client found for chain ${chainId}`);
        return undefined;
      }

      let results: ResultBlockHeight = await cosmosClient
        .get()
        .get<typeof results>(BLOCK_HEIGHT_ENDPOINT);
      // Cast result because some chains don't return the result property
      results = results.result ? results : { result: results as any };

      if (results && results.result.response && results.result.response.last_block_height) {
        const blockHeight = results.result.response.last_block_height;
        return BigInt(blockHeight);
      }
      return undefined;
    } catch (e) {
      this.handleError(`Error: ${e}`, "getBlockHeight");
      throw e;
    }
  }

  async getBlockLogs(
    chainId: number,
    blockNumber: bigint,
    attributesTypes: string[]
  ): Promise<WormchainBlockLogs> {
    try {
      const cosmosClient = this.cosmosPools.get(chainId)!;

      let block = await this.getBlockData(cosmosClient, blockNumber);
      const transactionHashes = block.data.txs;

      if (!transactionHashes || transactionHashes.length === 0) {
        return {
          transactions: [],
          blockHeight: BigInt(block.header.height),
          timestamp: Number(block.header.time),
        };
      }

      let transactionsData = await this.getTransactionsData(cosmosClient, transactionHashes);

      const cosmosTransactions: CosmosTransaction[] = [];
      for (let transaction of transactionsData) {
        // Cast result because some chains don't return the result property
        transaction = transaction.result ? transaction : { result: transaction as any };

        if (transaction && transaction.result.tx_result && transaction.result.tx_result.events) {
          const groupedAttributes: Attribute[] = [];

          // Group all attributes by tx hash
          transaction.result.tx_result.events
            .filter((event) => attributesTypes.includes(event.type))
            .forEach((event) => {
              event.attributes.forEach((attr) => {
                groupedAttributes.push(attr);
              });
            });

          if (groupedAttributes && groupedAttributes.length > 0) {
            const txToBase64 = Buffer.from(transaction.result.tx, "base64");

            cosmosTransactions.push({
              attributes: groupedAttributes,
              height: transaction.result.height,
              hash: `0x${transaction.result.hash}`.toLocaleLowerCase(),
              tx: txToBase64,
            });
          }
        }
      }

      return {
        transactions: cosmosTransactions || [],
        blockHeight: BigInt(block.header.height),
        timestamp: new Date(block.header.time).getTime(),
      };
    } catch (e) {
      this.handleError(`Error: ${e}`, "getBlockHeight");
      throw e;
    }
  }

  async getRedeems(ibcTransaction: IbcTransaction): Promise<CosmosRedeem[]> {
    try {
      const cosmosClient = this.cosmosPools.get(ibcTransaction.targetChain);

      if (!cosmosClient) {
        this.logger.warn(
          `[wormchain] No cosmos client found for chain ${ibcTransaction.targetChain}`
        );
        return [];
      }

      let resultTransactionSearch: ResultTransactionSearch | undefined;
      let isIBCTransferFinalized = false;
      let sleepTime = 300;
      let attempts = 1;

      const query = `"recv_packet.packet_sequence=${ibcTransaction.sequence} AND 
          recv_packet.packet_timeout_timestamp='${ibcTransaction.timestamp}' AND 
          recv_packet.packet_src_channel='${ibcTransaction.srcChannel}' AND 
          recv_packet.packet_dst_channel='${ibcTransaction.dstChannel}'"`;

      // The process to find the reedeem on target chain sometimes takes a while so we need to wait for it to be finalized before returning the data
      // we will try to get the transaction data every 300ms (and increasing) until it is finalized if it takes more than 10 attempts, we will throw an error
      while (!isIBCTransferFinalized && attempts <= MAX_ATTEMPTS) {
        try {
          await this.sleep(sleepTime);

          // Get cosmos transactions data
          resultTransactionSearch = await cosmosClient
            .get()
            .get<typeof resultTransactionSearch>(
              `${TRANSACTION_SEARCH_ENDPOINT}?query=${query}&prove=false&page=1&per_page=1`
            );

          if (
            resultTransactionSearch &&
            resultTransactionSearch.result &&
            resultTransactionSearch.result.txs &&
            resultTransactionSearch.result.txs.length > 0
          ) {
            isIBCTransferFinalized = true;
            break;
          }
        } catch (e) {
          this.handleError(
            `[${ibcTransaction.targetChain}] Get transaction error: ${e} with query \n${query}\n`,
            "getRedeems"
          );
        }

        this.logger.warn(
          `[getRedeems] Attempt ${attempts} to get transaction with chainId: ${ibcTransaction.targetChain}. Retrying in ${sleepTime}ms`
        );
        if (!isIBCTransferFinalized) {
          sleepTime += GROW_SLEEP_TIME;
          attempts++;
        }
      }

      if (!resultTransactionSearch || attempts > MAX_ATTEMPTS) {
        throw new Error(
          `[getRedeems] The transaction \n${query}\n with chainId: ${ibcTransaction.targetChain} never ended`
        );
      }

      return resultTransactionSearch.result.txs.map((tx) => {
        return {
          blockTimestamp: ibcTransaction.blockTimestamp,
          timestamp: ibcTransaction.timestamp,
          chainId: ibcTransaction.targetChain,
          events: tx.tx_result.events,
          height: tx.height,
          data: tx.tx_result.data,
          hash: tx.hash,
          tx: ibcTransaction.tx,
        };
      });
    } catch (e) {
      this.handleError(`Error: ${e}`, "getRedeems");
      throw e;
    }
  }

  /**
   * This method is used to get the block data from the blockchain
   */
  private async getBlockData(cosmosClient: ProviderPoolMap, blockNumber: bigint): Promise<Block> {
    const blockEndpoint = `${BLOCK_ENDPOINT}?height=${blockNumber}`;

    let blockResult: ResultBlock = await cosmosClient.get().get<typeof blockResult>(blockEndpoint);
    // Cast result because some chains don't return the result property
    blockResult = blockResult.result ? blockResult : { result: blockResult as any };
    return blockResult.result.block;
  }

  /**
   * This method is used to get the full transaction data from the blockchain
   */
  private async getTransactionsData(cosmosClient: ProviderPoolMap, transactionHashes: string[]) {
    let resultTransaction: ResultTransaction;
    const hashNumbers = new Set(transactionHashes.map((tx) => tx));
    const batches = divideIntoBatches(hashNumbers, 10);

    const resultTransactionPromises = batches.flatMap((batch) =>
      Array.from(batch).map((hashBatch) => {
        const hash: string = hexToHash(hashBatch);
        const txEndpoint = `${TRANSACTION_ENDPOINT}?hash=0x${hash}`;
        return cosmosClient.get().get<typeof resultTransaction>(txEndpoint);
      })
    );

    return await Promise.all(resultTransactionPromises);
  }

  private async sleep(sleepTime: number) {
    await setTimeout(sleepTime, null, { ref: false });
  }

  private handleError(e: any, method: string) {
    this.logger.error(`[wormchain] Error calling ${method}: ${e.message ?? e}`);
  }
}

type ResultBlockHeight = {
  result: {
    response: {
      last_block_height: string;
    };
  };
};

type ResultBlock = {
  result: {
    block_id: {
      hash: string;
      parts: {
        total: number;
        hash: string;
      };
    };
    block: Block;
  };
};

type Block = {
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

type ResultTransaction = {
  result: {
    height: string;
    hash: string;
    tx: string;
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

type ResultTransactionSearch = {
  result: {
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
      }
    ];
  };
};

type EventsType = {
  type: string;
  attributes: Attribute[];
};

type Attribute = {
  key: string;
  value: string;
  index: boolean;
};
