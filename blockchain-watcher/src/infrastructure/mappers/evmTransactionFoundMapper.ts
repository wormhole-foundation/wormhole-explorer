import { methodNameByAddressMapper } from "../../domain/actions/evm/mappers/methodNameByAddressMapper";
import { EvmTransaction, TransactionFound, TransactionFoundEvent } from "../../domain/entities";
import { parseVaa } from "@certusone/wormhole-sdk";

const TX_STATUS_SUCCESS = "0x1";
const TX_STATUS_FAIL_REVERTED = "0x0";

export const evmTransactionFoundMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<TransactionFound> => {
  const protocol = methodNameByAddressMapper(
    transaction.chain,
    transaction.environment,
    transaction
  );

  const vaaInformation = mappedVAA(transaction);
  const status = mappedStatus(transaction);

  return {
    name: "evm-transaction-found",
    address: transaction.to,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    blockTime: Date.now(), // TODO: See this value
    attributes: {
      name: protocol?.name,
      from: transaction.from,
      to: transaction.to,
      status: status,
      blockNumber: transaction.blockNumber,
      input: transaction.input,
      methodsByAddress: protocol?.method,
      timestamp: transaction.timestamp,
      blockHash: transaction.blockHash,
      gas: transaction.gas,
      gasPrice: transaction.gasPrice,
      maxFeePerGas: transaction.maxFeePerGas,
      maxPriorityFeePerGas: transaction.maxPriorityFeePerGas,
      nonce: transaction.nonce,
      r: transaction.r,
      s: transaction.s,
      transactionIndex: transaction.transactionIndex,
      type: transaction.type,
      v: transaction.v,
      value: transaction.value,
      sequence: vaaInformation.sequence,
      emitterAddress: vaaInformation.emitterAddress,
      emitterChain: vaaInformation.emitterChain,
    },
  };
};

const mappedVAA = (transaction: EvmTransaction) => {
  const input = transaction.input;
  const truncatedString: string = input.substring(137);
  const vaaBytes = Buffer.from(truncatedString, "hex");

  const VAA = parseVaa(vaaBytes);

  return {
    sequence: Number(VAA.sequence),
    emitterChain: VAA.emitterChain,
    emitterAddress: VAA.emitterAddress.toString("hex"),
  };
};

const mappedStatus = (transaction: EvmTransaction): string => {
  switch (transaction.status) {
    case TX_STATUS_SUCCESS:
      return status.TxStatusConfirmed;
      break;
    case TX_STATUS_FAIL_REVERTED:
      return status.TxStatusFailedToProcess;
      break;
    default:
      return status.TxStatusUnkonwn;
  }
};

export enum status {
  TxStatusFailedToProcess = "failed",
  TxStatusConfirmed = "completed",
  TxStatusUnkonwn = "unknown",
}
