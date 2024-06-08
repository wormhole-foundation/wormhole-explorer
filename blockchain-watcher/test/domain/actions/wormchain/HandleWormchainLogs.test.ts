import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { WormchainBlockLogs } from "../../../../src/domain/entities/wormchain";
import { StatRepository } from "../../../../src/domain/repositories";
import { LogFoundEvent } from "../../../../src/domain/entities";
import {
  HandleWormchainLogsOptions,
  HandleWormchainTransactions,
} from "../../../../src/domain/actions/wormchain/HandleWormchainTransactions";

let targetRepoSpy: jest.SpiedFunction<(typeof targetRepo)["save"]>;
let statsRepo: StatRepository;

let HandleWormchainTransactions: HandleWormchainTransactions;
let logs: WormchainBlockLogs[];
let cfg: HandleWormchainLogsOptions;

describe("HandleWormchainTransactions", () => {
  afterEach(async () => {});

  it("should be able to map source events log", async () => {
    // Given
    givenConfig();
    givenStatsRepository();
    givenHandleEvmLogs();

    // When
    const result = await HandleWormchainTransactions.handle(logs);

    // Then
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe("log-message-published");
    expect(result[0].chainId).toBe(3104);
    expect(result[0].txHash).toBe(
      "0x7f61bf387fdb700d32d2b40ccecfb70ae46a2f82775242d04202bb7a538667c6"
    );
    expect(result[0].address).toBe(
      "wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j"
    );
  });
});

const mapper = (addresses: string[], tx: WormchainBlockLogs) => {
  return [
    {
      name: "log-message-published",
      address: "wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j",
      chainId: 3104,
      txHash: "0x7f61bf387fdb700d32d2b40ccecfb70ae46a2f82775242d04202bb7a538667c6",
      blockHeight: 153549311n,
      blockTime: 1709645685704036,
      attributes: {
        sender: "wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j",
        sequence: 203,
        payload: "",
        nonce: 75952,
        consistencyLevel: 0,
        protocol: "Token Bridge",
      },
    },
  ];
};

const targetRepo = {
  save: async (events: LogFoundEvent<Record<string, string>>[]) => {
    Promise.resolve();
  },
  failingSave: async (events: LogFoundEvent<Record<string, string>>[]) => {
    Promise.reject();
  },
};

const givenHandleEvmLogs = (targetFn: "save" | "failingSave" = "save") => {
  targetRepoSpy = jest.spyOn(targetRepo, targetFn);
  HandleWormchainTransactions = new HandleWormchainTransactions(
    cfg,
    mapper,
    () => Promise.resolve(),
    statsRepo
  );
};

const givenConfig = () => {
  cfg = {
    id: "poll-log-message-published-wormchain",
    metricName: "process_source_event",
    filter: { addresses: ["wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j"] },
  };
};

const givenStatsRepository = () => {
  statsRepo = {
    count: () => {},
    measure: () => {},
    report: () => Promise.resolve(""),
  };
};

logs = [
  {
    transactions: [
      {
        hash: "0x7f61bf387fdb700d32d2b40ccecfb70ae46a2f82775242d04202bb7a538667c6",
        height: "7626736",
        tx: Buffer.from("7dm9am6Qx7cH64RB99Mzf7ZsLbEfmXM7ihXXCvMiT2X1", "hex"),
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
    ],
    blockHeight: BigInt(7606614),
    timestamp: 1711025896481,
  },
];
