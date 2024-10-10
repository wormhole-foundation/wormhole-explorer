import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { expect, describe, it } from "@jest/globals";
import { PublicKey } from "@solana/web3.js";
import { solana } from "../../../src/domain/entities";
import { Web3SolanaSlotRepository } from "../../../src/infrastructure/repositories";
import { InstrumentedConnectionWrapper } from "../../../src/infrastructure/rpc/http/InstrumentedConnectionWrapper";

describe("Web3SolanaSlotRepository", () => {
  describe("healthCheck", () => {
    it("should be able to validate rpcs", async () => {
      // Given
      const connectionMock = {
        rpcEndpoint: "http://solanafake.com",
        getSlot: () => Promise.resolve(100),
      };
      const poolMock = {
        get: () => connectionMock,
        getProviders: () => [
          new InstrumentedConnectionWrapper("http://solanafake.com", "finalized", 100, "solana"),
        ],
        setProviders: () => {},
      };
      const repository = new Web3SolanaSlotRepository(poolMock as any);

      // When
      const result = await repository.healthCheck("solana", "finalized", 100n);

      // Then
      expect(result).toBeInstanceOf(Array);
      expect(result[0].isHealthy).toEqual(true);
      expect(result[0].height).toEqual(100n);
      expect(result[0].url).toEqual("http://solanafake.com");
      expect(result[0].latency).toBeDefined();
    });
  });

  describe("getLatestSlot", () => {
    it("should return the latest slot number", async () => {
      // Given
      const connectionMock = {
        rpcEndpoint: "http://solanafake.com",
        getSlot: () => Promise.resolve(100),
      };
      const poolMock = {
        get: () => connectionMock,
      };
      const repository = new Web3SolanaSlotRepository(poolMock as any);

      // When
      const latestSlot = await repository.getLatestSlot("finalized");

      // Then
      expect(latestSlot).toBe(100);
    });
  });

  describe("getBlock", () => {
    it("should return a block for a given slot number", async () => {
      // Given
      const expected = {
        blockTime: 100,
        transactions: [
          {
            signature: "signature1",
            slot: 100,
            transaction: {
              message: {
                version: "legacy",
                accountKeys: [new PublicKey("3u8hJUVTA4jH1wYAyUur7FFZVQ8H635K3tSHHF4ssjQ5")],
                instructions: [],
                compiledInstructions: [],
              },
            },
          },
          {
            signature: "signature1",
            slot: 100,
            transaction: {
              message: {
                version: 0,
                staticAccountKeys: [new PublicKey("3u8hJUVTA4jH1wYAyUur7FFZVQ8H635K3tSHHF4ssjQ5")],
                instructions: [],
                compiledInstructions: [],
              },
            },
          },
        ],
      };
      const connectionMock = {
        rpcEndpoint: "http://solanafake.com",
        getBlock: (slot: number) => Promise.resolve(expected),
      };
      const poolMock = {
        get: () => connectionMock,
      };
      const repository = new Web3SolanaSlotRepository(poolMock as any);

      // When
      const block = (await repository.getBlock(100)).getValue();

      // Then
      expect(block.blockTime).toBe(expected.blockTime);
      expect(block.transactions).toHaveLength(expected.transactions.length);
    });

    it("should return an error when the block is not found", async () => {
      // Given
      const connectionMock = {
        rpcEndpoint: "http://solanafake.com",
        getBlock: (slot: number) => Promise.resolve(null),
        getUrl: () => "https://api.mainnet-beta.solana.com",
        setProviderOffline: () => new Date(),
      };
      const poolMock = {
        get: () => connectionMock,
      };
      const repository = new Web3SolanaSlotRepository(poolMock as any);

      try {
        // When
        await repository.getBlock(100);
      } catch (e) {
        // Then
        expect(e).toBeDefined();
      }
    });
  });

  describe("getSignaturesForAddress", () => {
    it("should return confirmed signature info for a given address", async () => {
      // Given
      const expected = [
        {
          signature: "signature1",
          slot: 100,
        },
        {
          signature: "signature2",
          slot: 200,
        },
      ];
      const connectionMock = {
        rpcEndpoint: "http://solanafake.com",
        getSignaturesForAddress: () => Promise.resolve(expected),
      };
      const poolMock = {
        get: () => connectionMock,
      };
      const repository = new Web3SolanaSlotRepository(poolMock as any);

      // When
      const signatures = await repository.getSignaturesForAddress(
        "BTcueXFisZiqE49Ne2xTZjHV9bT5paVZhpKc1k4L3n1c",
        "before",
        "after",
        10
      );

      // Then
      expect(signatures).toBe(expected);
    });
  });

  describe("getTransactions", () => {
    it("should return transactions for a given array of confirmed signature info", async () => {
      // Given
      const expected = [
        {
          signature: "signature1",
          slot: 100,
          chain: "solana",
          chainId: 1,
          transaction: {
            message: {
              version: "legacy",
              accountKeys: [],
              instructions: [],
              compiledInstructions: [],
            },
          },
        },
      ];
      const connectionMock = {
        rpcEndpoint: "http://solanafake.com",
        getTransactions: (sigs: solana.ConfirmedSignatureInfo[]) => Promise.resolve(expected),
      };
      const poolMock = {
        get: () => connectionMock,
      };
      const repository = new Web3SolanaSlotRepository(poolMock as any);

      // When
      const transactions = await repository.getTransactions([
        {
          signature: "signature1",
        },
      ]);

      // Then
      expect(transactions).toStrictEqual(expected);
    });
  });
});
