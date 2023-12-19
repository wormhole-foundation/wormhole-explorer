export type LogFoundEvent<T> = {
  name: string;
  address: string;
  chainId: number;
  txHash: string;
  blockHeight: bigint;
  blockTime: number;
  attributes: T;
};

export type LogMessagePublished = {
  sequence: number;
  sender: string;
  nonce: number;
  payload: string;
  consistencyLevel: number;
};

export type TransferRedeemed = {
  emitterChainId: number;
  emitterAddress: string;
  sequence: number;
};

export type StandardRelayDelivered = {
  recipientContract: string;
  sourceChain: number;
  sequence: number;
  deliveryVaaHash: string;
  status: number;
  gasUsed: number;
  refundStatus: number;
  additionalStatusInfo: string;
  overridesInfo: string;
};

export type TransactionFoundEvent<T> = {
  name: string;
  address: string;
  txHash: string;
  blockHeight: bigint;
  chainId: number;
  attributes: T;
};

export type FailedRedeemedTransaction = {
  from: string;
  to: string;
  status?: string;
  blockNumber: bigint;
  input: string;
  methodsByAddress?: string;
};

export type TransferRedeemedTransaction = {
  from: string;
  to: string;
  status?: string;
  blockNumber: bigint;
  input: string;
  methodsByAddress?: string;
};

export type StandardRelayDeliveredTransaction = {
  from: string;
  to: string;
  status?: string;
  blockNumber: bigint;
  input: string;
  methodsByAddress?: string;
};
