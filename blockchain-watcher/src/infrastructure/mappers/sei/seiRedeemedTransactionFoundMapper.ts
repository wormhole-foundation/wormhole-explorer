import { CosmosTransaction, WormchainBlockLogs } from "../../../domain/entities/wormchain";
import { TransactionFoundEvent } from "../../../domain/entities";
import { MsgExecuteContract } from "cosmjs-types/cosmwasm/wasm/v1/tx";
import { decodeTxRaw } from "@cosmjs/proto-signing";
import { mapChain } from "../../../common/wormchain";
import { parseVaa } from "@certusone/wormhole-sdk";
import { base64 } from "ethers/lib/utils";
import winston from "winston";

const MSG_EXECUTE_CONTRACT_TYPE_URL = "/cosmwasm.wasm.v1.MsgExecuteContract";
const SEI_CHAIN_ID = 32;
const PROTOCOL = "Token Bridge";

let logger: winston.Logger = winston.child({ module: "wormchainRedeemedTransactionFoundMapper" });

export const seiRedeemedTransactionFoundMapper = (
  _: string[],
  log: WormchainBlockLogs
): TransactionFoundEvent[] | undefined => {
  const result: TransactionFoundEvent[] = [];

  log.transactions?.forEach((tx) => {
    const vaaInformation = mappedVaaInformation(tx.tx);
    const txAttributes = transactionAttributes(tx);
    const chain = mapChain(SEI_CHAIN_ID);
    const hash = tx.hash;

    if (!vaaInformation || !txAttributes) {
      logger.warn(
        `[${chain}] Cannot mapper vaa information or sender: [hash: ${hash}][VAA: ${JSON.stringify(
          vaaInformation
        )}][txAttributes: ${JSON.stringify(txAttributes)}`
      );
      return undefined;
    }

    const emitterAddress = vaaInformation.emitterAddress;
    const emitterChain = vaaInformation.emitterChain;
    const sequence = vaaInformation.sequence;

    logger.info(
      `[${chain}] Redeemed transaction info: [hash: ${hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
    );

    result.push({
      name: "transfer-redeemed",
      address: txAttributes.to,
      chainId: SEI_CHAIN_ID,
      txHash: `0x${hash}`,
      blockHeight: BigInt(tx.height),
      blockTime: Math.floor(Number(log.timestamp) / 1000),
      attributes: {
        emitterAddress: emitterAddress,
        emitterChain: emitterChain,
        sequence: sequence,
        protocol: PROTOCOL,
        status: TxStatus.Completed,
      },
    });
  });

  return result;
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

function transactionAttributes(tx: CosmosTransaction): TransactionAttributes | undefined {
  let to: string | undefined;

  for (const attr of tx.attributes) {
    const key = Buffer.from(attr.key, "base64").toString().toLowerCase();
    const value = Buffer.from(attr.value, "base64").toString().toLowerCase();

    if (key !== "to") continue;

    to = value;
  }

  return to ? { to } : undefined;
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
  to: string;
};
