import { EvmTransaction, LogFoundEvent, ProposalCreated } from "../../../domain/entities";
import { HandleEvmConfig } from "../../../domain/actions";
import { findProtocol } from "../contractsMapper";
import { ethers } from "ethers";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "evmProposalCreatedMessageMapper" });

export const evmProposalCreatedMessageMapper = (
  transaction: EvmTransaction,
  cfg?: HandleEvmConfig
): LogFoundEvent<ProposalCreated> | undefined => {
  const proposalCreatedMessage = mappedProposalCreatedMessage(transaction, cfg!);

  if (!proposalCreatedMessage) {
    logger.warn(
      `[${transaction.chain}] Failed to parse proposal created message for [tx: ${transaction.hash}]`
    );
    return undefined;
  }

  const first10Characters = transaction.input.slice(0, 10);
  const protocol = findProtocol(
    transaction.chain,
    transaction.to,
    first10Characters,
    transaction.hash
  );

  logger.info(
    `[${transaction.chain}] Proposal created message info: [tx: ${transaction.hash}] [protocol: ${protocol.type}/${protocol.method}]`
  );

  return {
    name: "proposal-created",
    address: transaction.to,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    blockHeight: BigInt(transaction.blockNumber),
    blockTime: transaction.timestamp,
    attributes: {
      ...proposalCreatedMessage,
    },
  };
};

const mappedProposalCreatedMessage = (
  transaction: EvmTransaction,
  cfg: HandleEvmConfig
): ProposalCreated | undefined => {
  const filterLogs = transaction.logs.filter((log) => {
    return EVENT_TOPICS[log.topics[0]];
  });

  if (!filterLogs) return undefined;

  for (const log of filterLogs) {
    const mapper = EVENT_TOPICS[log.topics[0]];
    const bodyMessage = mapper(log, cfg, transaction.input, transaction.hash);

    if (bodyMessage) {
      return bodyMessage;
    }
  }
};

const mapLogFromTopics: LogToVaaMapper = (
  log: EvmTransactionLog,
  cfg: HandleEvmConfig,
  input: string,
  hash: string
) => {
  try {
    if (!log.topics[0]) {
      return undefined;
    }

    const abi = cfg.abis?.find((abi) => abi.topic === log.topics[0]);
    if (!abi) return undefined;

    const iface = new ethers.utils.Interface([`function ${abi.abi}`]);
    const decodedFulfillOrderFunction = iface.decodeFunctionData(abi.abi, input);

    return {
      description: decodedFulfillOrderFunction.description,
      callDatas: decodedFulfillOrderFunction.calldatas,
      targets: decodedFulfillOrderFunction.targets,
    };
  } catch (e) {
    logger.error(`[${cfg.chain}] Failed to parse proposal created message for [tx: ${hash}]`, e);
    return undefined;
  }
};

const EVENT_TOPICS: Record<string, LogToVaaMapper> = {
  "0x7d84a6263ae0d98d3329bd7b46bb4e8d6f98cd35a7adb45c274c8b7fd5ebd5e0": mapLogFromTopics, // ProposalCreated topic
};

type LogToVaaMapper = (
  log: EvmTransactionLog,
  cfg: HandleEvmConfig,
  input: string,
  hash: string
) => any | undefined;

type EvmTransactionLog = { address: string; topics: string[]; data: string };
