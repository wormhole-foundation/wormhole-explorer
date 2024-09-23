import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { WormchainTransaction } from "../../../../src/domain/entities/wormchain";
import { thenWaitForAssertion } from "../../../waitAssertion";
import {
  PollWormchainLogsMetadata,
  PollWormchainLogsConfig,
  PollWormchain,
} from "../../../../src/domain/actions/wormchain/PollWormchain";
import {
  WormchainRepository,
  MetadataRepository,
  StatRepository,
} from "../../../../src/domain/repositories";

let getBlockHeightSpy: jest.SpiedFunction<WormchainRepository["getBlockHeight"]>;
let getBlockLogsSpy: jest.SpiedFunction<WormchainRepository["getBlockLogs"]>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollWormchainLogsMetadata>["save"]>;

let handlerSpy: jest.SpiedFunction<(txs: WormchainTransaction[]) => Promise<void>>;

let metadataRepo: MetadataRepository<PollWormchainLogsMetadata>;
let wormchainRepo: WormchainRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (txs: WormchainTransaction[]) => Promise.resolve(),
  failing: (txs: WormchainTransaction[]) => Promise.reject(),
};
let pollWormchain: PollWormchain;

let props = {
  blockBatchSize: 100,
  from: 0n,
  limit: 0n,
  environment: "testnet",
  commitment: "immediate",
  addresses: ["wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j"],
  interval: 5000,
  topics: [],
  chainId: 3104,
  filter: {
    address: "wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j",
  },
  chain: "wormchain",
  id: "poll-log-message-published-wormchain",
};

let cfg = new PollWormchainLogsConfig(props);

describe("GetWormchainLogs", () => {
  afterEach(async () => {
    await pollWormchain.stop();
  });

  it("should be skip the transations blocks, because the transactions will be undefined", async () => {
    // Given
    givenWormchainBlockRepository(7606614n);
    givenMetadataRepository({ lastBlock: 7606613n });
    givenStatsRepository();
    givenPollWormchainTx(cfg);

    // When
    await whenPollWormchainLogsStarts();

    // Then
    await thenWaitForAssertion(() =>
      expect(getBlockLogsSpy).toBeCalledWith("wormchain", 7606614n, ["wasm"])
    );
  });

  it("should be process the log because it contains wasm transactions", async () => {
    // Given
    const log = {
      transactions: [
        {
          hash: "0xd84a9c85170c28b12a1436082e99c1ea2598cbf36f9e263bfc0b7fb79a972dfe",
          type: "wasm",
          attributes: [
            {
              key: "X2NvbnRyYWN0X2FkZHJlc3M=",
              value:
                "d29ybWhvbGUxNGhqMnRhdnE4ZnBlc2R3eHhjdTQ0cnR5M2hoOTB2aHVqcnZjbXN0bDR6cjN0eG1mdnc5c3JyZzQ2NQ==",
              index: true,
            },
            { key: "YWN0aW9u", value: "c3VibWl0X29ic2VydmF0aW9ucw==", index: true },
            {
              key: "b3duZXI=",
              value: "d29ybWhvbGUxcWdqZmRmNWczMnN4d2pqanduNHZwOGZnZzJmdzBtOHh4aDdheGM=",
              index: true,
            },
          ],
        },
        {
          hash: "0x9042d7f656f2292e8a4bfa9468ee8215fd6de9ff23b447e20f96f6a70559df68",
          type: "wasm",
          attributes: [
            {
              key: "X2NvbnRyYWN0X2FkZHJlc3M=",
              value:
                "d29ybWhvbGUxNGhqMnRhdnE4ZnBlc2R3eHhjdTQ0cnR5M2hoOTB2aHVqcnZjbXN0bDR6cjN0eG1mdnc5c3JyZzQ2NQ==",
              index: true,
            },
            { key: "YWN0aW9u", value: "c3VibWl0X29ic2VydmF0aW9ucw==", index: true },
            {
              key: "b3duZXI=",
              value: "d29ybWhvbGUxNW5rbTdhdnB4eHNuY3I0Z2c4dTJxbDdnY2tsbTJrcmt6d2U3N20=",
              index: true,
            },
          ],
        },
      ],
      blockHeight: "7606615",
      timestamp: 1711025902418,
    };

    givenWormchainBlockRepository(7606615n, log);
    givenMetadataRepository({ lastBlock: 7606614n });
    givenStatsRepository();
    givenPollWormchainTx(cfg);

    // When
    await whenPollWormchainLogsStarts();

    // Then
    await thenWaitForAssertion(() =>
      expect(getBlockLogsSpy).toBeCalledWith("wormchain", 7606615n, ["wasm"])
    );
  });
});

const givenWormchainBlockRepository = (blockHeigh: bigint, log: any = {}) => {
  wormchainRepo = {
    getBlockHeight: () => Promise.resolve(blockHeigh),
    getBlockLogs: () => Promise.resolve(log),
    getRedeems: () => Promise.resolve([]),
    healthCheck: () => Promise.resolve(),
  };

  getBlockHeightSpy = jest.spyOn(wormchainRepo, "getBlockHeight");
  getBlockLogsSpy = jest.spyOn(wormchainRepo, "getBlockLogs");
};

const givenMetadataRepository = (data?: PollWormchainLogsMetadata) => {
  metadataRepo = {
    get: () => Promise.resolve(data),
    save: () => Promise.resolve(),
  };
  metadataSaveSpy = jest.spyOn(metadataRepo, "save");
};

const givenStatsRepository = () => {
  statsRepo = {
    count: () => {},
    measure: () => {},
    report: () => Promise.resolve(""),
  };
};

const givenPollWormchainTx = (cfg: PollWormchainLogsConfig) => {
  pollWormchain = new PollWormchain(
    wormchainRepo,
    metadataRepo,
    statsRepo,
    cfg,
    "GetWormchainLogs"
  );
};

const whenPollWormchainLogsStarts = async () => {
  pollWormchain.run([handlers.working]);
};
