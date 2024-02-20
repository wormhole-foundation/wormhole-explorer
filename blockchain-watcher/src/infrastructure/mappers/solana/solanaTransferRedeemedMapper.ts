import { solana, TransactionFoundEvent, InstructionFound } from "../../../domain/entities";
import { CompiledInstruction, MessageCompiledInstruction } from "../../../domain/entities/solana";
import { Protocol, contractsMapperConfig } from "../contractsMapper";
import { Connection, Commitment } from "@solana/web3.js";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";
import { configuration } from "../../config";
import winston from "winston";
import bs58 from "bs58";

let logger: winston.Logger;
logger = winston.child({ module: "solanaTransferRedeemedMapper" });

enum Instruction {
  CompleteNativeTransfer = 0x02,
  CompleteWrappedTransfer = 0x03,
  CompleteNativeWithPayload = 0x09,
  CompleteWrappedWithPayload = 0x0a,
}

const TRANSACTION_STATUS_COMPLETED = "completed";
const TRANSACTION_STATUS_FAILED = "failed";

const connection = new Connection(configuration.chains.solana.rpcs[0]);

export const solanaTransferRedeemedMapper = async (
  transaction: solana.Transaction,
  { programId, commitment }: { programId: string; commitment?: Commitment }
): Promise<TransactionFoundEvent<InstructionFound>[]> => {
  const chain = transaction.chain;
  if (!transaction || !transaction.blockTime) {
    throw new Error(
      `[${chain}]Block time is missing for tx ${transaction?.transaction?.signatures} in slot ${transaction?.slot}`
    );
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
    if (isNotACompleteTransferInstruction(instruction.data)) {
      continue;
    }

    const accountAddress = accountKeys[instruction.accountKeyIndexes[2]];
    const { message } = await getPostedMessage(connection, accountAddress, commitment);
    const { sequence, emitterAddress, emitterChain } = message || {};
    const txHash = transaction.transaction.signatures[0];
    const protocol = findProtocol(instruction, programIdIndex, programId, chain);

    logger.debug(
      `[${chain}}] Redeemed transaction info: [hash: ${txHash}][VAA: ${emitterChain}/${emitterAddress.toString(
        "hex"
      )}/${sequence}]`
    );

    results.push({
      name: "transfer-redeemed",
      address: programId,
      chainId: transaction.chainId,
      txHash: txHash,
      blockHeight: BigInt(transaction.slot.toString()),
      blockTime: transaction.blockTime,
      attributes: {
        methodsByAddress: protocol.method,
        status: mappedStatus(transaction),
        emitterChain: emitterChain,
        emitterAddress: emitterAddress.toString("hex"),
        sequence: Number(sequence),
        protocol: protocol.type,
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

const findProtocol = (
  instruction: solana.MessageCompiledInstruction,
  programIdIndex: number,
  programId: string,
  chain: string
): Protocol => {
  const unknownInstructionResponse = {
    method: "unknownInstruction",
    type: "unknown",
  };
  const data = instruction.data;

  if (!programIdIndex || instruction.programIdIndex != Number(programIdIndex) || data.length == 0) {
    return unknownInstructionResponse;
  }

  const methodId = data[0];

  for (const contract of contractsMapperConfig.contracts) {
    if (contract.chain === chain) {
      const foundProtocol = contract.protocols.find((protocol) =>
        protocol.addresses.some((addr) => addr.toLowerCase() === programId.toLowerCase())
      );
      const foundMethod = foundProtocol?.methods.find(
        (method) => method.methodId === String(methodId)
      );

      if (foundMethod && foundProtocol) {
        return {
          method: foundMethod.method,
          type: foundProtocol.type,
        };
      }
    }
  }

  return unknownInstructionResponse;
};

/**
 * Checks if the instruction is not to complete a transfer.
 * @param instructionId - the instruction id
 * @returns true if the instruction is valid, false otherwise
 */
const isNotACompleteTransferInstruction = (instructionId: Uint8Array): boolean => {
  return (
    instructionId[0] !== Instruction.CompleteNativeTransfer &&
    instructionId[0] !== Instruction.CompleteWrappedTransfer &&
    instructionId[0] !== Instruction.CompleteNativeWithPayload &&
    instructionId[0] !== Instruction.CompleteWrappedWithPayload
  );
};
