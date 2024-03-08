import { TransactionFoundEvent } from "../../../domain/entities";
import { TransactionsByVersion } from "../../repositories/aptos/AptosJsonRPCBlockRepository";
import { findProtocol } from "../contractsMapper";
import { CHAIN_ID_APTOS } from "@certusone/wormhole-sdk";
import { TxStatus } from "../../../domain/entities/aptos";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "aptosRedeemedTransactionFoundMapper" });

const APTOS_CHAIN = "aptos";

export const aptosRedeemedTransactionFoundMapper = (
  tx: TransactionsByVersion
): TransactionFoundEvent | undefined => {
  const emitterAddress = tx.sender;

  const protocol = findProtocol(APTOS_CHAIN, tx.address, tx.type!, tx.hash);

  logger.info(
    `[${APTOS_CHAIN}] Redeemed transaction info: [hash: ${tx.hash}][VAA: ${tx.emitterChain}/${emitterAddress}/${tx.sequence}]`
  );

  if (protocol && protocol.type && protocol.method) {
    return {
      name: "transfer-redeemed",
      address: tx.address,
      blockHeight: tx.blockHeight,
      blockTime: tx.blockTime,
      chainId: CHAIN_ID_APTOS,
      txHash: tx.hash,
      attributes: {
        from: tx.sender,
        emitterChain: tx.emitterChain,
        emitterAddress: emitterAddress,
        sequence: Number(tx.sequence),
        status: tx?.status === true ? TxStatus.Confirmed : TxStatus.Failed,
        protocol: protocol.method,
      },
    };
  }
};
