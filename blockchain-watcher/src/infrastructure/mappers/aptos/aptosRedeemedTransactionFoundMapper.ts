import { TransactionFoundEvent } from "../../../domain/entities";
import { AptosTransaction } from "../../../domain/entities/aptos";
import { CHAIN_ID_APTOS } from "@certusone/wormhole-sdk";
import { findProtocol } from "../contractsMapper";
import { parseVaa } from "@certusone/wormhole-sdk";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "aptosRedeemedTransactionFoundMapper" });

const APTOS_CHAIN = "aptos";

export const aptosRedeemedTransactionFoundMapper = (
  tx: AptosTransaction
): TransactionFoundEvent | undefined => {
  const protocol = findProtocol(APTOS_CHAIN, tx.to, tx.type!, tx.hash);

  const vaaBuffer = Buffer.from(tx.payload?.arguments[0]?.substring(2), "hex");
  const vaa = parseVaa(vaaBuffer);

  const emitterAddress = vaa.emitterAddress.toString("hex");

  if (protocol && protocol.type && protocol.method) {
    logger.info(
      `[${APTOS_CHAIN}] Redeemed transaction info: [hash: ${tx.hash}][VAA: ${vaa.emitterChain}/${emitterAddress}/${vaa.sequence}]`
    );

    return {
      name: "transfer-redeemed",
      address: tx.to,
      blockHeight: tx.blockHeight,
      blockTime: vaa.timestamp,
      chainId: CHAIN_ID_APTOS,
      txHash: tx.hash,
      attributes: {
        from: tx.address,
        emitterChain: vaa.emitterChain,
        emitterAddress: emitterAddress,
        sequence: Number(vaa.sequence),
        status: tx.status === true ? TxStatus.Completed : TxStatus.Failed,
        protocol: protocol.method,
      },
    };
  }
};

enum TxStatus {
  Completed = "completed",
  Failed = "failed",
}
