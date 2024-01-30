import { afterEach, describe, expect, it, jest } from "@jest/globals";

import {
  PollSuiCheckpoints,
  PollSuiCheckpointsConfig,
  PollSuiCheckpointsMetadata,
} from "../../../../src/domain/actions/sui/PollSuiCheckpoints";
import {
  MetadataRepository,
  StatRepository,
  SuiRepository,
} from "../../../../src/domain/repositories";
import { mockMetadataRepository, mockStatsRepository } from "../../../mocks/repos";
import { thenWaitForAssertion } from "../../../wait-assertion";
import { GetSuiTransactions } from "../../../../src/domain/actions/sui/GetSuiTransactions";

let statsRepo: StatRepository;
let metadataRepo: MetadataRepository<PollSuiCheckpointsMetadata>;
let suiRepo: SuiRepository;
let poll: PollSuiCheckpoints;

let getLastCheckpointNumberSpy: jest.SpiedFunction<SuiRepository["getLastCheckpointNumber"]>;
let getCheckpointsSpy: jest.SpiedFunction<SuiRepository["getCheckpoints"]>;
let getTransactionBlockReceiptsSpy: jest.SpiedFunction<
  SuiRepository["getTransactionBlockReceipts"]
>;
let actionSpy: jest.SpiedFunction<GetSuiTransactions["execute"]>;

const handler = () => Promise.resolve();

describe("PollSuiCheckpoints", () => {
  afterEach(async () => {
    await poll.stop();
  });

  it("begins polling from the latest block when no range configured and starting from scratch", async () => {
    const lastCheckpoint = 10n;

    givenStatsRepository();
    givenMetadataRepository();
    givenSuiRepository(lastCheckpoint);
    givenPollSui();

    await whenPollingStarts();

    await thenWaitForAssertion(
      () => expect(getLastCheckpointNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getCheckpointsSpy).toHaveBeenCalledWith({
          from: lastCheckpoint,
          to: lastCheckpoint,
        }),
      () => expect(getTransactionBlockReceiptsSpy).toHaveBeenCalledTimes(1)
    );
  });

  it("begins polling from the last known block", async () => {
    const lastProcessed = 9n;
    const latestCheckpoint = 30n;

    givenStatsRepository();
    givenMetadataRepository({ lastCheckpoint: lastProcessed });
    givenSuiRepository(latestCheckpoint);
    givenPollSui();

    await whenPollingStarts();

    await thenWaitForAssertion(
      () => expect(getLastCheckpointNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getCheckpointsSpy).toHaveBeenCalledWith(
          // default batch size 10
          { from: 10n, to: 19n }
        ),
      () => expect(getTransactionBlockReceiptsSpy).toHaveBeenCalledTimes(1)
    );
  });

  it("polls with a batch size", async () => {
    const lastProcessed = 9n;
    const latestCheckpoint = 130n;

    givenStatsRepository();
    givenMetadataRepository({ lastCheckpoint: lastProcessed });
    givenSuiRepository(latestCheckpoint);
    givenPollSui({ batchSize: 50 });

    await whenPollingStarts();

    await thenWaitForAssertion(
      () => expect(getLastCheckpointNumberSpy).toHaveReturnedTimes(1),
      () => expect(getCheckpointsSpy).toHaveBeenCalledWith({ from: 10n, to: 59n }),
      () => expect(getTransactionBlockReceiptsSpy).toHaveBeenCalledTimes(1)
    );
  });

  it("it won't execute the action if it has reached the latest block", async () => {
    const lastProcessed = 30n;
    const latestCheckpoint = 30n;

    givenStatsRepository();
    givenMetadataRepository({ lastCheckpoint: lastProcessed });
    givenSuiRepository(latestCheckpoint);
    givenPollSui();

    await whenPollingStarts();

    await thenWaitForAssertion(
      () => expect(actionSpy).not.toHaveBeenCalled(),
      () => expect(getLastCheckpointNumberSpy).toHaveReturnedTimes(1),
      () => expect(getCheckpointsSpy).not.toHaveBeenCalled(),
      () => expect(getTransactionBlockReceiptsSpy).not.toHaveBeenCalled()
    );
  });

  it("should process blocks from a given range prior to the latest", async () => {
    const from = 20n;
    const to = 30n;
    const latestCheckpoint = 100n;

    givenStatsRepository();
    givenMetadataRepository();
    givenSuiRepository(latestCheckpoint);
    givenPollSui({ from, to });

    await whenPollingStarts();

    await thenWaitForAssertion(
      () => expect(getLastCheckpointNumberSpy).toHaveReturnedTimes(1),
      () => expect(getCheckpointsSpy).toHaveBeenCalledWith({ from: 20n, to: 29n }),
      () => expect(getTransactionBlockReceiptsSpy).toHaveBeenCalledTimes(1)
    );
  });

  it("should cap the range to the latest available", async () => {
    const from = 95n;
    const latestCheckpoint = 100n;

    givenStatsRepository();
    givenMetadataRepository();
    givenSuiRepository(latestCheckpoint);
    givenPollSui({ from });

    await whenPollingStarts();

    await thenWaitForAssertion(
      () => expect(getLastCheckpointNumberSpy).toHaveReturnedTimes(1),
      () => expect(getCheckpointsSpy).toHaveBeenCalledWith({ from: 95n, to: 100n }),
      () => expect(getTransactionBlockReceiptsSpy).toHaveBeenCalledTimes(1)
    );
  });

  it("should skip the cursor if it's prior to the range start", async () => {
    const lastProcessed = 10n;
    const from = 80n;
    const latestCheckpoint = 100n;

    givenStatsRepository();
    givenMetadataRepository({ lastCheckpoint: lastProcessed });
    givenSuiRepository(latestCheckpoint);
    givenPollSui({ from });

    await whenPollingStarts();

    await thenWaitForAssertion(
      () => expect(getLastCheckpointNumberSpy).toHaveReturnedTimes(1),
      () => expect(getCheckpointsSpy).toHaveBeenCalledWith({ from: 80n, to: 89n }),
      () => expect(getTransactionBlockReceiptsSpy).toHaveBeenCalledTimes(1)
    );
  });
});

