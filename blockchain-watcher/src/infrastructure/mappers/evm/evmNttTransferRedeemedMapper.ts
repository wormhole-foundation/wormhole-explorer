import { EvmTransaction, EvmTransactionLog, TransactionFoundEvent } from "../../../domain/entities";
import winston from "winston";
import { ethers } from "ethers";
import { EVMNTTManagerAttributes, NTTTransfer } from "./helpers/ntt";
import { toChainId } from "@wormhole-foundation/sdk-base";
import { LogMapperFn, mapLogDataByTopic, mapTxnStatus } from "./helpers/utils";
import { TRANSFER_REDEEMED_ABI } from "../../../abis/ntt";

let logger: winston.Logger = winston.child({ module: "evmNttTransferRedeemedMapper" });

export const evmNttTransferRedeemedMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> | undefined => {
  const txnStatus = mapTxnStatus(transaction.status);

  const nttTransferInfo = mapLogDataByTopic(NTT_MANAGER_TOPICS, transaction.logs);
  if (!nttTransferInfo) {
    logger.warn(`[${transaction.chain}] Couldn't map ntt transfer: [hash: ${transaction.hash}]`);
    return undefined;
  }

  return {
    name: nttTransferInfo.eventName,
    address: transaction.to,
    chainId: transaction.chainId,
    blockHeight: BigInt(transaction.blockNumber),
    txHash: transaction.hash.substring(2), // Remove 0x
    blockTime: transaction.timestamp,
    attributes: {
      eventName: nttTransferInfo.eventName,
      from: transaction.from,
      to: transaction.to,
      status: txnStatus,
      blockNumber: transaction.blockNumber,
      timestamp: transaction.timestamp,
      txHash: transaction.hash,
      gas: transaction.gas,
      gasPrice: transaction.gasPrice,
      gasUsed: transaction.gasUsed,
      effectiveGasPrice: transaction.effectiveGasPrice,
      nonce: transaction.nonce,
      cost: BigInt(transaction.gasUsed) * BigInt(transaction.effectiveGasPrice),
      digest: nttTransferInfo.digest,
      protocol: "NTT",
    },
    tags: {
      recipientChain: toChainId(transaction.chainId),
    },
  };
};

const mapLogDataFromTransferRedeemed: LogMapperFn<NTTTransfer> = (
  log: EvmTransactionLog
): NTTTransfer => {
  const iface = new ethers.utils.Interface(TRANSFER_REDEEMED_ABI);
  const parsedLog = iface.parseLog(log);

  return {
    eventName: "ntt-transfer-redeemed",
    digest: parsedLog.args.digest,
  };
};

const NTT_MANAGER_TOPICS: Record<string, LogMapperFn<NTTTransfer>> = {
  "0x504e6efe18ab9eed10dc6501a417f5b12a2f7f2b1593aed9b89f9bce3cf29a91":
    mapLogDataFromTransferRedeemed,
};
