import { Other, Payload, VAA } from '@certusone/wormhole-sdk';
import { checkIfDateIsInMilliseconds } from './date';

export type makeVAAInput = {
  timestamp: Date | string | number;
  nonce: number;
  emitterChain: number;
  emitterAddress: string;
  sequence: number;
  consistencyLevel: number;
  payloadAsHex: string;
};

export const makeSerializedVAA = async ({
  timestamp,
  nonce,
  emitterChain,
  emitterAddress,
  sequence,
  payloadAsHex,
  consistencyLevel,
}: makeVAAInput) => {
  // We use `Other` because we need to pass the payload as a hex string
  const PAYLOAD_TYPE = 'Other';
  let parsedTimestamp = timestamp as number;

  if (timestamp instanceof Date) {
    parsedTimestamp = timestamp.getTime();
  } else {
    parsedTimestamp = new Date(timestamp).getTime();
  }

  if (checkIfDateIsInMilliseconds(parsedTimestamp)) {
    parsedTimestamp = parsedTimestamp / 1000;
  }

  const vaaObject: VAA<Payload | Other> = {
    version: 1,
    guardianSetIndex: 0,
    signatures: [],
    timestamp: Math.floor(parsedTimestamp),
    nonce,
    emitterChain,
    emitterAddress,
    sequence: BigInt(sequence),
    consistencyLevel,
    payload: {
      type: PAYLOAD_TYPE,
      hex: payloadAsHex,
    },
  };

  // @ts-ignore: We pass in a VAA<Payload | Other> but the function expects a VAA<Payload>
  return serialiseVAA(vaaObject);
};
