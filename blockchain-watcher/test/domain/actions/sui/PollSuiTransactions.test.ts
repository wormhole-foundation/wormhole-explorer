import { afterEach, describe, expect, it, jest } from "@jest/globals";
import { Checkpoint } from "@mysten/sui.js/client";
import base58 from "bs58";

import { randomBytes, randomInt } from "crypto";
import {
  PollSuiTransactions,
  PollSuiTransactionsConfig,
  PollSuiTransactionsMetadata,
} from "../../../../src/domain/actions/sui/PollSuiTransactions";
import {
  MetadataRepository,
  StatRepository,
  SuiRepository,
} from "../../../../src/domain/repositories";
import { mockMetadataRepository, mockStatsRepository } from "../../../mocks/repos";
import { thenWaitForAssertion } from "../../../waitAssertion";
import { SuiTransactionBlockReceipt } from "../../../../src/domain/entities/sui";

let statsRepo: StatRepository;
let metadataRepo: MetadataRepository<PollSuiTransactionsMetadata>;
let suiRepo: SuiRepository;
let poll: PollSuiTransactions;

let queryTransactionsSpy: jest.SpiedFunction<SuiRepository["queryTransactions"]>;

let checkpoints: Checkpoint[];
let lastCheckpoint: Checkpoint;

const filters = [
  {
    MoveFunction: {
      package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
      module: "complete_transfer",
      function: "authorize_transfer",
    },
  },
];

let handler: jest.MockedFunction<() => Promise<any[]>>;

describe("PollSuiTransactions", () => {
  afterEach(async () => {
    await poll.stop();
  });

  it("begins polling from the last tx of the previous to latest block when no range configured and starting from scratch", async () => {
    givenCheckpointRange();
    givenStatsRepository();
    givenMetadataRepository();
    givenSuiRepository();
    givenPollSui();

    await whenPollingStarts();

    const previousToLast = checkpoints[checkpoints.length - 2];
    await thenWaitForAssertion(
      () => expect(queryTransactionsSpy).toHaveReturnedTimes(1),
      () =>
        expect(queryTransactionsSpy).toHaveBeenCalledWith(
          filters[0],
          previousToLast.transactions[previousToLast.transactions.length - 1]
        )
    );
  });

  it("begins polling from the last processed tx", async () => {
    givenCheckpointRange();
    givenStatsRepository();
    givenMetadataRepository({
      lastCursor: { checkpoint: 24937268n, digest: "FpebiL11dZtsf1Lmk9M2Lz4qkGqNMsk5xkKYsSfKz6ab" },
    });
    givenSuiRepository();
    givenPollSui();

    await whenPollingStarts();

    await thenWaitForAssertion(
      () => expect(queryTransactionsSpy).toHaveReturnedTimes(1),
      () =>
        expect(queryTransactionsSpy).toHaveBeenCalledWith(
          filters[0],
          "FpebiL11dZtsf1Lmk9M2Lz4qkGqNMsk5xkKYsSfKz6ab"
        )
    );
  });

  it("begins polling from the last tx of the checkpoint previous to the configured range start", async () => {
    const range = { from: 24937266n };

    givenCheckpointRange({ from: range.from - 1n });
    givenStatsRepository();
    givenMetadataRepository();
    givenSuiRepository();
    givenPollSui(range);

    await whenPollingStarts();

    const prev = checkpoints[0];
    const expectedCursor = prev.transactions[prev.transactions.length - 1];

    await thenWaitForAssertion(
      () => expect(queryTransactionsSpy).toHaveReturnedTimes(1),
      () => expect(queryTransactionsSpy).toHaveBeenCalledWith(filters[0], expectedCursor)
    );
  });

  it("should skip the cursor if it's prior to the range start", async () => {
    const range = { from: 24937266n };

    givenCheckpointRange({ from: range.from - 1n });
    givenStatsRepository();
    givenMetadataRepository({ lastCursor: { checkpoint: 24937000n, digest: randomDigest() } });
    givenSuiRepository();
    givenPollSui(range);

    await whenPollingStarts();

    const prev = checkpoints[0];
    const expectedCursor = prev.transactions[prev.transactions.length - 1];

    await thenWaitForAssertion(
      () => expect(queryTransactionsSpy).toHaveReturnedTimes(1),
      () => expect(queryTransactionsSpy).toHaveBeenCalledWith(filters[0], expectedCursor)
    );
  });

  it("should cap the range to the latest available", async () => {
    const range = { from: 24937210n, to: 24937214n };

    givenCheckpointRange({ from: range.from - 1n });
    givenStatsRepository();
    givenMetadataRepository();
    givenSuiRepository();
    givenPollSui(range);

    await whenPollingStarts();

    const expectedCheckpoints = checkpoints.filter(
      (c) => range.from <= BigInt(c.sequenceNumber) && BigInt(c.sequenceNumber) <= range.to
    );
    const expectedTxs = createTransactions(expectedCheckpoints);

    const prevToStart = checkpoints.find((c) => c.sequenceNumber === (range.from - 1n).toString())!;
    const expectedCursor = prevToStart.transactions[prevToStart.transactions.length - 1];

    await thenWaitForAssertion(
      () => expect(queryTransactionsSpy).toHaveReturnedTimes(1),
      () => expect(queryTransactionsSpy).toHaveBeenCalledWith(filters[0], expectedCursor),
      () => expect(handler).toHaveBeenCalledWith(expectedTxs)
    );
  });
});

