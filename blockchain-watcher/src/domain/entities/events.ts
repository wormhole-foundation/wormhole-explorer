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
  blockTime: number;
  attributes: T;
};

export type TransactionFound = {
  from: string;
  to: string;
  status?: string;
  blockNumber: bigint;
  input: string;
  methodsByAddress: string;
  timestamp: number;
  blockHash: string;
  gas: string;
  gasPrice: string;
  maxFeePerGas: string;
  maxPriorityFeePerGas: string;
  nonce: string;
  r: string;
  s: string;
  transactionIndex: string;
  type: string;
  v: string;
  value: string;
  sequence?: number;
  emitterChain?: number;
  emitterAddress?: string;
  protocol: string;
};

export type InstructionFound = {
  methodsByAddress: string;
  status: string;
  emitterChain: number;
  emitterAddress: string;
  sequence: number;
  protocol: string;
};
