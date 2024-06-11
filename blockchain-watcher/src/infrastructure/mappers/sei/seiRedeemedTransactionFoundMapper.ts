import { TransactionFoundEvent } from "../../../domain/entities";
import { MsgExecuteContract } from "cosmjs-types/cosmwasm/wasm/v1/tx";
import { decodeTxRaw } from "@cosmjs/proto-signing";
import { SeiRedeem } from "../../../domain/entities/sei";
import { parseVaa } from "@certusone/wormhole-sdk";
import { base64 } from "ethers/lib/utils";
import winston from "winston";

const MSG_EXECUTE_CONTRACT_TYPE_URL = "/cosmwasm.wasm.v1.MsgExecuteContract";
const SEI_CHAIN_ID = 32;
const PROTOCOL = "Token Bridge";

let logger: winston.Logger = winston.child({ module: "seiRedeemedTransactionFoundMapper" });

export const seiRedeemedTransactionFoundMapper = (
  addresses: string[],
  transaction: SeiRedeem
): TransactionFoundEvent | undefined => {
  const vaaInformation = mappedVaaInformation(transaction.tx);
  if (!vaaInformation) {
    return undefined;
  }
  const txAttributes = transactionAttributes(addresses, transaction);
  if (!txAttributes) {
    return undefined;
  }
  const hash = transaction.hash;

  const emitterAddress = vaaInformation.emitterAddress;
  const emitterChain = vaaInformation.emitterChain;
  const sequence = vaaInformation.sequence;

  logger.info(
    `[sei] Redeemed transaction info: [hash: ${hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
  );

  return {
    name: "transfer-redeemed",
    address: txAttributes.receiver,
    chainId: SEI_CHAIN_ID,
    txHash: `0x${hash}`,
    blockHeight: BigInt(transaction.height),
    blockTime: Math.floor(transaction.timestamp! / 1000),
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

function transactionAttributes(
  addresses: string[],
  tx: SeiRedeem
): TransactionAttributes | undefined {
  let receiver: string | undefined;

  for (const event of tx.events) {
    for (const attr of event.attributes) {
      const key = Buffer.from(attr.key, "base64").toString().toLowerCase();
      const value = Buffer.from(attr.value, "base64").toString().toLowerCase();

      switch (key) {
        case "_contract_address":
        case "contract_address":
          if (addresses.includes(value.toLowerCase())) {
            receiver = value.toLowerCase();
          }
          break;
      }
    }
  }
  if (receiver) {
    return { receiver };
  }
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

type TransactionAttributes = {
  receiver: string;
};
