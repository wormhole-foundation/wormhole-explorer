import axios from 'axios';
import { connect } from 'near-api-js';
import { Provider } from 'near-api-js/lib/providers';
import { AXIOS_CONFIG_JSON } from '../consts';
import {
  EventLog,
  GetTransactionsByAccountIdRequestParams,
  GetTransactionsByAccountIdResponse,
  Transaction,
  WormholePublishEventLog,
} from '../types/near';

// The following is obtained by going to: https://explorer.near.org/accounts/contract.wormhole_crypto.near
// and watching the network tab in the browser to see where the explorer is going.
const NEAR_EXPLORER_TRANSACTION_URL =
  'https://explorer-backend-mainnet-prod-24ktefolwq-uc.a.run.app/trpc/transaction.listByAccountId';
export const NEAR_ARCHIVE_RPC = 'https://archival-rpc.mainnet.near.org';

export const getNearProvider = async (rpc: string): Promise<Provider> => {
  const connection = await connect({ nodeUrl: rpc, networkId: 'mainnet' });
  const provider = connection.connection.provider;
  return provider;
};

export const getTransactionsByAccountId = async (
  accountId: string,
  batchSize: number,
  timestamp: string
): Promise<Transaction[]> => {
  const params: GetTransactionsByAccountIdRequestParams = {
    accountId,
    limit: batchSize,
    cursor: {
      timestamp,
      indexInChunk: 0,
    },
  };

  // using this api: https://github.com/near/near-explorer/blob/beead42ba2a91ad8d2ac3323c29b1148186eec98/backend/src/router/transaction/list.ts#L127
  const res = (
    (
      await axios.get(
        `${NEAR_EXPLORER_TRANSACTION_URL}?batch=1&input={"0":${JSON.stringify(params)}}`,
        AXIOS_CONFIG_JSON
      )
    ).data as GetTransactionsByAccountIdResponse
  )[0];
  if ('error' in res) throw new Error(res.error.message);
  return res.result.data.items
    .filter(
      (tx) => tx.status === 'success' && tx.actions.some((a) => a.kind === 'functionCall') // other actions don't generate logs
    )
    .reverse(); // return chronological order
};

export const isWormholePublishEventLog = (log: EventLog): log is WormholePublishEventLog => {
  return log.standard === 'wormhole' && log.event === 'publish';
};
