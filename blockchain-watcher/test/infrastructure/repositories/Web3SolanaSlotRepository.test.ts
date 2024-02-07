import { expect, describe, it } from "@jest/globals";
import { PublicKey } from "@solana/web3.js";
import { solana } from "../../../src/domain/entities";
import { Web3SolanaSlotRepository } from "../../../src/infrastructure/repositories";

describe("Web3SolanaSlotRepository", () => {
  describe("getLatestSlot", () => {
    it("should return the latest slot number", async () => {
      const connectionMock = {
        rpcEndpoint: "http://solanafake.com",
        getSlot: () => Promise.resolve(100),
      };
      const poolMock = {
        get: () => connectionMock,
      };
      const repository = new Web3SolanaSlotRepository(poolMock as any);

      const latestSlot = await repository.getLatestSlot("finalized");

      expect(latestSlot).toBe(100);
    });
  });

  describe("getBlock", () => {
    it("should return a block for a given slot number", async () => {
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

      const block = (await repository.getBlock(100)).getValue();

      expect(block.blockTime).toBe(expected.blockTime);
      expect(block.transactions).toHaveLength(expected.transactions.length);
    });

    it("should return an error when the block is not found", async () => {
      const connectionMock = {
        rpcEndpoint: "http://solanafake.com",
        getBlock: (slot: number) => Promise.resolve(null),
      };
      const poolMock = {
        get: () => connectionMock,
      };
      const repository = new Web3SolanaSlotRepository(poolMock as any);

      const block = await repository.getBlock(100);

      expect(block.getError()).toBeDefined();
    });
  });

  describe("getSignaturesForAddress", () => {
    it("should return confirmed signature info for a given address", async () => {
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

      const signatures = await repository.getSignaturesForAddress(
        "BTcueXFisZiqE49Ne2xTZjHV9bT5paVZhpKc1k4L3n1c",
        "before",
        "after",
        10
      );

      expect(signatures).toBe(expected);
    });
  });

  describe("getTransactions", () => {
    it("should return transactions for a given array of confirmed signature info", async () => {
      const expected = [
        {
          signature: "signature1",
          slot: 100,
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

      const transactions = await repository.getTransactions([
        {
          signature: "signature1",
        },
      ]);

      expect(transactions).toStrictEqual(expected);
    });
  });
});
