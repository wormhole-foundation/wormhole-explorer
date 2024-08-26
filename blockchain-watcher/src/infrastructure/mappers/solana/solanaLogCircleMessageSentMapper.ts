import { MessageTransmitter, MessageTransmitterIdl } from "./idl/messageTransmitter";
import { CircleMessageSent, LogFoundEvent, solana } from "../../../domain/entities";
import { CircleBurnMessage, CircleMessage } from "../evm/helpers/circle";
import { MessageProtocol, toCirceChain } from "../utils/circle";
import { normalizeCompileInstruction } from "./utils";
import { Commitment, Connection } from "@solana/web3.js";
import { HandleSolanaTxConfig } from "../../../domain/actions/solana/HandleSolanaTransactions";
import { configuration } from "../../config";
import { Program, web3 } from "@coral-xyz/anchor";
import { CircleBridge } from "@wormhole-foundation/sdk-definitions";
import winston from "winston";

const connection = new Connection(configuration.chains.solana.rpcs[0]);
const messageTransmitter = new Program<MessageTransmitter>(
  MessageTransmitterIdl,
  new web3.PublicKey("CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd"),
  { connection }
);

let logger: winston.Logger;
logger = winston.child({ module: "solanaLogCircleMessageSentMapper" });

export const solanaLogCircleMessageSentMapper = async (
  transaction: solana.Transaction,
  { programs }: solanaLogCircleMessageSentMapperOpts,
  cfg?: HandleSolanaTxConfig
): Promise<LogFoundEvent<CircleMessageSent> | undefined> => {
  const instructionsData = programs[0];
  const result = await processProgram(transaction, "programId[0]", instructionsData, cfg); // TODO: Change this to the actual programId
  if (result) {
    return result;
  }
  return undefined;
};

const processProgram = async (
  tx: solana.Transaction,
  programId: string,
  { instructions: instructionsData, vaaAccountIndex }: ProgramParams,
  cfg?: HandleSolanaTxConfig
) => {
  // Search for Circle send message ix
  const circleProgramIndex = tx.transaction.message.accountKeys.findIndex((i) => i === programId);
  if (!circleProgramIndex) return undefined;

  const innerInstructions =
    tx.meta?.innerInstructions?.flatMap((i) => i.instructions.map(normalizeCompileInstruction)) ||
    [];
  if (!innerInstructions || innerInstructions.length == 0) return undefined;

  const circleIx = innerInstructions.find((ix) => ix.programIdIndex === circleProgramIndex);
  if (!circleIx) return undefined;

  // Look for the sent message account
  const sentMessageAccountIndex = circleIx.accountKeyIndexes[vaaAccountIndex]; // 1
  const sentMessageAccountPubKey = tx.transaction.message.accountKeys[sentMessageAccountIndex]; // C83xWSWFhV3T5aGnnA752VLoAfTT5Fax2WTMMJfzVpuA
  const accountContent = await messageTransmitter.account.messageSent.fetch(
    sentMessageAccountPubKey
  );

  // Deserialize raw message bytes
  const [message, _] = CircleBridge.deserialize(accountContent.message);
  const messageProtocol = mappedMessageProtocol(tx, programId, innerInstructions);
  const circleMessageSent = mappedCircleMessageSent(message, cfg!);
  const hash = tx.transaction.signatures[0];

  logger.info(
    `[solana] Circle message sent event info: [tx: ${hash}] [protocol: ${circleMessageSent.protocol} - ${messageProtocol}]`
  );

  return {
    name: "circle-message-sent",
    address: programId,
    chainId: 1,
    txHash: hash,
    blockHeight: BigInt(tx.slot.toString()),
    blockTime: tx.blockTime!,
    attributes: {
      ...circleMessageSent,
      txHash: hash,
    },
    tags: {
      destinationDomain: circleMessageSent.destinationDomain,
      messageProtocol: messageProtocol,
      sourceDomain: circleMessageSent.sourceDomain,
      protocol: circleMessageSent.protocol,
      sender: circleMessageSent.sender,
    },
  };
};

const mappedCircleMessageSent = (
  message: CircleMessage<CircleBurnMessage>,
  cfg: HandleSolanaTxConfig
) => {
  return {
    destinationCaller: message.destinationCaller.toString(),
    destinationDomain: toCirceChain(cfg.environment, message.destinationDomain),
    messageSender: message.payload.messageSender.toString(),
    mintRecipient: message.payload.mintRecipient.toString(),
    sourceDomain: toCirceChain(cfg.environment, message.sourceDomain),
    burnToken: message.payload.burnToken.toString(),
    recipient: message.recipient.toString(),
    sender: message.sender.toString(),
    amount: message.payload.amount,
    nonce: message.nonce,
    protocol: "cctp",
  };
};

const mappedMessageProtocol = (
  tx: solana.Transaction,
  whProgramId: string,
  innerInstructions: web3.MessageCompiledInstruction[]
): string => {
  const programIndex = tx.transaction.message.accountKeys.findIndex((i) => i === whProgramId);
  const innerInstruction = innerInstructions.find((ix) => ix.programIdIndex === programIndex);
  return innerInstruction ? MessageProtocol.Wormhole : MessageProtocol.None;
};

interface ProgramParams {
  instructions: string[];
  vaaAccountIndex: number;
}

type solanaLogCircleMessageSentMapperOpts = {
  programs: Record<string, ProgramParams>;
  commitment?: Commitment;
};
