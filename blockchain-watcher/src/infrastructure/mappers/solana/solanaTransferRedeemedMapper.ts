import { solana, TransactionFoundEvent, InstructionFound } from "../../../domain/entities";
import { CompiledInstruction, MessageCompiledInstruction } from "../../../domain/entities/solana";
import { Connection, Commitment } from "@solana/web3.js";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";
import { configuration } from "../../config";
import { findProtocol } from "../contractsMapper";
import winston from "winston";
import bs58 from "bs58";

let logger: winston.Logger;
logger = winston.child({ module: "solanaTransferRedeemedMapper" });

const TRANSACTION_STATUS_COMPLETED = "completed";
const TRANSACTION_STATUS_FAILED = "failed";
const SOLANA_CHAIN = "solana";

const connection = new Connection(configuration.chains.solana.rpcs[0]);

export const solanaTransferRedeemedMapper = async (
  transaction: solana.Transaction,
  { programs, commitment }: SolanaTransferRedeemedMapperOpts
): Promise<TransactionFoundEvent<InstructionFound>[]> => {
  for (const programId in programs) {
    const instructionsData = programs[programId];
    const results = await processProgram(transaction, programId, instructionsData, commitment);
    if (results.length) {
      return results;
    }
  }

  return [];
};

const processProgram = async (
  transaction: solana.Transaction,
  programId: string,
  { instructions: instructionsData, vaaAccountIndex }: ProgramParams,
  commitment?: Commitment
) => {
  const chain = transaction.chain;
  if (!transaction || !transaction.blockTime) {
    throw new Error(
      `[${chain}]Block time is missing for tx ${transaction?.transaction?.signatures} in slot ${transaction?.slot}`
    );
  }
  if (transaction.meta?.err) {
    logger.info(
      `[${chain}] Ignoring tx ${transaction.transaction.signatures[0]} because it failed: ${transaction.meta.err}`
    );
    return [];
  }

  const message = transaction.transaction.message;
  const accountKeys = message.accountKeys;
  const programIdIndex = accountKeys.findIndex((i) => i === programId);
  const instructions = message.compiledInstructions;
  const innerInstructions =
    transaction.meta?.innerInstructions?.flatMap((i) =>
      i.instructions.map(normalizeCompileInstruction)
    ) || [];

  const whInstructions = innerInstructions
    .concat(instructions)
    .filter((i) => i.programIdIndex === programIdIndex);

  const results: TransactionFoundEvent<InstructionFound>[] = [];
  for (const instruction of whInstructions) {
    const hexData = Buffer.from(instruction.data).toString("hex");
    if (!instructionsData || !instructionsData.includes(hexData)) {
      continue;
    }

    const accountAddress = accountKeys[instruction.accountKeyIndexes[vaaAccountIndex]];
    const { message } = await getPostedMessage(connection, accountAddress, commitment);
    const { sequence, emitterAddress, emitterChain } = message || {};
    const txHash = transaction.transaction.signatures[0];
    const protocol = findProtocol(SOLANA_CHAIN, programId, hexData, txHash);
    const protocolMethod = protocol?.method ?? "unknown";
    const protocolType = protocol?.type ?? "unknown";
    const emitterAddressToHex = emitterAddress.toString("hex");

    logger.info(
      `[${chain}] Redeemed transaction info: [hash: ${txHash}][VAA: ${emitterChain}/${emitterAddressToHex}/${sequence}][protocol: ${protocolType}/${protocolMethod}]`
    );

    results.push({
      name: "transfer-redeemed",
      address: programId,
      chainId: transaction.chainId,
      txHash: txHash,
      blockHeight: BigInt(transaction.slot.toString()),
      blockTime: transaction.blockTime,
      attributes: {
        methodsByAddress: protocol?.method ?? "unknownInstruction",
        status: mappedStatus(transaction),
        emitterChain: emitterChain,
        emitterAddress: emitterAddressToHex,
        sequence: Number(sequence),
        protocol: protocolType,
        fee: transaction.meta?.fee,
        from: accountKeys[0], // signer,
        to: programId,
      },
    });
  }

  return results;
};

const mappedStatus = (transaction: solana.Transaction): string => {
  if (!transaction.meta || transaction.meta.err) TRANSACTION_STATUS_FAILED;
  return TRANSACTION_STATUS_COMPLETED;
};

const normalizeCompileInstruction = (
  instruction: CompiledInstruction | MessageCompiledInstruction
): MessageCompiledInstruction => {
  if ("accounts" in instruction) {
    return {
      accountKeyIndexes: instruction.accounts,
      data: bs58.decode(instruction.data),
      programIdIndex: instruction.programIdIndex,
    };
  } else {
    return instruction;
  }
};

export interface ProgramParams {
  instructions: string[];
  vaaAccountIndex: number;
}

export type SolanaTransferRedeemedMapperOpts = {
  programs: Record<string, ProgramParams>;
  commitment?: Commitment;
};
