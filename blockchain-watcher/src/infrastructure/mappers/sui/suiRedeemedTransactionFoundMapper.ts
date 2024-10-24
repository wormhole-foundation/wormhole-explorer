import { TransactionFoundEvent, TxStatus } from "../../../domain/entities";
import { SuiTransactionBlockReceipt } from "../../../domain/entities/sui";
import { CHAIN_ID_SUI } from "@certusone/wormhole-sdk";
import { findProtocol } from "../contractsMapper";
import { SuiEvent } from "@mysten/sui.js/client";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "suiRedeemedTransactionFoundMapper" });

const REDEEM_EVENT_TAIL = [
  "::complete_transfer::TransferRedeemed",
  "::receive_message::MessageReceived",
];
const SUI_CHAIN = "sui";

export const suiRedeemedTransactionFoundMapper = (
  receipt: SuiTransactionBlockReceipt
): TransactionFoundEvent | undefined => {
  const { events, effects } = receipt;

  const event = events.find((e) => REDEEM_EVENT_TAIL.some((tail) => e.type.endsWith(tail)));
  if (!event) return undefined;

  const protocol = findProtocol(
    SUI_CHAIN,
    event.packageId,
    event.transactionModule,
    receipt.digest
  );

  const vaa = extractVaaInfo(event);
  if (!vaa) return undefined;

  const { type: protocolType, method: protocolMethod } = protocol;
  const { emitterAddress, emitterChain, sequence } = vaa;

  logger.info(
    `[${SUI_CHAIN}] Redeemed transaction info: [digest: ${receipt.digest}][VAA: ${emitterChain}/${emitterAddress}/${sequence}][protocol: ${protocolType}/${protocolMethod}]`
  );

  return {
    name: "transfer-redeemed",
    address: event.packageId,
    blockHeight: BigInt(receipt.checkpoint || 0),
    blockTime: Math.floor(Number(receipt.timestampMs) / 1000), // convert to seconds
    chainId: CHAIN_ID_SUI,
    txHash: receipt.digest,
    attributes: {
      from: event.sender,
      emitterChain,
      emitterAddress,
      sequence,
      status: effects?.status?.status === "failure" ? TxStatus.Failed : TxStatus.Confirmed,
      protocol: protocolMethod,
    },
  };
};

function extractVaaInfo(event: SuiEvent): VaaInformation | undefined {
  const eventTypeTail = event.type.replace(/^0x[a-fA-F0-9]{64}/, "");

  const mapper = REDEEM_EVENTS[eventTypeTail];
  if (!mapper) return undefined;

  const vaaInformation = mapper(event);
  if (!vaaInformation) return undefined;

  return vaaInformation;
}

const mapByParsedJson: EventToVaaMapper = (event: SuiEvent) => {
  const json = event.parsedJson as SuiTransferRedeemedEvent;

  return {
    emitterAddress: Buffer.from(json.emitter_address.value.data).toString("hex"),
    emitterChain: json.emitter_chain,
    sequence: Number(json.sequence),
  };
};

const mapByMessageBody: EventToVaaMapper = (event: SuiEvent) => {
  const json = event.parsedJson as SuiTransferRedeemedEventWithMessageBody;

  return {};
};

export interface SuiRedeemedTransactionFoundMapperConfig {
  redeemEvent: string;
}

interface SuiTransferRedeemedEvent {
  emitter_chain: number;
  sequence: string;
  emitter_address: {
    value: {
      data: number[];
    };
  };
}

interface SuiTransferRedeemedEventWithMessageBody {
  message_body: Array<number>;
}

type VaaInformation = {
  emitterChain?: number;
  emitterAddress?: string;
  sequence?: number;
};

type EventToVaaMapper = (event: SuiEvent) => VaaInformation | undefined;

const REDEEM_EVENTS: Record<string, EventToVaaMapper> = {
  "::complete_transfer::TransferRedeemed": mapByParsedJson, // Token Bridge
  "::receive_message::MessageReceived": mapByMessageBody, // CCTP
};