const createRange = (n: number) => [...new Array(n).keys()];

const givenCheckpointRange = (range: { from?: bigint; last?: bigint } = {}) => {
  const from = range.from || 24937200n;
  const to = range.last || 24937300n;

  const count = Number(to - from + 1n);
  checkpoints = createRange(count).map((i) => createBlock(Number(from) + i));
  lastCheckpoint = checkpoints[checkpoints.length - 1];
};

const createBlock = (seq: number) => ({
  epoch: "292",
  sequenceNumber: seq.toString(),
  digest: randomDigest(),
  networkTotalTransactions: "1073625143",
  previousDigest: "CqRHfm8WTZY3gPDWsBPkcpUTD5Y2auAzwc2UFdnporv5",
  epochRollingGasCostSummary: {
    computationCost: "4869611922684",
    storageCost: "39035676152800",
    storageRebate: "37763566866036",
    nonRefundableStorageFee: "381450170364",
  },
  timestampMs: (1706633452390 + seq * 1000).toString(),
  transactions: createRange(randomInt(3) + 1).map(() => randomDigest()),
  checkpointCommitments: [],
  validatorSignature: "gN+/Ivd7hYc3fIzOS5Xcmzk9c6uUyYm1V8MkDoVUKltmhSeXicuwMNr/OCLan9tZ",
});

const randomDigest = () => base58.encode(randomBytes(32));

const givenStatsRepository = () => {
  statsRepo = mockStatsRepository();
};

const givenMetadataRepository = (metadata?: PollSuiTransactionsMetadata) => {
  metadataRepo = mockMetadataRepository(metadata);
};

const givenSuiRepository = () => {
  suiRepo = {
    getLastCheckpointNumber: () => Promise.resolve(BigInt(lastCheckpoint.sequenceNumber)),
    getCheckpoint: (seq) =>
      Promise.resolve(checkpoints.find((c) => c.sequenceNumber === seq.toString())!),
    getLastCheckpoint: () => Promise.resolve(lastCheckpoint),
    queryTransactions: (_, cursor) =>
      Promise.resolve(createTransactions(checkpoints, cursor).slice(0, 50)),
    queryTransactionsByEvent: (_, cursor) =>
      Promise.resolve(createTransactions(checkpoints, cursor).slice(0, 50)),
    getTransactionBlockReceipts: () => Promise.resolve([]),
    getCheckpoints: () => Promise.resolve([]),
  };

  queryTransactionsSpy = jest.spyOn(suiRepo, "queryTransactions");
};

const whenPollingStarts = async () => {
  handler = jest.fn();
  poll.run([handler]);
};

const givenPollSui = (cfg?: Partial<PollSuiTransactionsConfig>) => {
  poll = new PollSuiTransactions(
    new PollSuiTransactionsConfig({ ...cfg, filters, id: "poll-sui-transactions" }),
    statsRepo,
    metadataRepo,
    suiRepo
  );
};

const createTransactions = (
  checkpoints: Checkpoint[],
  cursor?: string
): SuiTransactionBlockReceipt[] => {
  const cursorCheckpoint = cursor
    ? checkpoints.find((c) => c.transactions.includes(cursor))
    : undefined;
  const cursorCheckpointIndex = cursorCheckpoint ? checkpoints.indexOf(cursorCheckpoint) : -1;

  return checkpoints
    .slice(cursorCheckpointIndex + 1)
    .map((c) =>
      c.transactions.map((digest) => ({
        checkpoint: c.sequenceNumber,
        digest,
        timestampMs: c.timestampMs,
        events: [],
        errors: [],
        transaction: {} as any,
      }))
    )
    .flat();
};
