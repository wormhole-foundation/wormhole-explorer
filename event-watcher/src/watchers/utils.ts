import {
  ChainName,
  CosmWasmChainName,
  EVMChainName,
} from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
import { Other, Payload, serialiseVAA, VAA } from '@certusone/wormhole-sdk';
import { AlgorandWatcher } from './AlgorandWatcher';
import { AptosWatcher } from './AptosWatcher';
import { ArbitrumWatcher } from './ArbitrumWatcher';
import { BSCWatcher } from './BSCWatcher';
import { CosmwasmWatcher } from './CosmwasmWatcher';
import { EVMWatcher } from './EVMWatcher';
import { InjectiveExplorerWatcher } from './InjectiveExplorerWatcher';
import { MoonbeamWatcher } from './MoonbeamWatcher';
import { NearWatcher } from './NearWatcher';
import { PolygonWatcher } from './PolygonWatcher';
import { SolanaWatcher } from './SolanaWatcher';
import { TerraExplorerWatcher } from './TerraExplorerWatcher';
import { SuiWatcher } from './SuiWatcher';
import { makeVAAInput, WatcherOptionTypes } from './types';
import { SeiExplorerWatcher } from './SeiExplorerWatcher';
import { checkIfDateIsInMilliseconds } from '../utils/date';
import { WormchainWatcher } from './WormchainWatcher';

export function makeFinalizedWatcher(chainName: ChainName): WatcherOptionTypes {
  if (chainName === 'solana') {
    return new SolanaWatcher();
  } else if (['ethereum', 'karura', 'acala'].includes(chainName)) {
    return new EVMWatcher(chainName as EVMChainName, 'finalized');
  } else if (chainName === 'bsc') {
    return new BSCWatcher();
  } else if (chainName === 'polygon') {
    return new PolygonWatcher();
  } else if (
    ['avalanche', 'oasis', 'fantom', 'klaytn', 'celo', 'optimism', 'base'].includes(chainName)
  ) {
    return new EVMWatcher(chainName as EVMChainName);
  } else if (chainName === 'algorand') {
    return new AlgorandWatcher();
  } else if (chainName === 'moonbeam') {
    return new MoonbeamWatcher();
  } else if (chainName === 'arbitrum') {
    return new ArbitrumWatcher();
  } else if (chainName === 'aptos') {
    return new AptosWatcher();
  } else if (chainName === 'near') {
    return new NearWatcher();
  } else if (chainName === 'injective') {
    return new InjectiveExplorerWatcher();
  } else if (chainName === 'terra') {
    return new TerraExplorerWatcher('terra');
  } else if (['terra2', 'xpla'].includes(chainName)) {
    return new CosmwasmWatcher(chainName as CosmWasmChainName);
  } else if (chainName === 'sui') {
    return new SuiWatcher();
  } else if (chainName === 'sei') {
    return new SeiExplorerWatcher();
  } else if (chainName === 'wormchain') {
    return new WormchainWatcher();
  } else {
    throw new Error(`Attempted to create Event Watcher for unsupported chain: [${chainName}]`);
  }
}

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
