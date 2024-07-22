import { TransactionFoundEvent } from "../../../domain/entities";
import { MsgExecuteContract } from "cosmjs-types/cosmwasm/wasm/v1/tx";
import { CosmosTransaction } from "../../../domain/entities/cosmos";
import { decodeTxRaw } from "@cosmjs/proto-signing";
import { parseVaa } from "@certusone/wormhole-sdk";
import { base64 } from "ethers/lib/utils";
import winston from "winston";

const MSG_EXECUTE_CONTRACT_TYPE_URL = "/cosmwasm.wasm.v1.MsgExecuteContract";
const PROTOCOL = "Token Bridge";

let logger: winston.Logger = winston.child({ module: "cosmosRedeemedTransactionFoundMapper" });

export const cosmosRedeemedTransactionFoundMapper = (
  addresses: string[],
  transaction: CosmosTransaction
): TransactionFoundEvent | undefined => {
  const vaaInformation = mappedVaaInformation(transaction.tx);
  if (!vaaInformation) {
    return undefined;
  }
  const txAttributes = transactionAttributes(addresses, transaction);
  if (!txAttributes || !transaction.timestamp) {
    return undefined;
  }
  const emitterAddress = vaaInformation.emitterAddress;
  const emitterChain = vaaInformation.emitterChain;
  const sequence = vaaInformation.sequence;
  const hash = transaction.hash;

  logger.info(
    `[${transaction.chain}] Redeemed transaction info: [hash: ${hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
  );

  return {
    name: "transfer-redeemed",
    address: txAttributes.receiver,
    chainId: transaction.chainId,
    txHash: hash,
    blockHeight: BigInt(transaction.height),
    blockTime: transaction.timestamp,
    attributes: {
      emitterAddress: emitterAddress,
      emitterChain: emitterChain,
      sequence: sequence,
      protocol: PROTOCOL,
      status: TxStatus.Completed,
      chain: transaction.chain,
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
    const base64Vaa =
      instruction?.complete_transfer_and_convert?.vaa || instruction?.submit_vaa?.data;

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
  transaction: CosmosTransaction
): TransactionAttributes | undefined {
  let receiver: string | undefined;

  for (const event of transaction.events) {
    for (const attr of event.attributes) {
      const { key, value } = decodeAttributes(transaction.chain!, attr);

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

function decodeAttributes(
  chain: string,
  attr: {
    index: boolean;
    value: string;
    key: string;
  }
): { key: string; value: string } {
  // Dependes the chain, we need to decode the key and value from base64, Terra and Terra2 are already decoded and other chains are not
  if (["terra", "terra2"].includes(chain)) {
    return {
      key: attr.key,
      value: attr.value,
    };
  } else {
    return {
      key: Buffer.from(attr.key, "base64").toString().toLowerCase(),
      value: Buffer.from(attr.value, "base64").toString().toLowerCase(),
    };
  }
}

type TransactionAttributes = {
  receiver: string;
};

type VaaInformation = {
  emitterAddress?: string;
  emitterChain?: number;
  sequence?: number;
};

enum TxStatus {
  Completed = "completed",
  Failed = "failed",
}
