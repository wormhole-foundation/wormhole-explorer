import { solana, TransactionFoundEvent, InstructionFound } from "../../../domain/entities";
import { normalizeCompileInstruction } from "./utils";
import { Connection, Commitment } from "@solana/web3.js";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";
import { configuration } from "../../config";
import { findProtocol } from "../contractsMapper";
import winston from "winston";

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
    const programParams = programs[programId];
    const results = await processProgram(transaction, programId, programParams, commitment);
    if (results.length) {
      return results;
    }
  }
  return [];
};

const processProgram = async (
  transaction: solana.Transaction,
  programId: string,
  programParams: ProgramParams[],
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
      `[${chain}] Ignoring tx ${
        transaction.transaction.signatures[0]
      } because it failed: ${JSON.stringify(transaction.meta.err)}`
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
    const hexData = normalizeInstructionData(instruction.data, programId);
    const programParam = programParams.find((program) => program.instructions.includes(hexData));
    if (!programParam || !programParam.instructions || !programParam.vaaAccountIndex) {
      continue;
    }

    const accountAddress = accountKeys[instruction.accountKeyIndexes[programParam.vaaAccountIndex]];
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

const normalizeInstructionData = (data: Uint8Array, programId: string): string => {
  const hexData = Buffer.from(data).toString("hex");
  const mapper = PROGRAMS_ID[programId];
  if (mapper) {
    return mapper(hexData);
  }
  // Some instruction data contains only two characteres like token bridge: 02
  // and other contains 16 characteres like fast transfer or NTT
  return hexData.length > 2 ? hexData.slice(0, 16) : hexData;
};

const mapperHexDataWithTwoCaracteres: InstructionDataMaper = (hexData: string) => {
  return hexData.slice(0, 2);
};

const PROGRAMS_ID: Record<string, InstructionDataMaper> = {
  FC4eXxkyrMPTjiYUpp4EAnkmwMbQyZ6NDCh1kfLn6vsf: mapperHexDataWithTwoCaracteres, // Mayan
};

export interface ProgramParams {
  instructions: string[];
  vaaAccountIndex: number;
}

export type SolanaTransferRedeemedMapperOpts = {
  programs: Record<string, ProgramParams[]>;
  commitment?: Commitment;
};

type InstructionDataMaper = (hex: string) => string;
