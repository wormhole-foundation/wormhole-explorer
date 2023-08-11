import { ChainId, ChainName } from '@certusone/wormhole-sdk/lib/cjs/utils/consts';
export declare const INITIAL_DEPLOYMENT_BLOCK_BY_CHAIN: {
    [key in ChainName]?: string;
};
export declare const TOKEN_BRIDGE_EMITTERS: {
    [key in ChainName]?: string;
};
export declare const isTokenBridgeEmitter: (chain: ChainId | ChainName, emitter: string) => boolean;
export declare const NFT_BRIDGE_EMITTERS: {
    [key in ChainName]?: string;
};
export declare const isNFTBridgeEmitter: (chain: ChainId | ChainName, emitter: string) => boolean;
