import { describe, it, expect } from "@jest/globals";
import { evmRedeemedTransactionFoundMapper } from "../../../../src/infrastructure/mappers/evm/evmRedeemedTransactionFoundMapper";
import { HandleEvmTransactions } from "../../../../src/domain/actions";

const address = "0xf890982f9310df57d00f659cf4fd87e65aded8d7";
const topic = "0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e";
const txHash = "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d";

const handler = new HandleEvmTransactions(
  {
    filter: { addresses: [address], topics: [topic] },
    abi: "event Delivery(address indexed recipientContract, uint16 indexed sourceChain, uint64 indexed sequence, bytes32 deliveryVaaHash, uint8 status, uint256 gasUsed, uint8 refundStatus, bytes additionalStatusInfo, bytes overridesInfo)",
  },
  evmRedeemedTransactionFoundMapper,
  async () => {}
);

describe("evmRedeemedTransactionFoundMapper", () => {
  it("should be able to map log to evmRedeemedTransactionFoundMapper without vaaInformation", async () => {
    // When
    const [result] = await handler.handle([
      {
        blockHash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        blockNumber: 0x11ec2bcn,
        chainId: 1,
        from: "0xfb070adcd21361a3946a0584dc84a7b89faa68e3",
        gas: "0x14485",
        gasPrice: "0xfc518561e",
        hash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        input:
          "0xc68785190000000000000000000000000000000000000000000000000000000000000001637651ef71f834be28b8fab1dce9c228c2fe1813831bbc3673cfd3abde6dbb3d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080420000",
        maxFeePerGas: "0x1610f75b9a",
        maxPriorityFeePerGas: "0x5f5e100",
        nonce: "0x1",
        r: "0xf5794b0970386d73b693b17f147fae0427db278e951e45465ac2c9835537e5a9",
        s: "0x6dccc8cfee216bc43a9d66525fa94905da234ad32d6cc3220845bef78f25dd42",
        status: "0x1",
        timestamp: 1702663079,
        to: "0xf890982f9310df57d00f659cf4fd87e65aded8d7",
        transactionIndex: "0x6f",
        type: "0x2",
        v: "0x1",
        value: "0x5b09cd3e5e90000",
        environment: "testnet",
        chain: "ethereum",
        logs: [
          {
            address: "0xf890982f9310df57d00f659cf4fd87e65aded8d7",
            topics: [
              "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759b7e0c00a",
              "0x0000000000000000000000000000000000000000000000000000000000000017",
              "0x0000000000000000000000002703483b1a5a7c577e8680de9df8be03c6f30e3c",
              "0x000000000000000000000000000000000000000000000000000000000000250f",
            ],
          },
        ],
      },
    ]);

    // Then
    expect(result.name).toBe("transfer-redeemed");
    expect(result.chainId).toBe(1);
    expect(result.txHash).toBe(txHash);
    expect(result.blockHeight).toBe(18793148n);
    expect(result.attributes.blockNumber).toBe(18793148n);
    expect(result.attributes.from).toBe("0xfb070adcd21361a3946a0584dc84a7b89faa68e3");
    expect(result.attributes.to).toBe("0xf890982f9310df57d00f659cf4fd87e65aded8d7");
    expect(result.attributes.methodsByAddress).toBe("MethodCompleteTransfer");
    expect(result.attributes.emitterChain).toBe(undefined);
    expect(result.attributes.emitterAddress).toBe(undefined);
    expect(result.attributes.sequence).toBe(undefined);
  });

  it("should be able to map log to evmRedeemedTransactionFoundMapper with vaaInformation", async () => {
    // When
    const [result] = await handler.handle([
      {
        blockHash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        blockNumber: 0x11ec2bcn,
        chainId: 1,
        from: "0xfb070adcd21361a3946a0584dc84a7b89faa68e3",
        gas: "0x14485",
        gasPrice: "0xfc518561e",
        hash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        input:
          "0xc68785190000000000000000000000000000000000000000000000000000000000000001637651ef71f834be28b8fab1dce9c228c2fe1813831bbc3673cfd3abde6dbb3d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080420000",
        maxFeePerGas: "0x1610f75b9a",
        maxPriorityFeePerGas: "0x5f5e100",
        nonce: "0x1",
        r: "0xf5794b0970386d73b693b17f147fae0427db278e951e45465ac2c9835537e5a9",
        s: "0x6dccc8cfee216bc43a9d66525fa94905da234ad32d6cc3220845bef78f25dd42",
        status: "0x1",
        timestamp: 1702663079,
        to: "0xf890982f9310df57d00f659cf4fd87e65aded8d7",
        transactionIndex: "0x6f",
        type: "0x2",
        v: "0x1",
        value: "0x5b09cd3e5e90000",
        environment: "testnet",
        chain: "ethereum",
        logs: [
          {
            address: "0xf890982f9310df57d00f659cf4fd87e65aded8d7",
            topics: [
              "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e",
              "0x0000000000000000000000000000000000000000000000000000000000000017",
              "0x0000000000000000000000002703483b1a5a7c577e8680de9df8be03c6f30e3c",
              "0x000000000000000000000000000000000000000000000000000000000000250f",
            ],
          },
        ],
      },
    ]);

    // Then
    expect(result.name).toBe("transfer-redeemed");
    expect(result.chainId).toBe(1);
    expect(result.txHash).toBe(txHash);
    expect(result.blockHeight).toBe(18793148n);
    expect(result.attributes.blockNumber).toBe(18793148n);
    expect(result.attributes.from).toBe("0xfb070adcd21361a3946a0584dc84a7b89faa68e3");
    expect(result.attributes.to).toBe("0xf890982f9310df57d00f659cf4fd87e65aded8d7");
    expect(result.attributes.methodsByAddress).toBe("MethodCompleteTransfer");
    expect(result.attributes.emitterChain).toBe(23);
    expect(result.attributes.emitterAddress).toBe(
      "0000000000000000000000002703483B1A5A7C577E8680DE9DF8BE03C6F30E3C"
    );
    expect(result.attributes.sequence).toBe(9487);
  });
});
