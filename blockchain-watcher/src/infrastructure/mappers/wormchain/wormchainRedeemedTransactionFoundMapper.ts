import { TransactionFoundEvent } from "../../../domain/entities";
import { MsgExecuteContract } from "cosmjs-types/cosmwasm/wasm/v1/tx";
import { CosmosRedeem } from "../../../domain/entities/wormchain";
import { decodeTxRaw } from "@cosmjs/proto-signing";
import { mapChain } from "../../../common/wormchain";
import { parseVaa } from "@certusone/wormhole-sdk";
import { base64 } from "ethers/lib/utils";
import winston from "winston";

const MSG_EXECUTE_CONTRACT_TYPE_URL = "/cosmwasm.wasm.v1.MsgExecuteContract";
const PROTOCOL = "Wormhole Gateway";

let logger: winston.Logger = winston.child({ module: "wormchainRedeemedTransactionFoundMapper" });

export const wormchainRedeemedTransactionFoundMapper = (
  cosmosRedeem: CosmosRedeem
): TransactionFoundEvent | undefined => {
  const vaaInformation = mappedVaaInformation(cosmosRedeem.tx);
  const chainId = cosmosRedeem.chainId;
  const chain = mapChain(chainId);
  const hash = cosmosRedeem.hash;

  if (!vaaInformation) {
    logger.warn(`[${chain}] Cannot mapper vaa information: [hash: ${hash}][protocol: ${PROTOCOL}]`);
    return undefined;
  }

  const emitterAddress = vaaInformation.emitterAddress;
  const emitterChain = vaaInformation.emitterChain;
  const sequence = vaaInformation.sequence;
  const sender = senderFromEventAttribute(cosmosRedeem.events);

  logger.info(
    `[${chain}] Redeemed transaction info: [hash: ${hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
  );

  return {
    name: "transfer-redeemed",
    address: sender,
    chainId: chainId,
    txHash: `0x${hash}`,
    blockHeight: BigInt(cosmosRedeem.height),
    blockTime: Math.floor(Number(cosmosRedeem.blockTimestamp) / 1000),
    attributes: {
      emitterAddress: emitterAddress,
      emitterChain: emitterChain,
      sequence: sequence,
      protocol: PROTOCOL,
      status: TxStatus.Completed,
    },
  };
};

function mappedVaaInformation(tx: Buffer): VaaInformation | undefined {
  const decodedTx = decodeTxRaw(tx);
  const message = decodedTx.body.messages.find(
    (tx) => tx.typeUrl === MSG_EXECUTE_CONTRACT_TYPE_URL
  );

  if (message) {
    const parsedMessage = MsgExecuteContract.decode(message.value);

    const instruction = JSON.parse(Buffer.from(parsedMessage.msg).toString());
    const base64Vaa = instruction?.complete_transfer_and_convert?.vaa;

    if (base64Vaa) {
      const vaa = parseVaa(base64.decode(base64Vaa));

      return {
        emitterAddress: vaa.emitterAddress.toString("hex"),
        emitterChain: vaa.emitterChain,
        sequence: Number(vaa.sequence),
      };
    }
  }
}

function senderFromEventAttribute(events: EventsType[]): string {
  const sender = events
    .find((event) => event.type === "message")
    ?.attributes.find((attr) => attr.key === "sender")?.value;

  return sender || "";
}

type VaaInformation = {
  emitterChain?: number;
  emitterAddress?: string;
  sequence?: number;
};

enum TxStatus {
  Completed = "completed",
  Failed = "failed",
}

type EventsType = {
  type: string;
  attributes: {
    key: string;
    value: string;
    index: boolean;
  }[];
};
