import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { RateLimitedWormchainJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/wormchain/RateLimitedWormchainJsonRPCBlockRepository";
import { RateLimitedAptosJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/aptos/RateLimitedAptosJsonRPCBlockRepository";
import { RateLimitedEvmJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/evm/RateLimitedEvmJsonRPCBlockRepository";
import { RateLimitedSuiJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/sui/RateLimitedSuiJsonRPCBlockRepository";
import { describe, expect, it } from "@jest/globals";
import { RepositoriesBuilder } from "../../../src/infrastructure/repositories/RepositoriesBuilder";
import { configMock } from "../../mocks/configMock";
import {
  FileMetadataRepository,
  PromStatRepository,
  RateLimitedSolanaSlotRepository,
  SnsEventRepository,
} from "../../../src/infrastructure/repositories";

describe("RepositoriesBuilder", () => {
  it("should be throw error because dose not have any chain", async () => {
    try {
      // When
      new RepositoriesBuilder(configMock());
    } catch (e: Error | any) {
      // Then
      expect(e).toBeInstanceOf(Error);
    }
  });

  it("should be throw error because dose not support test chain", async () => {
    try {
      // When
      new RepositoriesBuilder(configMock());
    } catch (e) {
      // Then
      expect(e).toBeInstanceOf(Error);
    }
  });

  it("should be return all repositories instances", async () => {
    // When
    const repos = new RepositoriesBuilder(configMock());
    // Then
    const job = repos.getJobsRepository();
    expect(job).toBeTruthy();

    expect(repos.getEvmBlockRepository("ethereum")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("ethereum-sepolia")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("bsc")).toBeInstanceOf(RateLimitedEvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("polygon")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("avalanche")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("oasis")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("fantom")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("karura")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("acala")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("klaytn")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("celo")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("arbitrum")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("arbitrum-sepolia")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("moonbeam")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("optimism")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("optimism-sepolia")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("base")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("base-sepolia")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("ethereum-holesky")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("scroll")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("polygon-sepolia")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("blast")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("mantle")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getAptosRepository()).toBeInstanceOf(RateLimitedAptosJsonRPCBlockRepository);
    expect(repos.getMetadataRepository()).toBeInstanceOf(FileMetadataRepository);
    expect(repos.getSnsEventRepository()).toBeInstanceOf(SnsEventRepository);
    expect(repos.getStatsRepository()).toBeInstanceOf(PromStatRepository);
    expect(repos.getSolanaSlotRepository()).toBeInstanceOf(RateLimitedSolanaSlotRepository);
    expect(repos.getSuiRepository()).toBeInstanceOf(RateLimitedSuiJsonRPCBlockRepository);
    expect(repos.getAptosRepository()).toBeInstanceOf(RateLimitedAptosJsonRPCBlockRepository);
    expect(repos.getWormchainRepository()).toBeInstanceOf(
      RateLimitedWormchainJsonRPCBlockRepository
    );
  });
});
