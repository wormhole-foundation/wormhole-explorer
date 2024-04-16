import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { thenWaitForAssertion } from "../../../wait-assertion";
import { WormchainBlockLogs } from "../../../../src/domain/entities/wormchain";
import {
  PollWormchainLogsMetadata,
  PollWormchainLogsConfig,
  PollWormchain,
} from "../../../../src/domain/actions";
import {
  WormchainRepository,
  MetadataRepository,
  StatRepository,
} from "../../../../src/domain/repositories";

let cfg = new PollWormchainLogsConfig({
  chain: "wormchain",
  fromBlock: 7626734n,
  addresses: [],
  chainId: 3104,
});

let getBlockHeightSpy: jest.SpiedFunction<WormchainRepository["getBlockHeight"]>;
let getBlockLogsSpy: jest.SpiedFunction<WormchainRepository["getBlockLogs"]>;
let handlerSpy: jest.SpiedFunction<(logs: WormchainBlockLogs[]) => Promise<void>>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollWormchainLogsMetadata>["save"]>;

let metadataRepo: MetadataRepository<PollWormchainLogsMetadata>;
let wormchainBlockRepo: WormchainRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (logs: WormchainBlockLogs[]) => Promise.resolve(),
  failing: (logs: WormchainBlockLogs[]) => Promise.reject(),
};
let pollWormchain: PollWormchain;

describe("PollWormchain", () => {
  afterEach(async () => {
    await pollWormchain.stop();
  });

  it("should be able to read logs from latest block when no fromBlock is configured", async () => {
    const currentHeight = 10n;

    const logs = {
      transactions: [
        {
          hash: "0x47a54890a16ea9d924c32a1fa6fd1cf39176be532c8ba454d33f628d89be3388",
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
              value: "d29ybWhvbGUxOHl3NmY4OHA3Znc2bTk5eDlrbnJmejNwMHk2OTNoaDBhaDh5Mm0=",
              index: true,
            },
          ],
        },
        {
          hash: "0x56e974e33c5c7403d23a5fe7fa414d9f1d6dd4f1b67601342100093c604b5d70",
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
              value: "d29ybWhvbGUxYWNxYTV2bDJudW5oc250djBldGpzNnllZTN2NnZjOGw5bTRxOGU=",
              index: true,
            },
          ],
        },
      ],
      blockHeight: "7626735",
      timestamp: 1711143216257,
    };
    givenEvmBlockRepository(currentHeight, logs);
    givenMetadataRepository();
    givenStatsRepository();
    givenPollWormchainLogs();

    await whenPollWormchainLogsStarts();

    await thenWaitForAssertion(
      () => expect(getBlockHeightSpy).toHaveReturnedTimes(1),
      () => expect(getBlockLogsSpy).toHaveBeenCalledWith(3104, currentHeight)
    );
  });

  it("should be return an empty array because to block is more greater than from block", async () => {
    const currentHeight = 10n;

    givenEvmBlockRepository(currentHeight);
    givenMetadataRepository({ lastBlock: 15n });
    givenStatsRepository();
    givenPollWormchainLogs();

    await whenPollWormchainLogsStarts();

    await thenWaitForAssertion(() => expect(getBlockHeightSpy).toHaveReturnedTimes(1));
  });
});

const givenEvmBlockRepository = (height?: bigint, logs: any = []) => {
  wormchainBlockRepo = {
    getBlockHeight: () => Promise.resolve(height),
    getBlockLogs: () => Promise.resolve(logs),
  };

  getBlockHeightSpy = jest.spyOn(wormchainBlockRepo, "getBlockHeight");
  getBlockLogsSpy = jest.spyOn(wormchainBlockRepo, "getBlockLogs");
  handlerSpy = jest.spyOn(handlers, "working");
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

const givenPollWormchainLogs = (from?: bigint) => {
  cfg.setFromBlock(from);
  pollWormchain = new PollWormchain(
    wormchainBlockRepo,
    metadataRepo,
    statsRepo,
    cfg,
    "GetWormchainLogs"
  );
};

const whenPollWormchainLogsStarts = async () => {
  pollWormchain.run([handlers.working]);
};
