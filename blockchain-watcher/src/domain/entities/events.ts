export type LogFoundEvent<T> = {
  name: string;
  address: string;
  chainId: number;
  txHash: string;
  blockHeight: bigint;
  /* value is in seconds */
  blockTime: number;
  attributes: T;
  tags?: Record<string, string>;
};

export type LogMessagePublished = {
  sequence: number;
  sender: string;
  nonce: number;
  payload: string;
  consistencyLevel: number;
  chain?: string;
};

export type MessageSent = {
  destinationCaller: string;
  destinationDomain: string;
  messageSender: string;
  mintRecipient: string;
  sourceDomain: string;
  burnToken: string;
  recipient: string;
  protocol: string;
  sender: string;
  amount: bigint;
  nonce: bigint;
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

export type TransactionFoundEvent<
  T extends TransactionFoundAttributes = TransactionFoundAttributes
> = {
  name: string;
  chainId: number;
  address: string;
  txHash: string;
  blockHeight: bigint;
  blockTime: number;
  attributes: T;
  tags?: Record<string, unknown>;
};

export type TransactionFound = {
  from: string;
  to: string;
  status?: string;
};

// TODO: some of these attributes might not make sense for all chains so no point
// on keeping them on this base type
export type TransactionFoundAttributes = {
  name?: string;
  emitterChain?: number;
  emitterAddress?: string;
  sequence?: number;
  methodsByAddress?: string;
  from?: string;
  to?: string;
  status?: string;
  protocol: string;
  chain?: string;
};

export type EvmTransactionFoundAttributes = TransactionFoundAttributes & {
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
  protocol: string;
  gasUsed: string;
  effectiveGasPrice: string;
};

export type InstructionFound = {
  methodsByAddress: string;
  status: string;
  emitterChain: number;
  emitterAddress: string;
  sequence: number;
  protocol: string;
  fee: number | undefined;
  from: string;
  to: string;
};

export type NearTransactionFoundAttributes = TransactionFoundAttributes & {
  consistencyLevel?: number;
  nonce?: number;
};

export enum TxStatus {
  Confirmed = "completed",
  Unkonwn = "unknown",
  Failed = "failed",
}
