import { describe, expect, it } from "@jest/globals";
import { RepositoriesBuilder } from "../../../src/infrastructure/repositories/RepositoriesBuilder";
import { configMock } from "../../mocks/configMock";
import {
  BscEvmJsonRPCBlockRepository,
  EvmJsonRPCBlockRepository,
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
      expect(false).toBe(true);
    } catch (e: Error | any) {
      // Then
      expect(e).toBeInstanceOf(Error);
    }
  });

  it("should be return all repositories instances", async () => {
    // When
    const repos = new RepositoriesBuilder(configMock());
    await repos.init();
    // Then
    const job = repos.getJobsRepository();
    expect(job).toBeTruthy();

    expect(repos.getEvmBlockRepository("ethereum")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("bsc")).toBeInstanceOf(BscEvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("avalanche")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("oasis")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("fantom")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("karura")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("acala")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("klaytn")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("celo")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("optimism")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getEvmBlockRepository("base")).toBeInstanceOf(EvmJsonRPCBlockRepository);
    expect(repos.getMetadataRepository()).toBeInstanceOf(FileMetadataRepository);
    expect(repos.getSnsEventRepository()).toBeInstanceOf(SnsEventRepository);
    expect(repos.getStatsRepository()).toBeInstanceOf(PromStatRepository);
    expect(repos.getSolanaSlotRepository()).toBeInstanceOf(RateLimitedSolanaSlotRepository);
  });
});
