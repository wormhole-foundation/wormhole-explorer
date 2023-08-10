export declare function sleep(timeout: number): Promise<unknown>;
export declare const assertEnvironmentVariable: (varName: string) => string;
export declare const MAX_UINT_16 = "65535";
export declare const padUint16: (s: string) => string;
export declare const MAX_UINT_64 = "18446744073709551615";
export declare const padUint64: (s: string) => string;
export declare const makeSignedVAAsRowKey: (chain: number, emitter: string, sequence: string) => string;
