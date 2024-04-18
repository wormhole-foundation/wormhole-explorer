import { TransactionFoundEvent } from "../../../domain/entities";
import { AptosTransaction } from "../../../domain/entities/aptos";
import { CHAIN_ID_APTOS } from "@certusone/wormhole-sdk";
import { findProtocol } from "../contractsMapper";
import { hexZeroPad } from "ethers/lib/utils";
import { parseVaa } from "@certusone/wormhole-sdk";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "aptosRedeemedTransactionFoundMapper" });

const WORMHOLE_EVENT = "::state::WormholeMessage";
const APTOS_CHAIN = "aptos";

export const aptosRedeemedTransactionFoundMapper = (
  transaction: AptosTransaction
): TransactionFoundEvent | undefined => {
  const address = transaction.payload.function.split("::")[0];
  const type = transaction.payload.function;

  const protocol = findProtocol(APTOS_CHAIN, address, type, transaction.hash);

  if (!protocol) {
    return undefined;
  }

  const { type: protocolType, method: protocolMethod } = protocol;
  const vaaInformation = mappedVaaInformation(transaction);

  if (!vaaInformation) {
    logger.warn(
      `[${CHAIN_ID_APTOS}] Cannot mapper vaa information: [tx hash: ${transaction.hash}][protocol: ${protocolType}/${protocolMethod}]`
    );
    return undefined;
  }

  const emitterAddress = vaaInformation.emitterAddress;
  const emitterChain = vaaInformation.emitterChain;
  const timestamp = vaaInformation.timestamp;
  const sequence = vaaInformation.sequence;

  logger.info(
    `[${APTOS_CHAIN}] Redeemed transaction info: [hash: ${transaction.hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
  );

  return {
    name: "transfer-redeemed",
    address: address,
    blockHeight: transaction.blockHeight,
    blockTime: timestamp!,
    chainId: CHAIN_ID_APTOS,
    txHash: transaction.hash,
    attributes: {
      from: address,
      emitterChain: emitterChain,
      emitterAddress: emitterAddress,
      sequence: Number(sequence),
      status: transaction.status === true ? TxStatus.Completed : TxStatus.Failed,
      protocol: protocolType,
    },
  };
};

enum TxStatus {
  Completed = "completed",
  Failed = "failed",
}

/**
 * Mapped vaa information from transaction using the function value to map the correct mapper
 */
const mappedVaaInformation = (transaction: AptosTransaction): VaaInformation | undefined => {
  const mapper = REDEEM_FUNCTIONS[transaction.payload.function];
  const vaaInformation = mapper(transaction);
  return vaaInformation;
};

const mapVaaFromArguments: LogToVaaMapper = (transaction: AptosTransaction) => {
  const vaaBuffer = Buffer.from(transaction.payload?.arguments[0]?.substring(2), "hex");
  const vaa = parseVaa(vaaBuffer);

  return {
    emitterAddress: vaa.emitterAddress.toString("hex"),
    emitterChain: vaa.emitterChain,
    timestamp: Number(vaa.timestamp),
    sequence: Number(vaa.sequence),
  };
};

const mapVaaFromEvents: LogToVaaMapper = (transaction: AptosTransaction) => {
  const data = transaction.events.find(
    (e: { type: string | string[] }) => e.type.includes(WORMHOLE_EVENT) // Try to find wormhole event
  )?.data;

  if (!data) {
    return undefined;
  }

  const emitterAddress = hexZeroPad(`0x${parseInt(data?.sender).toString(16)}`, 32).substring(2);

  return {
    emitterAddress: emitterAddress,
    emitterChain: CHAIN_ID_APTOS,
    timestamp: Number(data.timestamp),
    sequence: Number(data.sequence),
  };
};

type VaaInformation = {
  emitterAddress?: string;
  emitterChain?: number;
  timestamp?: number;
  sequence?: number;
};

type LogToVaaMapper = (log: AptosTransaction) => VaaInformation | undefined;

const REDEEM_FUNCTIONS: Record<string, LogToVaaMapper> = {
  "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f::complete_transfer::submit_vaa_and_register_entry":
    mapVaaFromArguments, // Token Bridge
  "0x1bdffae984043833ed7fe223f7af7a3f8902d04129b14f801823e64827da7130::transfer_nft::transfer_nft_entry":
    mapVaaFromEvents, // NFT
};
