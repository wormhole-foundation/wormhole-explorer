import { describe, it, expect } from "@jest/globals";
import { evmTransferRedeemedMapper } from "../../../src/infrastructure/mappers/evmTransferRedeemedMapper";
import { HandleEvmTransactions } from "../../../src/domain/actions";

const address = "0x3ee18b2214aff97000d974cf647e7c347e8fa585";
const topic = "0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169";
const txHash = "0x1359819238ea89f49c20e42eb5603bf0541589d838d971984b60c7cdb391d9c2";

const handler = new HandleEvmTransactions(
  {
    filter: { addresses: [address], topics: [topic] },
    abi: "event TransferRedeemed(uint16 indexed emitterChainId, bytes32 indexed emitterAddress, uint64 indexed sequence)",
  },
  evmTransferRedeemedMapper,
  async () => {}
);

describe("evmTransferRedeemedMapper", () => {
  it("should be able to map log to TransferRedeemedTransaction", async () => {
    const [result] = await handler.handle([
      {
        hash: "0x1359819238ea89f49c20e42eb5603bf0541589d838d971984b60c7cdb391d9c2",
        blockNumber: 0x11ec2bcn,
        chainId: 1,
        from: "0xfb070adcd21361a3946a0584dc84a7b89faa68e3",
        input:
          "0x9981509f0000000000000000000000000000000000000000000000000000000000000001637651ef71f834be28b8fab1dce9c228c2fe1813831bbc3673cfd3abde6dbb3d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080420000",
        methodsByAddress: "test",
        status: "0x1",
        to: "0x3ee18b2214aff97000d974cf647e7c347e8fa585",
        timestamp: 12313123,
      },
    ]);

    expect(result.name).toBe("transfer-redeemed");
    expect(result.chainId).toBe(1);
    expect(result.txHash).toBe(txHash);
    expect(result.blockHeight).toBe(18793148n);
    expect(result.attributes.status).toBe("0x1");
    expect(result.attributes.blockNumber).toBe(0x11ec2bcn);
    expect(result.attributes.from).toBe("0xfb070adcd21361a3946a0584dc84a7b89faa68e3");
    expect(result.attributes.to).toBe("0x3ee18b2214aff97000d974cf647e7c347e8fa585");
    expect(result.attributes.methodsByAddress).toBe("test");
  });
});
