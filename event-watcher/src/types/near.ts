// https://nomicon.io/Standards/EventsFormat
export type EventLog = {
  event: string;
  standard: string;
  data?: unknown;
  version?: string; // this is supposed to exist but is missing in WH logs
};

export type WormholePublishEventLog = {
  standard: 'wormhole';
  event: 'publish';
  data: string;
  nonce: number;
  emitter: string;
  seq: number;
  block: number;
};

export type GetTransactionsByAccountIdResponse = [
  | {
      id: string | null;
      result: {
        type: string;
        data: {
          items: Transaction[];
        };
      };
    }
  | {
      id: string | null;
      error: {
        message: string;
        code: number;
        data: {
          code: string;
          httpStatus: number;
          path: string;
        };
      };
    }
];

export type Transaction = {
  hash: string;
  signerId: string;
  receiverId: string;
  blockHash: string;
  blockTimestamp: number;
  actions: Action[];
  status: 'unknown' | 'failure' | 'success';
};

export type GetTransactionsByAccountIdRequestParams = {
  accountId: string;
  limit: number;
  cursor?: {
    timestamp: string; // paginate with timestamp
    indexInChunk: number;
  };
};

type Action =
  | {
      kind: 'createAccount';
      args: {};
    }
  | {
      kind: 'deployContract';
      args: {
        code: string;
      };
    }
  | {
      kind: 'functionCall';
      args: {
        methodName: string;
        args: string;
        gas: number;
        deposit: string;
      };
    }
  | {
      kind: 'transfer';
      args: {
        deposit: string;
      };
    }
  | {
      kind: 'stake';
      args: {
        stake: string;
        publicKey: string;
      };
    }
  | {
      kind: 'addKey';
      args: {
        publicKey: string;
        accessKey: {
          nonce: number;
          permission:
            | {
                type: 'fullAccess';
              }
            | {
                type: 'functionCall';
                contractId: string;
                methodNames: string[];
              };
        };
      };
    }
  | {
      kind: 'deleteKey';
      args: {
        publicKey: string;
      };
    }
  | {
      kind: 'deleteAccount';
      args: {
        beneficiaryId: string;
      };
    };
