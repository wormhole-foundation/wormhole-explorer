import { TransactionFoundEvent } from "../../../domain/entities";
import { findProtocol } from "../contractsMapper";
import { parseVaa } from "@certusone/wormhole-sdk";
import { NearTransaction } from "../../../domain/entities/near";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "nearRedeemedTransactionFoundMapper" });

export const nearRedeemedTransactionFoundMapper = (
  transaction: NearTransaction
): TransactionFoundEvent | undefined => {
  logger.info(`[${1}] Redeemed transaction info: [hash: ${transaction.hash}][VAA: ${1}/${1}/${1}]`);

  return {
    name: "transfer-redeemed",
    address: address,
    blockHeight: transaction.blockHeight,
    blockTime: vaa.timestamp,
    chainId: CHAIN_ID_APTOS,
    txHash: transaction.hash,
    attributes: {
      from: address,
      emitterChain: vaa.emitterChain,
      emitterAddress: emitterAddress,
      sequence: Number(vaa.sequence),
      status: TxStatus.Completed, // TODO
      protocol: protocol.method,
    },
  };
};

enum TxStatus {
  Completed = "completed",
  Failed = "failed",
}
