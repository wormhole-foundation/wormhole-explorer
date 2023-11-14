import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { setTimeout } from "timers/promises";
import {
  PollEvmLogsMetadata,
  PollEvmLogs,
  PollEvmLogsConfig,
} from "../../src/domain/actions/PollEvmLogs";
import {
  EvmBlockRepository,
  MetadataRepository,
  StatRepository,
} from "../../src/domain/repositories";
import { EvmBlock, EvmLog } from "../../src/domain/entities";

let cfg = PollEvmLogsConfig.fromBlock(0n);

let getBlocksSpy: jest.SpiedFunction<EvmBlockRepository["getBlocks"]>;
let getLogsSpy: jest.SpiedFunction<EvmBlockRepository["getFilteredLogs"]>;
let handlerSpy: jest.SpiedFunction<(logs: EvmLog[]) => Promise<void>>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollEvmLogsMetadata>["save"]>;

let metadataRepo: MetadataRepository<PollEvmLogsMetadata>;
let evmBlockRepo: EvmBlockRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (logs: EvmLog[]) => Promise.resolve(),
  failing: (logs: EvmLog[]) => Promise.reject(),
};
let pollEvmLogs: PollEvmLogs;

describe("PollEvmLogs", () => {
  afterEach(async () => {
    await pollEvmLogs.stop();
  });

  it("should be able to read logs from latest block when no fromBlock is configured", async () => {
    const currentHeight = 10n;
    const blocksAhead = 1n;
    givenEvmBlockRepository(currentHeight, blocksAhead);
    givenMetadataRepository();
    givenStatsRepository();
    givenPollEvmLogs();

    await whenPollEvmLogsStarts();

    await thenWaitForAssertion(
      () => expect(getBlocksSpy).toHaveReturnedTimes(1),
      () => expect(getBlocksSpy).toHaveBeenCalledWith(new Set([currentHeight, currentHeight + 1n])),
      () =>
        expect(getLogsSpy).toBeCalledWith({
          addresses: cfg.addresses,
          topics: cfg.topics,
          fromBlock: currentHeight + blocksAhead,
          toBlock: currentHeight + blocksAhead,
        })
    );
  });

  it("should be able to read logs from last known block when configured from is before", async () => {
    const lastExtractedBlock = 10n;
    const blocksAhead = 10n;
    givenEvmBlockRepository(lastExtractedBlock, blocksAhead);
    givenMetadataRepository({ lastBlock: lastExtractedBlock });
    givenStatsRepository();
    givenPollEvmLogs(lastExtractedBlock - 10n);

    await whenPollEvmLogsStarts();

    await thenWaitForAssertion(
      () => () =>
        expect(getBlocksSpy).toHaveBeenCalledWith(
          new Set([lastExtractedBlock, lastExtractedBlock + 1n])
        ),
      () =>
        expect(getLogsSpy).toBeCalledWith({
          addresses: cfg.addresses,
          topics: cfg.topics,
          fromBlock: lastExtractedBlock + 1n,
          toBlock: lastExtractedBlock + blocksAhead,
        })
    );
  });

  it("should pass logs to handlers and persist metadata", async () => {
    const currentHeight = 10n;
    const blocksAhead = 1n;
    givenEvmBlockRepository(currentHeight, blocksAhead);
    givenMetadataRepository();
    givenStatsRepository();
    givenPollEvmLogs(currentHeight);

    await whenPollEvmLogsStarts();

    await thenWaitForAssertion(
      () => expect(handlerSpy).toHaveBeenCalledWith(expect.any(Array)),
      () =>
        expect(metadataSaveSpy).toBeCalledWith("watch-evm-logs", {
          lastBlock: currentHeight + blocksAhead,
        })
    );
  });
});

const givenEvmBlockRepository = (height?: bigint, blocksAhead?: bigint) => {
  const logsResponse: EvmLog[] = [];
  const blocksResponse: Record<string, EvmBlock> = {};
  if (height) {
    for (let index = 0n; index <= (blocksAhead ?? 1n); index++) {
      logsResponse.push({
        blockNumber: height + index,
        blockHash: `0x0${index}`,
        blockTime: 0,
        address: "",
        removed: false,
        data: "",
        transactionHash: "",
        transactionIndex: "",
        topics: [],
        logIndex: 0,
      });
      blocksResponse[`0x0${index}`] = {
        timestamp: 0,
        hash: `0x0${index}`,
        number: height + index,
      };
    }
  }

  evmBlockRepo = {
    getBlocks: () => Promise.resolve(blocksResponse),
    getBlockHeight: () => Promise.resolve(height ? height + (blocksAhead ?? 10n) : 10n),
    getFilteredLogs: () => Promise.resolve(logsResponse),
  };

  getBlocksSpy = jest.spyOn(evmBlockRepo, "getBlocks");
  getLogsSpy = jest.spyOn(evmBlockRepo, "getFilteredLogs");
  handlerSpy = jest.spyOn(handlers, "working");
};

const givenMetadataRepository = (data?: PollEvmLogsMetadata) => {
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

const givenPollEvmLogs = (from?: bigint) => {
  cfg.setFromBlock(from);
  pollEvmLogs = new PollEvmLogs(evmBlockRepo, metadataRepo, statsRepo, cfg);
};

const whenPollEvmLogsStarts = async () => {
  await pollEvmLogs.start([handlers.working]);
};

const thenWaitForAssertion = async (...assertions: (() => void)[]) => {
  for (let index = 1; index < 5; index++) {
    try {
      for (const assertion of assertions) {
        assertion();
      }
      break;
    } catch (error) {
      if (index === 4) {
        throw error;
      }
      await setTimeout(10, undefined, { ref: false });
    }
  }
};
