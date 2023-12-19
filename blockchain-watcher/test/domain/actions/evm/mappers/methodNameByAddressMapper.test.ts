import { methodNameByAddressMapper } from "../../../../../src/domain/actions/evm/mappers/methodNameByAddressMapper";
import { describe, it, expect } from "@jest/globals";

describe("methodNameByAddressMapper", () => {
  it("should be return an empty method name in testnet environment", async () => {
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
    expect(result).toEqual("");
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
    expect(result).toEqual("MethodCompleteTransfer");
  });

  it("should be return an empty method name in mainnet environment", async () => {
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
    expect(result).toEqual("");
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
    expect(result).toEqual("MethodCompleteTransfer");
  });
});

const getTransactions = (to: string, input: string) => {
  return {
    hash: "0x1359819238ea89f49c20e42eb5603bf0541589d838d971984b60c7cdb391d9c2",
    blockNumber: 0x11ec2bcn,
    chainId: "0x2",
    from: "0xfb070adcd21361a3946a0584dc84a7b89faa68e3",
    input: input.toLowerCase(),
    methodsByAddress: "",
    status: "0x0",
    to: to.toLowerCase(),
    timestamp: 12313123,
  };
};
