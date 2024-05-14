import { TransactionFoundEvent } from "../../../domain/entities";
import { CosmosRedeem } from "../../../domain/entities/wormchain";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "wormchainRedeemedTransactionFoundMapper" });

export const wormchainRedeemedTransactionFoundMapper = (
  cosmosRedeem: CosmosRedeem
): TransactionFoundEvent | undefined => {
  const emitterAddress = cosmosRedeem.vaaEmitterAddress;
  const emitterChain = cosmosRedeem.vaaEmitterChain;
  const sequence = Number(cosmosRedeem.vaaSequence);
  const chainId = cosmosRedeem.chainId;
  const sender = senderFromEventAttribute(cosmosRedeem.events);
  const hash = cosmosRedeem.hash;

  logger.info(
    `[${chainId}] Redeemed transaction info: [hash: ${hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
  );

  return {
    name: "transfer-redeemed",
    address: sender,
    chainId: chainId,
    txHash: hash,
    blockHeight: BigInt(cosmosRedeem.height),
    blockTime: Math.floor(Number(cosmosRedeem.blockTimestamp) / 1000),
    attributes: {
      emitterAddress: cosmosRedeem.vaaEmitterAddress,
      emitterChain: emitterChain,
      sequence: sequence,
      protocol: "Wormhole Gateway",
      status: TxStatus.Completed,
    },
  };
};

function senderFromEventAttribute(events: EventsType[]): string {
  const sender = events
    .find((event) => event.type === "message")
    ?.attributes.find((attr) => attr.key === "sender")?.value;

  return sender || "";
}

enum TxStatus {
  Completed = "completed",
  Failed = "failed",
}

type EventsType = {
  type: string;
  attributes: [
    {
      key: string;
      value: string;
      index: boolean;
    }
  ];
};
