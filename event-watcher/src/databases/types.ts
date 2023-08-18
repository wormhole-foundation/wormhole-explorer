import { ChainId } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { Row } from '@google-cloud/bigtable';
export type VaaLog = {
  id: string;
  chainId: number;
  chainName: string;
  emitter: string;
  sequence: number;
  txHash: string;
  sender: string;
  payload: any;
  blockNumber: number;
};
export type VaasByBlock = { [blockInfo: string]: string[] };
export type DB = { [chain in ChainId]?: VaasByBlock };
export type LastBlockByChain = { [chain in ChainId]?: string };
export type JSONArray = string;
export type BigtableMessagesRow = {
  key: string;
  data: {
    // column family
    info: {
      // columns
      timestamp?: { value: string; timestamp: string };
      txHash?: { value: string; timestamp: string };
      hasSignedVaa?: { value: number; timestamp: string };
    };
  };
};
export interface BigtableSignedVAAsRow {
  key: string;
  data: {
    // column family
    info: {
      // columns
      bytes: { value: Buffer; timestamp: string };
    };
  };
}
export interface BigtableVAAsByTxHashRow {
  key: string;
  data: {
    // column family
    info: {
      // columns
      vaaKeys: { value: JSONArray; timestamp: string };
    };
  };
}
export interface BigtableMessagesResultRow extends Row {
  key: string;
  data: {
    // column family
    info: {
      // columns
      timestamp?: [{ value: string; timestamp: string }];
      txHash?: [{ value: string; timestamp: string }];
      hasSignedVaa?: [{ value: number; timestamp: string }];
    };
  };
}
export interface BigtableSignedVAAsResultRow extends Row {
  key: string;
  data: {
    // column family
    info: {
      // columns
      bytes: [{ value: Buffer; timestamp: string }];
    };
  };
}
