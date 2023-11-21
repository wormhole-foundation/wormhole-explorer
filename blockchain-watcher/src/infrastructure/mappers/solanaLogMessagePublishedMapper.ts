import { BigNumber } from "ethers";
import { solana, LogFoundEvent, LogMessagePublished } from "../../domain/entities";

export const solanaLogMessagePublishedMapper = (
  tx: solana.Transaction
): LogFoundEvent<LogMessagePublished>[] => {
  if (!tx || !tx.blockTime) {
    throw new Error(`Block time is missing for tx in slot ${tx?.slot} @ time ${tx?.blockTime}`);
  }

  /*
        
        const message = res.transaction.message;
        const accountKeys = isLegacyMessage(message)
          ? message.accountKeys
          : message.staticAccountKeys;
        const programIdIndex = accountKeys.findIndex((i) => i.toBase58() === WORMHOLE_PROGRAM_ID);
        const instructions = message.compiledInstructions;
        const innerInstructions =
          res.meta?.innerInstructions?.flatMap((i) =>
            i.instructions.map(normalizeCompileInstruction),
          ) || [];
        const whInstructions = innerInstructions
          .concat(instructions)
          .filter((i) => i.programIdIndex === programIdIndex);
        for (const instruction of whInstructions) {
          // skip if not postMessage instruction

          const instructionId = instruction.data;
          if (instructionId[0] !== 0x01) continue;

          const accountId = accountKeys[instruction.accountKeyIndexes[1]];
          const { message } = await getPostedMessage(connection, accountId.toBase58(), COMMITMENT);
          const {
            sequence,
            emitterAddress,
            emitterChain,
            submissionTime: timestamp,
            nonce,
            payload,
            consistencyLevel,
          } = message || {};

          // We store `blockNumber` with the slot number.
          const blockNumber = res.slot.toString();
          const chainId = emitterChain;
          const emitter = emitterAddress.toString('hex');
          const parsePayload = payload.toString('hex');
          const parseSequence = Number(sequence);
          const txHash = res.transaction.signatures[0];

        }

  */

  return [
    {
      name: "log-message-published",
      address: log.address, //
      chainId: 1,
      txHash: log.transactionHash,
      blockHeight: log.blockNumber,
      blockTime: log.blockTime,
      attributes: {
        sender: parsedArgs[0],
        sequence: (parsedArgs[1] as BigNumber).toNumber(),
        payload: parsedArgs[3],
        nonce: parsedArgs[2],
        consistencyLevel: parsedArgs[4],
      },
    },
  ];
};
