import { describe, it, expect } from "@jest/globals";
import { evmLogMessagePublishedMapper } from "../../../src/infrastructure/mappers/evmLogMessagePublishedMapper";
import { HandleEvmLogs } from "../../../src/domain/actions";

const address = "0x98f3c9e6e3face36baad05fe09d375ef1464288b";
const topic = "0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2";
const txHash = "0xcbdefc83080a8f60cbde7785eb2978548fd5c1f7d0ea2c024cce537845d339c7";

const handler = new HandleEvmLogs(
  {
    filter: { addresses: [address], topics: [topic] },
    abi: "event LogMessagePublished(address indexed sender, uint64 sequence, uint32 nonce, bytes payload, uint8 consistencyLevel)",
  },
  evmLogMessagePublishedMapper,
  async () => {}
);

describe("evmLogMessagePublished", () => {
  it("should be able to map log to LogMessagePublished", async () => {
    const [result] = await handler.handle([
      {
        blockTime: 1699443287,
        blockNumber: 18521386n,
        blockHash: "0x894136d03446d47116319d59b5ec3190c05248e16c8728c2848bf7452732341c",
        address: "0x98f3c9e6e3face36baad05fe09d375ef1464288b",
        removed: false,
        data: "0x00000000000000000000000000000000000000000000000000000000000212b20000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000085010000000000000000000000000000000000000000000000000000000045be2810000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb480002f022f6b3e80ec1219065fee8e46eb34c1cfd056a8d52d93df2c7e0165eaf364b00010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        transactionHash: txHash,
        transactionIndex: "0x62",
        topics: [topic, "0x0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585"],
        logIndex: 0,
        chainId: 2,
      },
    ]);

    expect(result.name).toBe("log-message-published");
    expect(result.chainId).toBe(2);
    expect(result.txHash).toBe(
      "0xcbdefc83080a8f60cbde7785eb2978548fd5c1f7d0ea2c024cce537845d339c7"
    );
    expect(result.blockHeight).toBe(18521386n);
    expect(result.blockTime).toBe(1699443287);

    expect(result.attributes.sequence).toBe(135858);
    expect(result.attributes.sender.toLowerCase()).toBe(
      "0x3ee18b2214aff97000d974cf647e7c347e8fa585"
    );
    expect(result.attributes.nonce).toBe(0);
    expect(result.attributes.consistencyLevel).toBe(1);
  });
});
