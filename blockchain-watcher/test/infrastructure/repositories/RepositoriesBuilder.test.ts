import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { RateLimitedWormchainJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/wormchain/RateLimitedWormchainJsonRPCBlockRepository";
import { RateLimitedAlgorandJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/algorand/RateLimitedAlgorandJsonRPCBlockRepository";
import { RateLimitedCosmosJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/cosmos/RateLimitedCosmosJsonRPCBlockRepository";
import { RateLimitedAptosJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/aptos/RateLimitedAptosJsonRPCBlockRepository";
import { RateLimitedEvmJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/evm/RateLimitedEvmJsonRPCBlockRepository";
import { RateLimitedSuiJsonRPCBlockRepository } from "../../../src/infrastructure/repositories/sui/RateLimitedSuiJsonRPCBlockRepository";
import { describe, expect, it } from "@jest/globals";
import { RepositoriesBuilder } from "../../../src/infrastructure/repositories/RepositoriesBuilder";
import { configMock } from "../../mocks/configMock";
import {
  RateLimitedSolanaSlotRepository,
  FileMetadataRepository,
  PromStatRepository,
  SnsEventRepository,
} from "../../../src/infrastructure/repositories";

describe("RepositoriesBuilder", () => {
  it("should throw error because does not have any chain", async () => {
    try {
      // When
      new RepositoriesBuilder(configMock());
    } catch (e: Error | any) {
      // Then
      expect(e).toBeInstanceOf(Error);
    }
  });

  it("should throw error because dose not support test chain", async () => {
    try {
      // When
      new RepositoriesBuilder(configMock());
    } catch (e) {
      // Then
      expect(e).toBeInstanceOf(Error);
    }
  });

  it("should return all repositories instances", async () => {
    // When
    const repos = new RepositoriesBuilder(configMock());
    // Then
    const jobs = repos.getJobs();
    expect(jobs).toBeTruthy();

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
    expect(repos.getEvmBlockRepository("xlayer")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("berachain")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getEvmBlockRepository("unichain")).toBeInstanceOf(
      RateLimitedEvmJsonRPCBlockRepository
    );
    expect(repos.getAlgorandRepository()).toBeInstanceOf(RateLimitedAlgorandJsonRPCBlockRepository);
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
    expect(repos.getCosmosRepository()).toBeInstanceOf(RateLimitedCosmosJsonRPCBlockRepository);
  });
});
