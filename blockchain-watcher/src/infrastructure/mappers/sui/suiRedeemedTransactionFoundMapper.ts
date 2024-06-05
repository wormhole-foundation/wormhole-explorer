import { TransactionFoundEvent, TransferRedeemed, TxStatus } from "../../../domain/entities";
import { SuiTransactionBlockReceipt } from "../../../domain/entities/sui";
import { CHAIN_ID_SUI } from "@certusone/wormhole-sdk";
import { findProtocol } from "../contractsMapper";
import { SuiEvent } from "@mysten/sui.js/client";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "suiRedeemedTransactionFoundMapper" });

const REDEEM_EVENT_TAIL = "::complete_transfer::TransferRedeemed";
const SUI_CHAIN = "sui";

export const suiRedeemedTransactionFoundMapper = (
  receipt: SuiTransactionBlockReceipt
): TransactionFoundEvent | undefined => {
  const { events, effects } = receipt;

  const event = events.find((e) => e.type.endsWith(REDEEM_EVENT_TAIL));
  if (!event) return undefined;

  const protocol = findProtocol(
    SUI_CHAIN,
    event.packageId,
    event.transactionModule,
    receipt.digest
  );

  const vaaInformation = extractRedeemInfo(event);
  if (!vaaInformation) return undefined;

  const { emitterAddress, emitterChainId: emitterChain, sequence } = vaaInformation;

  if (protocol && protocol.type && emitterAddress && emitterChain && sequence) {
    const { type: protocolType, method: protocolMethod } = protocol;

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
  }
};

function extractRedeemInfo(event: SuiEvent): TransferRedeemed | undefined {
  const json = event.parsedJson as SuiTransferRedeemedEvent;

  return {
    emitterAddress: Buffer.from(json.emitter_address.value.data).toString("hex"),
    emitterChainId: json.emitter_chain,
    sequence: Number(json.sequence),
  };
}

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
