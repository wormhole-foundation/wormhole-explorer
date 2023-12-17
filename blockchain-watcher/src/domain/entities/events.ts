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
  blockTime: number;
  attributes: T;
};

export type FailedRedeemed = {
  hash: string;
  from: string;
  to: string;
  status?: string;
  blockNumber: bigint;
  blockTimestamp?: number;
  input: string;
  methodsByAddress?: string;
};