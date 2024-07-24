import { EvmTransaction, EvmTransactionLog, TransactionFoundEvent } from "../../../domain/entities";
import winston from "winston";
import { ethers } from "ethers";
import { EVMNTTManagerAttributes, NTTTransfer } from "./helpers/ntt";
import { toChainId } from "@wormhole-foundation/sdk-base";
import { LogMapperFn, mapLogDataByTopic, mapTxnStatus } from "./helpers/utils";
import { MESSAGE_ATTESTED_TO_ABI } from "../../../abis/ntt";

let logger: winston.Logger = winston.child({ module: "evmNttMessageAttestedToMapper" });

export const evmNttMessageAttestedToMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> | undefined => {
  const txnStatus = mapTxnStatus(transaction.status);

  const nttTransferInfo = mapLogDataByTopic(NTT_TRANSCEIVER_TOPICS, transaction.logs);
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
      trasceiverType: nttTransferInfo.transceiverType,
    },
  };
};

// event MessageAttestedTo (bytes32 digest, address transceiver, uint8 index)
const mapLogDataFromMessageAttestedTo: LogMapperFn<NTTTransfer> = (
  log: EvmTransactionLog
): NTTTransfer => {
  const iface = new ethers.utils.Interface(MESSAGE_ATTESTED_TO_ABI);
  const parsedLog = iface.parseLog(log);

  return {
    eventName: "ntt-message-attested-to",
    digest: parsedLog.args.digest,
    transceiverType: mapTransceiverIndex(Number(parsedLog.args.index)),
  };
};

const NTT_TRANSCEIVER_TOPICS: Record<string, LogMapperFn<NTTTransfer>> = {
  "0x35a2101eaac94b493e0dfca061f9a7f087913fde8678e7cde0aca9897edba0e5":
    mapLogDataFromMessageAttestedTo,
};

const mapTransceiverIndex = (index: number): "axelar" | "wormhole" => {
  return index === 0 ? "wormhole" : "axelar";
};
