import { Program, web3 } from "@coral-xyz/anchor";
import { Commitment, Connection } from "@solana/web3.js";
import { CircleBridge } from "@wormhole-foundation/sdk-definitions";
import winston from "winston";
import { CircleMessageSent, LogFoundEvent, solana } from "../../../domain/entities";
import { configuration } from "../../config";
import { MessageTransmitter, MessageTransmitterIdl } from "./idl/messageTransmitter";
import { normalizeCompileInstruction } from "./utils";

const connection = new Connection(configuration.chains.solana.rpcs[0]); // TODO: should be better to inject this to improve testability
const messageTransmitter = new Program<MessageTransmitter>(
  MessageTransmitterIdl,
  new web3.PublicKey("asd"),
  { connection }
);

let logger: winston.Logger;
logger = winston.child({ module: "solanaLogMessagePublishedMapper" });

interface SolanacircleMessageSentMapperOpts {
  circleProgramId: string;
  whProgramId: string;
  commitment?: Commitment;
}

export const solanaLogMessagePublishedMapper = async (
  tx: solana.Transaction,
  { circleProgramId, whProgramId, commitment }: SolanacircleMessageSentMapperOpts
): Promise<LogFoundEvent<CircleMessageSent> | undefined> => {
  const innerInstructions =
    tx.meta?.innerInstructions?.flatMap((i) => i.instructions.map(normalizeCompileInstruction)) ||
    [];

  // Search for Circle send message ix
  // TODO: check the ix data is for send message
  const circleProgramIndex = tx.transaction.message.accountKeys.findIndex(
    (i) => i === circleProgramId
  );
  const circleIx = innerInstructions.find((ix) => ix.programIdIndex === circleProgramIndex);
  if (!circleIx) return undefined;

  // Search for Wormhole publish message ix
  let protocol = "";
  const whProgramIndex = tx.transaction.message.accountKeys.findIndex((i) => i === whProgramId);
  const whIx = innerInstructions.find((ix) => ix.programIdIndex === whProgramIndex);
  if (whIx) protocol = "wormhole";

  // look for the sent message account
  const sentMessageAccountIndex = circleIx.accountKeyIndexes[3]; // 1
  const sentMessageAccountPubKey = tx.transaction.message.accountKeys[sentMessageAccountIndex]; // C83xWSWFhV3T5aGnnA752VLoAfTT5Fax2WTMMJfzVpuA
  const accountContent = await messageTransmitter.account.messageSent.fetch(
    sentMessageAccountPubKey
  ); // deserialize raw message bytes
  const [message, _] = CircleBridge.deserialize(accountContent.message);

  const sourceDomain = message.sourceDomain;
  const destinationDomain = message.destinationDomain;
  const messageProtocol = "";
  const sender = "";

  return {
    name: "circle-message-sent",
    address: programId,
    chainId: 1,
    txHash: tx.transaction.signatures[0],
    blockHeight: BigInt(tx.slot.toString()),
    blockTime: tx.blockTime!,
    attributes: {
      sourceDomain,
      destinationDomain,
      nonce: message.nonce,
      sender: message.sender,
      recipient: message.recipient,
      destinationCaller: message.destinationCaller,
      burnToken: message.payload.burnToken,
      mintRecipient: message.payload.mintRecipient,
      amount: message.payload.amount,
      messageSender: message.payload.messageSender,
      txHash: tx.transaction.signatures[0],
    },
    tags: {
      destinationDomain,
      messageProtocol,
      sourceDomain,
      protocol,
      sender,
    },
  };
};
