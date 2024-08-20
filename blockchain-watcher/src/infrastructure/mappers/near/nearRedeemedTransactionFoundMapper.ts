import { TransactionFoundEvent, TxStatus } from "../../../domain/entities";
import { NearTransaction } from "../../../domain/entities/near";
import { parseVaa } from "@certusone/wormhole-sdk";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "nearRedeemedTransactionFoundMapper" });

const PROTOCOL = "Token Bridge";

export const nearRedeemedTransactionFoundMapper = (
  transaction: NearTransaction
): TransactionFoundEvent | undefined => {
  const vaaInformation = mappedVaaInformation(transaction.actions[0].functionCall.args);

  if (!vaaInformation) {
    logger.warn(`[near] Cannot mapper vaa information: [hash: ${transaction.hash}]`);
    return undefined;
  }
  const emitterAddress = vaaInformation.emitterAddress;
  const emitterChain = vaaInformation.emitterChain;
  const sequence = vaaInformation.sequence;

  logger.info(
    `[near] Redeemed transaction info: [hash: ${transaction.hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}][protocol: ${PROTOCOL}]`
  );

  return {
    name: "transfer-redeemed",
    address: transaction.receiverId,
    blockHeight: transaction.blockHeight,
    blockTime: transaction.timestamp,
    chainId: transaction.chainId,
    txHash: transaction.hash,
    attributes: {
      from: transaction.signerId,
      emitterChain: emitterChain,
      emitterAddress: emitterAddress,
      sequence: sequence,
      status: TxStatus.Confirmed,
      protocol: PROTOCOL,
    },
  };
};

function mappedVaaInformation(args: string): VaaInformation | undefined {
  const argsToString = Buffer.from(args, "base64").toString("utf-8");
  const data = JSON.parse(argsToString) as VAA;

  if (data && data.vaa) {
    const byteArray = Buffer.from(data.vaa, "hex");
    const vaaParsed = parseVaa(byteArray);

    if (vaaParsed) {
      return {
        emitterAddress: vaaParsed.emitterAddress.toString("hex"),
        emitterChain: vaaParsed.emitterChain,
        sequence: Number(vaaParsed.sequence),
      };
    }
  }
}

type VaaInformation = {
  emitterChain?: number;
  emitterAddress?: string;
  sequence?: number;
  consistencyLevel?: number;
  nonce?: number;
};

type VAA = {
  vaa: string;
};
