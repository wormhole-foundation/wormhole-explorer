import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { setTimeout } from "timers/promises";
import {
  PollEvmLogsMetadata,
  PollEvmLogs,
  PollEvmLogsConfig,
} from "../../src/domain/actions/PollEvmLogs";
import { EvmBlockRepository, MetadataRepository } from "../../src/domain/repositories";
import { EvmBlock, EvmLog } from "../../src/domain/entities";

let cfg = PollEvmLogsConfig.fromBlock(0n);

let getBlocksSpy: jest.SpiedFunction<EvmBlockRepository["getBlocks"]>;
let getLogsSpy: jest.SpiedFunction<EvmBlockRepository["getFilteredLogs"]>;

let metadataRepo: MetadataRepository<PollEvmLogsMetadata>;
let evmBlockRepo: EvmBlockRepository;
let pollEvmLogs: PollEvmLogs;

describe("PollEvmLogs", () => {
  afterEach(async () => {
    await pollEvmLogs.stop();
  });

  it("should be able to read logs from given start", async () => {
    const currentHeight = 10n;
    const blocksAhead = 1n;
    givenEvmBlockRepository(currentHeight, blocksAhead);
    givenMetadataRepository();
    givenPollEvmLogs(currentHeight);

    await whenPollEvmLogsStarts();

    await thenWaitForAssertion(
      () => expect(getBlocksSpy).toHaveBeenCalledWith(new Set([currentHeight, currentHeight + 1n])),
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
    getBlockHeight: () => Promise.resolve(height ? height + (blocksAhead ?? 10n) : 10n),
    getFilteredLogs: () => Promise.resolve(logsResponse),
  };

  getBlocksSpy = jest.spyOn(evmBlockRepo, "getBlocks");
  getLogsSpy = jest.spyOn(evmBlockRepo, "getFilteredLogs");
};

const givenMetadataRepository = (data?: PollEvmLogsMetadata) => {
  metadataRepo = {
    getMetadata: () => Promise.resolve(data),
  };
};

const givenPollEvmLogs = (from?: bigint) => {
  cfg.fromBlock = from ?? cfg.fromBlock;
  pollEvmLogs = new PollEvmLogs(evmBlockRepo, metadataRepo, cfg);
};

const whenPollEvmLogsStarts = async () => {
  await pollEvmLogs.start();
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
