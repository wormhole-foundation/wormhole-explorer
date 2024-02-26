import {
  SuiTransactionFoundAttributes,
  TransactionFoundEvent,
  TransferRedeemed,
  TxStatus,
} from "../../../domain/entities";
import { Protocol, contractsMapperConfig } from "../contractsMapper";
import { SuiTransactionBlockReceipt } from "../../../domain/entities/sui";
import { CHAIN_ID_SUI } from "@certusone/wormhole-sdk";
import { SuiEvent } from "@mysten/sui.js/client";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "suiRedeemedTransactionFoundMapper" });

const REDEEM_EVENT_TAIL = "::complete_transfer::TransferRedeemed";

export const suiRedeemedTransactionFoundMapper = (
  receipt: SuiTransactionBlockReceipt
): TransactionFoundEvent<SuiTransactionFoundAttributes> | undefined => {
  const { events, effects } = receipt;

  const event = events.find((e) => e.type.endsWith(REDEEM_EVENT_TAIL));
  if (!event) return undefined;

  const protocol = findProtocol(event.packageId, event.transactionModule, receipt.digest);

  const vaa = extractRedeemInfo(event);
  if (!vaa) return undefined;
  const { emitterAddress, emitterChainId: emitterChain, sequence } = vaa;

  if (protocol && protocol.type && protocol.method) {
    logger.info(
      `[sui] Redeemed transaction info: [digest: ${receipt.digest}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
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
        protocol: protocol.type,
      },
    };
  }
};

const findProtocol = (address: string, method: string, hash: string): Protocol | undefined => {
  for (const contract of contractsMapperConfig.contracts) {
    if (contract.chain === "sui") {
      const foundProtocol = contract.protocols.find((protocol) =>
        protocol.addresses.some((addr) => addr.toLowerCase() === address.toLowerCase())
      );
      const foundMethod = foundProtocol?.methods.find((m) => m.method === method);

      if (foundMethod && foundProtocol) {
        return {
          method: foundMethod.method,
          type: foundMethod.methodId,
        };
      }
    }
  }
  logger.warn(
    `[sui] Protocol not found, [tx hash: ${hash}][address: ${address}][method: ${method}}]`
  );
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
