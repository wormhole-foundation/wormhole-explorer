import { describe, it, expect } from "@jest/globals";
import { evmTransferRedeemedMapper } from "../../../src/infrastructure/mappers/evmTransferRedeemedMapper";
import { HandleEvmLogs } from "../../../src/domain/actions";

const address = "0x98f3c9e6e3face36baad05fe09d375ef1464288b";
const topic = "0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169";
const txHash = "0xcbdefc83080a8f60cbde7785eb2978548fd5c1f7d0ea2c024cce537845d339c7";

const handler = new HandleEvmLogs(
  {
    filter: { addresses: [address], topics: [topic] },
    abi: "event TransferRedeemed(uint16 indexed emitterChainId, bytes32 indexed emitterAddress, uint64 indexed sequence)",
  },
  evmTransferRedeemedMapper,
  async () => {}
);

describe("evmTransferRedeemed", () => {
  it("should be able to map log to TransferRedeeemed", async () => {
    const [result] = await handler.handle([
      {
        chainId: 2,
        address,
        topics: [
          "0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169",
          "0x0000000000000000000000000000000000000000000000000000000000000001",
          "0xec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
          "0x0000000000000000000000000000000000000000000000000000000000052a3e",
        ],
        data: "0x",
        blockNumber: 18708192n,
        blockTime: 1699443287,
        transactionHash: txHash,
        transactionIndex: "0x3e",
        blockHash: "0x241fa85f3494c654d59859b46af586bd43f37ec434f5cf0018a53e46c42da393",
        logIndex: 216,
        removed: false,
      },
    ]);

    expect(result.name).toBe("transfer-redeemed");
    expect(result.chainId).toBe(2);
    expect(result.txHash).toBe(txHash);
    expect(result.blockHeight).toBe(18708192n);
    expect(result.blockTime).toBe(1699443287);

    expect(result.attributes.sequence).toBe(338494);
    expect(result.attributes.emitterAddress.toLowerCase()).toBe(
      "0xec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5"
    );
    expect(result.attributes.emitterChainId).toBe(1);
  });
});
