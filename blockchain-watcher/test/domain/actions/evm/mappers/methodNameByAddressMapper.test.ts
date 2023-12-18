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
    hash: "dasdasfpialsfijlasfsahuf",
    from: "0x3ee123456786797000d974cf647e7c347e8fa585",
    to: to.toLowerCase(),
    blockNumber: 1n,
    topics: [],
    input: input.toLowerCase(),
    data: "",
    chainId: 1,
  };
};
