"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.makeSignedVAAsRowKey = exports.padUint64 = exports.MAX_UINT_64 = exports.padUint16 = exports.MAX_UINT_16 = exports.assertEnvironmentVariable = exports.sleep = void 0;
async function sleep(timeout) {
    return new Promise((resolve) => setTimeout(resolve, timeout));
}
exports.sleep = sleep;
const assertEnvironmentVariable = (varName) => {
    if (varName in process.env)
        return process.env[varName];
    throw new Error(`Missing required environment variable: ${varName}`);
};
exports.assertEnvironmentVariable = assertEnvironmentVariable;
exports.MAX_UINT_16 = '65535';
const padUint16 = (s) => s.padStart(exports.MAX_UINT_16.length, '0');
exports.padUint16 = padUint16;
exports.MAX_UINT_64 = '18446744073709551615';
const padUint64 = (s) => s.padStart(exports.MAX_UINT_64.length, '0');
exports.padUint64 = padUint64;
// make a bigtable row key for the `signedVAAs` table
const makeSignedVAAsRowKey = (chain, emitter, sequence) => `${(0, exports.padUint16)(chain.toString())}/${emitter}/${(0, exports.padUint64)(sequence)}`;
exports.makeSignedVAAsRowKey = makeSignedVAAsRowKey;
//# sourceMappingURL=utils.js.map