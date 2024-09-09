import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { expect, describe, it } from "@jest/globals";
import {
  Web3SolanaSlotRepository,
  RateLimitedSolanaSlotRepository,
} from "../../../src/infrastructure/repositories";

const repoMock = {
  getSlot: () => Promise.resolve(100),
  getLatestSlot: () => Promise.resolve(100),
  getBlock: () => Promise.resolve({ blockTime: 100, transactions: [] }),
  getSignaturesForAddress: () => Promise.resolve([]),
  getTransactions: () => Promise.resolve([]),
} as any as Web3SolanaSlotRepository;

describe("RateLimitedSolanaSlotRepository", () => {
  describe("getLatestSlot", () => {
    it("should fail when ratelimit is exceeded", async () => {
      const repository = new RateLimitedSolanaSlotRepository(repoMock, "solana", {
        period: 1000,
        limit: 1,
        interval: 1_000,
        attempts: 10,
      });

      await repository.getLatestSlot("confirmed");
      await expect(repository.getLatestSlot("confirmed")).rejects.toThrowError();
    });
  });

  describe("getBlock", () => {
    it("should fail when ratelimit is exceeded", async () => {
      const repository = new RateLimitedSolanaSlotRepository(repoMock, "solana", {
        period: 1000,
        limit: 1,
        interval: 1_000,
        attempts: 10,
      });

      await repository.getBlock(1);
      const failure = await repository.getBlock(1);

      expect(failure.getError()).toHaveProperty("message", "Ratelimited");
    });
  });

  describe("getSignaturesForAddress", () => {
    it("should fail when ratelimit is exceeded", async () => {
      const repository = new RateLimitedSolanaSlotRepository(repoMock, "solana", {
        period: 1000,
        limit: 1,
        interval: 1_000,
        attempts: 10,
      });

      await repository.getSignaturesForAddress("address", "before", "after", 1);
      await expect(
        repository.getSignaturesForAddress("address", "before", "after", 1)
      ).rejects.toThrowError();
    });
  });

  describe("getTransactions", () => {
    it("should fail when ratelimit is exceeded", async () => {
      const repository = new RateLimitedSolanaSlotRepository(repoMock, "solana", {
        period: 1000,
        limit: 1,
        interval: 1_000,
        attempts: 10,
      });

      await repository.getTransactions([]);
      await expect(repository.getTransactions([])).rejects.toThrowError();
    });
  });
});
