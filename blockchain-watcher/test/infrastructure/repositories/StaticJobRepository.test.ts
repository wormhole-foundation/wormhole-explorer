import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { SnsEventRepository, StaticJobRepository } from "../../../src/infrastructure/repositories";
import { beforeEach, describe, expect, it } from "@jest/globals";
import fs from "fs";
import {
  SolanaSlotRepository,
  WormchainRepository,
  EvmBlockRepository,
  MetadataRepository,
  AlgorandRepository,
  CosmosRepository,
  AptosRepository,
  StatRepository,
  SuiRepository,
  NearRepository,
} from "../../../src/domain/repositories";

const dirPath = "./metadata-repo/jobs";
const blockRepo: EvmBlockRepository = {} as any as EvmBlockRepository;
const metadataRepo = {} as MetadataRepository<any>;
const statsRepo = {} as any as StatRepository;
const snsRepo = {} as any as SnsEventRepository;
const solanaSlotRepo = {} as any as SolanaSlotRepository;
const suiRepo = {} as any as SuiRepository;
const aptosRepo = {} as any as AptosRepository;
const wormchainRepo = {} as any as WormchainRepository;
const cosmosRepo = {} as any as CosmosRepository;
const algorandRepo = {} as any as AlgorandRepository;
const nearRepo = {} as any as NearRepository;

let repo: StaticJobRepository;

describe("StaticJobRepository", () => {
  beforeEach(() => {
    if (fs.existsSync(dirPath)) {
      fs.rmSync(dirPath, { recursive: true, force: true });
    }
    repo = new StaticJobRepository("testnet", dirPath, false, () => blockRepo, {
      metadataRepo,
      statsRepo,
      snsRepo,
      solanaSlotRepo,
      suiRepo,
      aptosRepo,
      wormchainRepo,
      cosmosRepo,
      algorandRepo,
      nearRepo,
    });
  });

  it("should return empty when no file available", async () => {
    const jobs = await repo.getJobDefinitions();
    expect(jobs).toHaveLength(0);
  });

  it("should read jobs from file", async () => {
    givenJobsPresent();
    const jobs = await repo.getJobDefinitions();
    expect(jobs).toHaveLength(1);
    expect(jobs[0].id).toEqual("poll-redeemed-transactions-ethereum");
  });
});

const givenJobsPresent = () => {
  const jobs = [
    {
      id: "poll-redeemed-transactions-ethereum",
      chain: "ethereum",
      source: {
        action: "PollEvm",
        records: "GetEvmTransactions",
        config: {
          blockBatchSize: 100,
          commitment: "latest",
          interval: 15000,
          filters: [
            {
              addresses: [],
              type: "Portal Token Bridge (Connect, Portico, Omniswap, tBTC, etc)",
              topics: ["0xcaf280c8cfeba144da67230d9b009c8f868a75bac9a528fa0474be1ba317c169"],
              strategy: "GetTransactionsByLogFiltersStrategy",
            },
            {
              addresses: [],
              type: "CCTP",
              topics: ["0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e"],
              strategy: "GetTransactionsByLogFiltersStrategy",
            },
            {
              addresses: [],
              type: "Standard Relayer",
              topics: ["0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e"],
              strategy: "GetTransactionsByLogFiltersStrategy",
            },
            {
              addresses: [],
              type: "NTT",
              topics: ["0xf6fc529540981400dc64edf649eb5e2e0eb5812a27f8c81bac2c1d317e71a5f0"],
              strategy: "GetTransactionsByLogFiltersStrategy",
            },
            {
              addresses: ["0x6FFd7EdE62328b3Af38FCD61461Bbfc52F5651fE"],
              type: "NFT",
              topics: ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"],
              strategy: "GetTransactionsByBlocksStrategy",
            },
          ],
          chain: "ethereum",
          chainId: 2,
        },
      },
      handlers: [
        {
          action: "HandleEvmTransactions",
          target: "sns",
          mapper: "evmRedeemedTransactionFoundMapper",
          config: {
            abis: [
              {
                abi: "event LogMessagePublished(address indexed sender, uint64 sequence, uint32 nonce, bytes payload, uint8 consistencyLevel)",
                topic: "0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2",
              },
            ],
            metricName: "process_vaa_event",
          },
        },
      ],
    },
  ];
  fs.writeFileSync(dirPath + "/jobs.json", JSON.stringify(jobs));
};
