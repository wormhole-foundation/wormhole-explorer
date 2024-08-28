import { MessageTransmitter, MessageTransmitterIdl } from "./idl/messageTransmitter";
import { CircleMessageSent, LogFoundEvent, solana } from "../../../domain/entities";
import { CircleBurnMessage, CircleMessage } from "../evm/helpers/circle";
import { MessageProtocol, toCirceChain } from "../utils/circle";
import { normalizeCompileInstruction } from "./utils";
import { configuration } from "../../config";
import { Program, web3 } from "@coral-xyz/anchor";
import { CircleBridge } from "@wormhole-foundation/sdk-definitions";
import { Connection } from "@solana/web3.js";
import winston from "winston";

const WORMHOLE_CORE_CONTRACT = "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth";
const WORMHOLE_METHOD = "01";

const connection = new Connection(configuration.chains.solana.rpcs[0]);
const messageTransmitter = new Program<MessageTransmitter>(
  MessageTransmitterIdl,
  new web3.PublicKey("CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd"), // MessageTransmitter programId by circle
  { connection }
);

let logger: winston.Logger;
logger = winston.child({ module: "solanaLogCircleMessageSentMapper" });

export const solanaLogCircleMessageSentMapper = async (
  transaction: solana.Transaction,
  { programs, environment }: solanaLogCircleMessageSentMapperOpts
): Promise<LogFoundEvent<CircleMessageSent>[]> => {
  for (const programId in programs) {
    const instructionsData = programs[programId];
    const results = await processProgram(transaction, programId, instructionsData, environment);
    if (results && results.length > 0) {
      return results;
    }
  }
  return [];
};

const processProgram = async (
  tx: solana.Transaction,
  programId: string,
  { vaaAccountIndex }: ProgramParams,
  environment: string
) => {
  // Find the index of the programId in the account keys
  const programIdIndex = tx.transaction.message.accountKeys.findIndex((i) => i === programId);
  if (!programIdIndex || programIdIndex === -1) return undefined;

  const innerInstructions =
    tx.meta?.innerInstructions?.flatMap((i) => i.instructions.map(normalizeCompileInstruction)) ||
    [];
  if (!innerInstructions || innerInstructions.length == 0) return undefined;

  // Find the instruction with the index of the programId
  const innerInstruction = innerInstructions.find((ix) => ix.programIdIndex === programIdIndex);
  if (!innerInstruction) return undefined;

  // Find the account index of the sent message inner instruction (should be 1)
  const sentMessageAccountIndex = innerInstruction.accountKeyIndexes[vaaAccountIndex];
  if (!sentMessageAccountIndex) return undefined;

  // Find the public key of the sent message
  const sentMessageAccountPubKey = tx.transaction.message.accountKeys[sentMessageAccountIndex];
  const hash = tx.transaction.signatures[0];

  // Get the account content of the sent message account
  const accountContent = await mapAccountContent(hash, sentMessageAccountPubKey);
  if (!accountContent) return undefined;

  const results: LogFoundEvent<CircleMessageSent>[] = [];

  // Deserialize the account content to get the message data
  const [message, _] = CircleBridge.deserialize(accountContent!.message);
  const circleMessageSent = mappedCircleMessageSent(message, environment);
  const messageProtocol = mappedMessageProtocol(tx, innerInstructions);

  logger.info(
    `[solana] Circle message sent event info: [tx: ${hash}] [protocol: ${circleMessageSent.protocol} - ${messageProtocol}]`
  );

  results.push({
    name: "circle-message-sent",
    address: programId,
    chainId: 1,
    txHash: hash,
    blockHeight: BigInt(tx.slot),
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
  });
  return results;
};

const mapAccountContent = async (hash: string, sentMessageAccountPubKey: string) => {
  try {
    return await messageTransmitter.account.messageSent.fetch(sentMessageAccountPubKey);
  } catch (e) {
    logger.warn(
      `[solana] Error mapping account content [tx: ${hash}] [pubKey: ${sentMessageAccountPubKey}]. ${e}`
    );
  }
};

const mappedCircleMessageSent = (
  message: CircleMessage<CircleBurnMessage>,
  environment: string
) => {
  return {
    destinationCaller: message.destinationCaller.toString(),
    destinationDomain: toCirceChain(environment, message.destinationDomain),
    messageSender: message.payload.messageSender.toString(),
    mintRecipient: message.payload.mintRecipient.toString(),
    sourceDomain: toCirceChain(environment, message.sourceDomain),
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
  innerInstructions: web3.MessageCompiledInstruction[]
): string => {
  // Search the index of the programId wormhole core contract in the account keys if it exists
  const programIndexWH = tx.transaction.message.accountKeys.findIndex(
    (i) => i === WORMHOLE_CORE_CONTRACT
  );

  if (programIndexWH !== -1) {
    const innerInstruction = innerInstructions.find((ix) => ix.programIdIndex === programIndexWH);
    if (innerInstruction) {
      const hexData = Buffer.from(innerInstruction.data).toString("hex");
      if (hexData.startsWith(WORMHOLE_METHOD)) {
        return MessageProtocol.Wormhole;
      }
    }
  }
  return MessageProtocol.None;
};

interface ProgramParams {
  vaaAccountIndex: number;
}

type solanaLogCircleMessageSentMapperOpts = {
  programs: Record<string, ProgramParams>;
  environment: string;
};
