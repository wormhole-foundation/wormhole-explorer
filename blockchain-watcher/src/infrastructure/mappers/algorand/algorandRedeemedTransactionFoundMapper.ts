import { TransactionFoundEvent } from "../../../domain/entities";
import { AlgorandTransaction } from "../../../domain/entities/algorand";
import { CHAIN_ID_APTOS } from "@certusone/wormhole-sdk";
import { findProtocol } from "../contractsMapper";
import { parseVaa } from "@certusone/wormhole-sdk";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "algorandRedeemedTransactionFoundMapper" });

const ALGORAND_CHAIN = "algorand";

export const algorandRedeemedTransactionFoundMapper = (
  transaction: AlgorandTransaction,
  filters: {
    applicationsIds: string;
    applicationAddress: string;
  }[]
): TransactionFoundEvent | undefined => {
  const applicationId = String(transaction.applicationId);

  const protocol = findProtocol(ALGORAND_CHAIN, applicationId, applicationId, transaction.hash);
  const vaaInformation = mappedVaaInformation(transaction.payload);

  if (!vaaInformation) {
    logger.warn(
      `[algorand] Cannot mapper vaa information: [hash: ${transaction.hash}][protocol: ${protocol.type}/${protocol.method}]`
    );
    return undefined;
  }

  const filter = filters.find((filter) => filter.applicationsIds === applicationId);

  const { emitterChain, emitterAddress, sequence } = vaaInformation;

  logger.info(
    `[${ALGORAND_CHAIN}] Redeemed transaction info: [hash: ${transaction.hash}][VAA: ${emitterChain}/${emitterAddress}/${sequence}]`
  );

  return {
    name: "transfer-redeemed",
    address: filter?.applicationAddress ?? applicationId,
    blockHeight: BigInt(transaction.blockNumber),
    blockTime: transaction.timestamp,
    chainId: CHAIN_ID_APTOS,
    txHash: transaction.hash,
    attributes: {
      from: transaction.sender,
      emitterChain: emitterChain,
      emitterAddress: emitterAddress,
      sequence: Number(sequence),
      status: TxStatus.Completed,
      protocol: protocol.method,
    },
  };
};

const mappedVaaInformation = (payload: string): VaaInformation | undefined => {
  if (payload) {
    const payloadToHex = Buffer.from(payload, "base64").toString("hex");
    const buffer = Buffer.from(payloadToHex, "hex");
    const vaa = parseVaa(buffer);

    return {
      emitterChain: vaa.emitterChain,
      emitterAddress: vaa.emitterAddress.toString("hex").toUpperCase(),
      sequence: Number(vaa.sequence),
    };
  }
};

type VaaInformation = {
  emitterChain: number;
  emitterAddress: string;
  sequence: number;
  formAddress?: string;
  toAddress?: string;
};

enum TxStatus {
  Completed = "completed",
  Failed = "failed",
}
