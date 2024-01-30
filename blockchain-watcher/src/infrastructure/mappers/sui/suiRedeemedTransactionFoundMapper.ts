import { CHAIN_ID_SUI } from "@certusone/wormhole-sdk";
import { MoveCallSuiTransaction, SuiEvent, SuiTransaction } from "@mysten/sui.js/client";
import winston from "winston";
import { TransactionFoundEvent, TransferRedeemed } from "../../../domain/entities";
import { SuiTransactionBlockReceipt } from "../../../domain/entities/sui";

let logger: winston.Logger = winston.child({ module: "suiRedeemedTransactionFoundMapper" });

export const suiRedeemedTransactionFoundMapper = (
  receipt: SuiTransactionBlockReceipt
): TransactionFoundEvent | undefined => {
  const { events } = receipt;
  const event = events.find((e) => e.type.endsWith("::complete_transfer::TransferRedeemed"));
  if (!event) return undefined;

  const vaa = extractRedeemInfo(event);
  if (!vaa) return undefined;
  const { emitterAddress, emitterChainId: emitterChain, sequence } = vaa;

  logger.info(
    `[sui][suiRedeemedTransactionFoundMapper] Redeemed Transfer info: [digest: ${receipt.digest}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
  );

  return {
    name: "transfer-redeemed",
    address: event.packageId,
    blockHeight: BigInt(receipt.checkpoint || 0),
    blockTime: Number(receipt.timestampMs),
    chainId: CHAIN_ID_SUI,
    txHash: receipt.digest,
    attributes: {
      from: event.sender,
      emitterChain,
      emitterAddress,
      sequence,
      status: "completed",
    },
  };
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
