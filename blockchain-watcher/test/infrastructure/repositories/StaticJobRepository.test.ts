import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { beforeEach, describe, expect, it } from "@jest/globals";
import fs from "fs";
import { SnsEventRepository, StaticJobRepository } from "../../../src/infrastructure/repositories";
import {
  AptosRepository,
  EvmBlockRepository,
  MetadataRepository,
  SolanaSlotRepository,
  StatRepository,
  SuiRepository,
  WormchainRepository,
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
    expect(jobs[0].id).toEqual("poll-log-message-published-ethereum");
  });
});

const givenJobsPresent = () => {
  const jobs = [
    {
      id: "poll-log-message-published-ethereum",
      chain: "ethereum",
      source: {
        action: "PollEvm",
        config: {
          fromBlock: 10012499n,
          blockBatchSize: 100,
          commitment: "latest",
          interval: 15_000,
          addresses: ["0x706abc4E45D419950511e474C7B9Ed348A4a716c"],
          chain: "ethereum",
          topics: [],
        },
      },
      handlers: [
        {
          action: "HandleEvmLogs",
          target: "sns",
          mapper: "evmLogMessagePublishedMapper",
          config: {
            abi: "event LogMessagePublished(address indexed sender, uint64 sequence, uint32 nonce, bytes payload, uint8 consistencyLevel)",
            filter: {
              addresses: ["0x706abc4E45D419950511e474C7B9Ed348A4a716c"],
              topics: ["0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2"],
            },
          },
        },
      ],
    },
  ];
  fs.writeFileSync(dirPath + "/jobs.json", JSON.stringify(jobs));
};
