import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { AptosTransaction } from "../../../../src/domain/entities/aptos";
import { StatRepository } from "../../../../src/domain/repositories";
import { LogFoundEvent } from "../../../../src/domain/entities";
import {
  HandleAlgorandTransactionsOptions,
  HandleAlgorandTransactions,
} from "../../../../src/domain/actions/algorand/HandleAlgorandTransactions";
import { AlgorandTransaction } from "../../../../src/domain/entities/algorand";

let targetRepoSpy: jest.SpiedFunction<(typeof targetRepo)["save"]>;
let statsRepo: StatRepository;

let handleAptosTransactions: HandleAlgorandTransactions;
let txs: AlgorandTransaction[];
let cfg: HandleAlgorandTransactionsOptions;

describe("HandleAlgorandTransactions", () => {
  afterEach(async () => {});

  it("should be able to map source events tx", async () => {
    // Given
    givenConfig();
    givenStatsRepository();
    givenHandleEvmLogs();

    // When
    const result = await handleAptosTransactions.handle(txs);

    // Then
    expect(result).toHaveLength(1);
    expect(result[0].name).toBe("log-message-published");
    expect(result[0].chainId).toBe(8);
    expect(result[0].txHash).toBe("SQA7S37MCLGHQRMFZHRNUNUFJ6PJKRZN5RO52NMEWJU5B365SINQ");
    expect(result[0].address).toBe("MG3DIJNS3JTVKUAQGFV5BQTDAK26OUM3SRXSLIFWVUS67V54VPKDUJQTOQ");
  });
});

const mapper = (tx: AlgorandTransaction) => {
  return {
    name: "log-message-published",
    address: "MG3DIJNS3JTVKUAQGFV5BQTDAK26OUM3SRXSLIFWVUS67V54VPKDUJQTOQ",
    chainId: 8,
    txHash: "SQA7S37MCLGHQRMFZHRNUNUFJ6PJKRZN5RO52NMEWJU5B365SINQ",
    blockHeight: 40085318n,
    blockTime: 1719311180,
    attributes: {
      sender: "67e93fa6c8ac5c819990aa7340c0c16b508abb1178be9b30d024b8ac25193d45",
      sequence: 10576,
      payload: "AAAAADcXNho=",
      nonce: 0,
      consistencyLevel: 0,
      protocol: "Token Bridge",
    },
  };
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
  handleAptosTransactions = new HandleAlgorandTransactions(
    cfg,
    mapper,
    () => Promise.resolve(),
    statsRepo
  );
};

const givenConfig = () => {
  cfg = {
    id: "poll-log-message-published-algorand",
    metricName: "process_source_event",
    filter: [
      {
        applicationIds: "842125965",
        applicationAddress: "J476J725L4JTOI2YU6DAI4E23LYUECLZR7RCYZ3LK6QFHX4M54ZI53SGXQ",
      },
    ],
  };
};

const givenStatsRepository = () => {
  statsRepo = {
    count: () => {},
    measure: () => {},
    report: () => Promise.resolve(""),
  };
};

txs = [
  {
    payload: "AAAAADcXNho=",
    applicationId: "842126029",
    blockNumber: 40085318,
    timestamp: 1719311180,
    innerTxs: [
      {
        logs: ["AAAAAAAAKVA="],
        sender: "M7UT7JWIVROIDGMQVJZUBQGBNNIIVOYRPC7JWMGQES4KYJIZHVCRZEGFRQ",
      },
    ],
    sender: "MG3DIJNS3JTVKUAQGFV5BQTDAK26OUM3SRXSLIFWVUS67V54VPKDUJQTOQ",
    hash: "SQA7S37MCLGHQRMFZHRNUNUFJ6PJKRZN5RO52NMEWJU5B365SINQ",
  },
];
