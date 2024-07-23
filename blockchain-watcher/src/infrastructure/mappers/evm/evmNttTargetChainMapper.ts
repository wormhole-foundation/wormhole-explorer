import { EvmTransaction, EvmTransactionLog, TransactionFoundEvent } from "../../../domain/entities";
import winston from "winston";
import { ethers } from "ethers";
import { EVMNTTManagerAttributes, NTTTransfer } from "./helpers/ntt";
import { toChainId, chainIdToChain } from "@wormhole-foundation/sdk-base";
import { LogToNTTTransfer, mapLogDataByTopic, mappedTxnStatus } from "./helpers/utils";
import { UniversalAddress } from "@wormhole-foundation/sdk-definitions";

let logger: winston.Logger = winston.child({ module: "evmTargetChainNttMapper" });

export const evmTargetChainNttMapper = (
  transaction: EvmTransaction
): TransactionFoundEvent<EVMNTTManagerAttributes> | undefined => {
  const vaaInformation = mapLogDataByTopic(RECEIVED_MESSAGE_TOPIC, transaction.logs);
  const txnStatus = mappedTxnStatus(transaction.status);

  const emitterChainId = vaaInformation?.emitterChainId;
  const emitterAddress = vaaInformation?.emitterAddress;
  const sequence = vaaInformation?.sequence;

  logger.info(
    `[${transaction.chain}] Redeemed transaction info: [hash: ${transaction.hash}][VAA: ${
      emitterChainId ?? ""
    }/${emitterAddress ?? ""}/${sequence ?? ""}]`
  );

  const nttTransferInfo = mapLogDataByTopic(MAIN_TOPICS, transaction.logs);
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
      protocol: "ntt",
    },
    tags: {
      recipientChain: toChainId(transaction.chainId),
      ...(emitterChainId && {
        emitterChain: toChainId(emitterChainId),
      }),
    },
  };
};

const mapLogDataFromTransferRedeemed: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog
): NTTTransfer => {
  const abi = "event TransferRedeemed(bytes32 indexed digest);";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);

  return {
    eventName: "transfer-redeemed",
    digest: parsedLog.args.digest,
  };
};

const mapLogDataFromReceivedRelayedMessage: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog
): NTTTransfer => {
  const abi =
    "event ReceivedRelayedMessage(bytes32 digest, uint16 emitterChainId, bytes32 emitterAddress)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);

  return {
    eventName: "received-relayed-message",
    digest: parsedLog.args.digest,
  };
};

const mapLogDataFromMessageAttestedTo: LogToNTTTransfer<NTTTransfer> = (
  log: EvmTransactionLog
): NTTTransfer => {
  const abi = "event MessageAttestedTo (bytes32 digest, address transceiver, uint8 index)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);

  return {
    eventName: "message-attested-to",
    digest: parsedLog.args.digest,
  };
};

const mapLogDataToVaaInfo = (log: EvmTransactionLog) => {
  const abi =
    "event ReceivedMessage(bytes32 digest, uint16 emitterChainId, bytes32 emitterAddress, uint64 sequence)";
  const iface = new ethers.utils.Interface([abi]);
  const parsedLog = iface.parseLog(log);
  const emitterChainId = toChainId(parsedLog.args.emitterChainId.toString());

  return {
    emitterChainId,
    emitterAddress: new UniversalAddress(parsedLog.args.emitterAddress).toNative(
      chainIdToChain(emitterChainId)
    ),
    sequence: Number(parsedLog.args.sequence),
  };
};

const RECEIVED_MESSAGE_TOPIC = {
  "0xaa8267908e8d2BEfeB601f88A7Cf3ec148039423": mapLogDataToVaaInfo,
};

const MAIN_TOPICS: Record<string, LogToNTTTransfer<NTTTransfer>> = {
  "0x504e6efe18ab9eed10dc6501a417f5b12a2f7f2b1593aed9b89f9bce3cf29a91":
    mapLogDataFromTransferRedeemed,
  "0xf557dbbb087662f52c815f6c7ee350628a37a51eae9608ff840d996b65f87475":
    mapLogDataFromReceivedRelayedMessage,
  "0x35a2101eaac94b493e0dfca061f9a7f087913fde8678e7cde0aca9897edba0e5":
    mapLogDataFromMessageAttestedTo,
};