const givenStatsRepository = () => {
  statsRepo = mockStatsRepository();
};

const givenMetadataRepository = (metadata?: PollSuiCheckpointsMetadata) => {
  metadataRepo = mockMetadataRepository(metadata);
};

const givenSuiRepository = (last?: bigint) => {
  suiRepo = {
    getLastCheckpointNumber: () => Promise.resolve(last || 100n),
    getTransactionBlockReceipts: () => Promise.resolve([]),
    getCheckpoints: () => Promise.resolve([]),

    getCheckpoint: () => Promise.resolve({} as any),
    getLastCheckpoint: () => Promise.resolve({} as any),
    queryTransactions: () => Promise.resolve([]),
  };

  getLastCheckpointNumberSpy = jest.spyOn(suiRepo, "getLastCheckpointNumber");
  getCheckpointsSpy = jest.spyOn(suiRepo, "getCheckpoints");
  getTransactionBlockReceiptsSpy = jest.spyOn(suiRepo, "getTransactionBlockReceipts");
};

const whenPollingStarts = async () => {
  poll.run([handler]);
};

const givenPollSui = (cfg?: Partial<PollSuiCheckpointsConfig>) => {
  const action = new GetSuiTransactions(suiRepo);
  actionSpy = jest.spyOn(action, "execute");
  poll = new PollSuiCheckpoints(
    new PollSuiCheckpointsConfig({ ...cfg, id: "poll-sui-checkpoints" }),
    statsRepo,
    metadataRepo,
    suiRepo,
    action
  );
};
