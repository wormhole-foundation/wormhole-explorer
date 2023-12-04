import { describe, it, expect } from "@jest/globals";
import { evmStandardRelayDelivered } from "../../../src/infrastructure/mappers/evmStandardRelayDelivered";
import { HandleEvmLogs } from "../../../src/domain/actions";

const address = "0x27428dd2d3dd32a4d7f7c497eaaa23130d894911";
const topic = "0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e";
const txHash = "0xcbdefc83080a8f60cbde7785eb2978548fd5c1f7d0ea2c024cce537845d339c7";

const handler = new HandleEvmLogs(
  {
    filter: { addresses: [address], topics: [topic] },
    abi: "event Delivery(address indexed recipientContract, uint16 indexed sourceChain, uint64 indexed sequence, bytes32 deliveryVaaHash, uint8 status, uint256 gasUsed, uint8 refundStatus, bytes additionalStatusInfo, bytes overridesInfo)",
  },
  evmStandardRelayDelivered,
  async () => {}
);

describe("evmStandardRelayDelivered", () => {
  it("should be able to map log to TransferRedeeemed", async () => {
    const [result] = await handler.handle([
      {
        chainId: 2,
        address,
        blockTime: 1699443287,
        transactionHash: txHash,
        topics: [
          "0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e",
          "0x000000000000000000000000f80cf52922b512b22d46aa8916bd7767524305d9",
          "0x000000000000000000000000000000000000000000000000000000000000001e",
          "0x0000000000000000000000000000000000000000000000000000000000000900",
        ],
        data: "0xf29cac97156fa11c205eda95c0655e4a6e2a9c247245bab4d3d8257c41fc11d200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000013a89000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        blockNumber: 18708316n,
        transactionIndex: "0x3b",
        blockHash: "0x8c55cbd97c96f8322bed4d1790c7ac4a84b1cff46c157bf86fc35eb5886be451",
        logIndex: 5,
        removed: false,
      },
    ]);

    expect(result.name).toBe("standard-relay-delivered");
    expect(result.chainId).toBe(2);
    expect(result.txHash).toBe(txHash);
    expect(result.blockHeight).toBe(18708316n);
    expect(result.blockTime).toBe(1699443287);

    expect(result.attributes.recipientContract.toLowerCase()).toBe(
      "0xf80cf52922b512b22d46aa8916bd7767524305d9"
    );
    expect(result.attributes.sourceChain).toBe(30);
    expect(result.attributes.sequence).toBe(2304);
    expect(result.attributes.deliveryVaaHash.toLowerCase()).toBe(
      "0xf29cac97156fa11c205eda95c0655e4a6e2a9c247245bab4d3d8257c41fc11d2"
    );
    expect(result.attributes.status).toBe(0);
    expect(result.attributes.gasUsed).toBe(80521);
    expect(result.attributes.refundStatus).toBe(0);
  });
});
