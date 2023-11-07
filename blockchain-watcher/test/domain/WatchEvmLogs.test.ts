import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { setTimeout } from "timers/promises";
import {
  WatchEvmLogsMetadata,
  WatchEvmLogs,
  WatchEvmLogsConfig,
} from "../../src/domain/actions/WatchEvmLogs";
import {
  EvmBlockRepository,
  MetadataRepository,
} from "../../src/domain/repositories";
import { EvmBlock, EvmLog } from "../../src/domain/entities";

let cfg = WatchEvmLogsConfig.fromBlock(0n);

let getBlocksSpy: jest.SpiedFunction<EvmBlockRepository["getBlocks"]>;
let getLogsSpy: jest.SpiedFunction<EvmBlockRepository["getFilteredLogs"]>;

let metadataRepo: MetadataRepository<WatchEvmLogsMetadata>;
let evmBlockRepo: EvmBlockRepository;
let watchEvmLogs: WatchEvmLogs;

describe("WatchEvmLogs", () => {
  afterEach(async () => {
    await watchEvmLogs.stop();
  });

  it("should be able to read logs from given start", async () => {
    const currentHeight = 10n;
    const blocksAhead = 1n;
    givenEvmBlockRepository(currentHeight, blocksAhead);
    givenMetadataRepository();
    givenWatchEvmLogs(currentHeight);

    await whenWatchEvmLogsStarts();

    await thenWaitForAssertion(
      () =>
        expect(getBlocksSpy).toHaveBeenCalledWith(
          new Set([currentHeight, currentHeight + 1n])
        ),
      () =>
        expect(getLogsSpy).toBeCalledWith({
          addresses: cfg.addresses,
          topics: cfg.topics,
          fromBlock: currentHeight,
          toBlock: currentHeight + blocksAhead,
        })
    );
  });

  it("should be able to read logs from last known block when configured from is before", async () => {
    const lastExtractedBlock = 10n;
    const blocksAhead = 10n;
    givenEvmBlockRepository(lastExtractedBlock, blocksAhead);
    givenMetadataRepository({ lastBlock: lastExtractedBlock });
    givenWatchEvmLogs(lastExtractedBlock - 10n);

    await whenWatchEvmLogsStarts();

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
});

const givenEvmBlockRepository = (height?: bigint, blocksAhead?: bigint) => {
  const logsResponse: EvmLog[] = [];
  if (height) {
    for (let index = 0n; index <= (blocksAhead ?? 1n); index++) {
      logsResponse.push({
        blockNumber: height + index,
        blockHash: "",
        address: "",
        removed: false,
        data: "",
        transactionHash: "",
        transactionIndex: "",
        topics: [],
        logIndex: 0,
      });
    }
  }

  evmBlockRepo = {
    getBlocks: () => Promise.resolve([]),
    getBlockHeight: () =>
      Promise.resolve(height ? height + (blocksAhead ?? 10n) : 10n),
    getFilteredLogs: () => Promise.resolve(logsResponse),
  };

  getBlocksSpy = jest.spyOn(evmBlockRepo, "getBlocks");
  getLogsSpy = jest.spyOn(evmBlockRepo, "getFilteredLogs");
};

const givenMetadataRepository = (data?: WatchEvmLogsMetadata) => {
  metadataRepo = {
    getMetadata: () => Promise.resolve(data),
  };
};

const givenWatchEvmLogs = (from?: bigint) => {
  cfg.fromBlock = from ?? cfg.fromBlock;
  watchEvmLogs = new WatchEvmLogs(evmBlockRepo, metadataRepo, cfg);
};

const whenWatchEvmLogsStarts = async () => {
  await watchEvmLogs.start();
};

const thenWaitForAssertion = async (...assertions: (() => void)[]) => {
  for (let index = 1; index < 5; index++) {
    try {
      for (const assertion of assertions) {
        assertion();
      }
      break;
    } catch (error) {
      await setTimeout(10, undefined, { ref: false });
      if (index === 4) {
        throw error;
      }
    }
  }
};
