
import { methodNameByAddressMapper } from "../../../../src/infrastructure/mappers/evm/methodNameByAddressMapper";
import { describe, it, expect } from "@jest/globals";

describe("methodNameByAddressMapper", () => {
  it("should be throw error because cannot find method name in testnet environment", async () => {
    // Given
    const transaction = getTransactions(
      "0xF890982f9310df57d00f659cf4fd87e65adEd8d7",
      "0xc65465587851912312421412124124"
    );
    const environment = "testnet";
    const chain = "ethereum";

    // When
    const result = methodNameByAddressMapper(chain, environment, transaction);

    // Then
    expect(result).toBeUndefined();
  });

  it("should be return a method name in testnet environment", async () => {
    // Given
    const transaction = getTransactions(
      "0xF890982f9310df57d00f659cf4fd87e65adEd8d7",
      "0xc687851912312421412124124"
    );
    const environment = "testnet";
    const chain = "ethereum";

    // When
    const result = methodNameByAddressMapper(chain, environment, transaction);

    // Then
    expect(result).toEqual({ method: "MethodCompleteTransfer", name: "transfer-redeemed" });
  });

  it("should be throw error because cannot find method name in in mainnet environment", async () => {
    // Given
    const transaction = getTransactions(
      "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
      "0xc65465587851912312421412124124"
    );
    const environment = "mainnet";
    const chain = "ethereum";

    // When
    const result = methodNameByAddressMapper(chain, environment, transaction);

    // Then
    expect(result).toBeUndefined();
  });

  it("should be return a method name in mainnet environment", async () => {
    // Given
    const transaction = getTransactions(
      "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
      "0xc687851912312421412124124"
    );
    const environment = "mainnet";
    const chain = "ethereum";

    // When
    const result = methodNameByAddressMapper(chain, environment, transaction);

    // Then
    expect(result).toEqual({ method: "MethodCompleteTransfer", name: "transfer-redeemed" });
  });
});

const getTransactions = (to: string, input: string) => {
  return {
    blockHash: "0x1359819238ea89f49c20e42eb5603bf0541589d838d971984b60c7cdb391d9c2",
    blockNumber: 0x11ec2bcn,
    chainId: 1,
    from: "0xfb070adcd21361a3946a0584dc84a7b89faa68e3",
    gas: "0x14485",
    gasPrice: "0xfc518561e",
    hash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
    input: input.toLowerCase(),
    maxFeePerGas: "0x1610f75b9a",
    maxPriorityFeePerGas: "0x5f5e100",
    methodsByAddress: undefined,
    nonce: "0x1",
    r: "0xf5794b0970386d73b693b17f147fae0427db278e951e45465ac2c9835537e5a9",
    s: "0x6dccc8cfee216bc43a9d66525fa94905da234ad32d6cc3220845bef78f25dd42",
    status: "0x1",
    timestamp: 1702663079,
    to: to.toLowerCase(),
    transactionIndex: "0x6f",
    type: "0x2",
    v: "0x1",
    value: "0x5b09cd3e5e90000",
    environment: "testnet",
    chain: "ethereum",
  };
};
