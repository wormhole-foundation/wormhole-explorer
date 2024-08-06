import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { NearTransaction } from "../../../../src/domain/entities/near";
import { StatRepository } from "../../../../src/domain/repositories";
import { LogFoundEvent } from "../../../../src/domain/entities";
import {
  HandleNearTransactionsOptions,
  HandleNearTransactions,
} from "../../../../src/domain/actions/near/HandleNearTransactions";

let targetRepoSpy: jest.SpiedFunction<(typeof targetRepo)["save"]>;
let statsRepo: StatRepository;

let handleNearTransactions: HandleNearTransactions;
let txs: NearTransaction[];
let cfg: HandleNearTransactionsOptions;

describe("HandleNearTransactions", () => {
  afterEach(async () => {});

  it("should be able to map source events tx", async () => {
    // Given
    givenConfig();
    givenStatsRepository();
    givenHandleNearLogs();

    // When
    const result = await handleNearTransactions.handle(txs);

    // Then
    expect(result).toHaveLength(1);
    expect(result[0].chainId).toBe(15);
    expect(result[0].txHash).toBe("DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj");
    expect(result[0].name).toBe("transfer-redeemed");
  });
});

const mapper = (tx: NearTransaction) => {
  return {
    name: "transfer-redeemed",
    address: "contract.portalbridge.near",
    blockHeight: 124531378n,
    blockTime: 1722257013,
    chainId: 15,
    txHash: "DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj",
    attributes: {
      consistencyLevel: 32,
      from: "tkng.near",
      emitterChain: 1,
      emitterAddress: "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
      sequence: 917786,
      nonce: 16419,
      status: "completed",
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

const givenHandleNearLogs = (targetFn: "save" | "failingSave" = "save") => {
  targetRepoSpy = jest.spyOn(targetRepo, targetFn);
  handleNearTransactions = new HandleNearTransactions(
    cfg,
    mapper,
    () => Promise.resolve(),
    statsRepo
  );
};

const givenConfig = () => {
  cfg = {
    metricName: "process_source_event",
    id: "poll-redeemed-transactions-near",
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
    receiverId: "contract.portalbridge.near",
    signerId: "tkng.near",
    timestamp: 1722257013,
    blockHeight: 124531378n,
    chainId: 15,
    hash: "DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj",
    logs: [
      {
        outcome: {
          logs: [],
        },
      },
      {
        outcome: {
          logs: ["token-bridge/src/lib.rs#1265: refunding 99220000000000000000000 to tkng.near?"],
        },
      },
    ],
    actions: [
      {
        functionCall: {
          method: "submit_vaa",
          args: "eyJ2YWEiOiIwMTAwMDAwMDA0MGQwMDEzZTA2ZDFkNzgyMDFiZjQwNDhlMDg4MmJmZTJlNjQ5NjE4ODQzYjI1MDg4NjM3NDVlMTdiYTlhYmM4ZWFlMjAzOTA0MmE1ZTY5NmU3ZmU0ZjRiNjNkOGRjMGI4ZDdhNWZkNDNkNWViMjVhYWI2MmIwZDRlMTFjODZjZjUwMTliMDAwMmUxMjAxNzA5Y2Q4NmZiNDhjYWQyZjlhMTY3OGY5NzZlYmZiYzQ5NGZhZmYzNzZhNDRiMmQwNmY2NGQ5MzI5OWQ2MTQzMjJiYWIzZmJkMzFkZGFjMzZkZjQxYjI2MjM1OGI3MjRmYjg2ZDFjN2U0ODY3ZDk5MDkyYjFhZjgzMzY2MDAwNDE1M2VlMjIwZDAxNzlmNzY5ZmUxZjdjNmQ2ODk5ZWQ5NWRhZDAxNzZjMmFhM2FjZTk2ZjQyOWU5MzU5MjQ2NWY3OGM0MTBhM2FkNWMzZDFkNjZmNzkxZjBhZmE5M2QyZTEwMzcyOWMzOGZmMTg4MzQwNjhlOTU0NWMyZDUxNGE1MDAwNjYxMDUwMjk5ZmJkZjJjYmY5OTM0ZTRhMGE3OTk1OWQzMDhiNzE5MWI2OTNjMWVkODM5MWY4MzZiOTZkNWYxZmQ1ZGQ4ZDQyNmM1NWMwZTZhZDBkMTllNWEzYzVhMGQ4OTQ5MjY4ZTg3NDk3MmVhN2MxNjc4MmE3MDhjZTI5ZDNjMDAwN2Q2ZmZhZjYzODFmMGY2ZWI4NGNlODMyOTIwMGFmNTc5YWEyNTE0YjVmYzQxOTcxMDhkNDU3YmIwOTc4OGNkODk0MmYxNWQ4NWJkZDljOGFmNzBlNjIzZjgxOWM2M2IyNDZlZWQyZGYxZjcyZTI3MWJjYjM4MTNiNWNjNmI5MDFiMDAwOGZjNGNjODhiYTg3NzNhYTQ1Mzc2MjQ1OTY2Y2I3OTY5NDE4MWM5ZTRkYTE2ZTU0ZTU4ZjI4M2VmZTU2YmEyMTM0NjY0NWE1YzY1NzVlNWEyNTFjNDM5Yjk0Yzg3YzEwNzFiNjA5YWE0ZjkyMzU3ZjQ0YTU0NzgyM2FjYWRmNjAxMDAwOWFhYWVkNGY1NDZhMzQ4ZWRmYWIyOGExODMwZGNmZmU3ODIyOTU2ZTU1ZmU3OTFhMDQ5YTdkMDMyMThjY2U0OTMxN2IyZTE3MDk4OGY3ZDRkYWU0YzIyY2MwYjg1OWZmMGFiNTlmZGYzYzMzNTVlNDk3ODAzZjJjYzg0ZjM1YjVmMDEwYTRmZTFiMjcxZGU2ODc2YzIzYjQ5MTFkMTk1YmUwZjY1ZDljNmZiYzBiNjAzZThkNzU2ZDBjZjc5ZDM0YTA2YzI2N2IyODk3ZTY3MGU4M2Y3MzEwN2U3ZjQ4ZmNlMWM3YmFjYjAwNDJmN2M1OGZjNzcxYmMxNzA5ZDk1MmQ1NDY1MDEwZWI2ODM2MGRmNDNlM2U4ZTg1ZGVjODdlYzVmYzc3NDhhZWU2NDAxMzMwMzJhYmQxMTZiMjg0OTdmN2YxZGJlYTM2NjU4Yjg5N2FhZjI3NzEzODNlMWNhZDk4NmIyODcyYWRiYjhmNmQxZTI4ZTAwYjdlMGUyZGQ4ODk2MTZkMWJhMDAwZmYzMWNlYTMzZWRmZTBkNzYzOTI3NWE4Y2RkMmIyODRhYWZkMGRhOWFlMjU3NzBmOGEwMzU3NmQyNzEwNGRlNzc2ZDNmMzk3N2Y4NDk1OTAxZGRmYzQ0MGVkYTc1MDkyNGFiNWE1NWM3ZTA3MWI1N2M5ZjE0NWZhNzEyOWY5OTJlMDExMGYzM2JkNTlmZTc5MmFjOWFhNTlmMDNlNjBjYTU3M2VhM2I2MGQyYzVkNjkzZTY5ODM3ZGJlMTI2MDNhZDU1ZmE1YTIyYWNmMjQyNzc0YWJlMWRmNWI0N2E1NjcyNzBhNWRkZWE3YmQ3NTU5Mjg5OWNjMTA2YzYzOWQzMGJjNmIzMDAxMTg1MzY4NjAwYWRjZTBlMjcyZTU2MGNmMTUwNTAzMDNjMTcyOTg0YTcyZjgwNjYzZDkwNmNmZGVkNWZjNzcwYWMxODdkOTA2Y2RkN2IxM2RiMjE0Njc3YmYyNDRmYjk3N2ZjOTc3MDg5ODZjYzViZTdhMDBiMDE5MzNlZTU5ZDAzMDExMjgxN2U2ODRlYzhiZmQ0MWE5ZGU5NjM1ZjBhMTU3NzIzZDUwZjU5Yjg2ODAxNjkyN2ZjZjU2MDc3MWY0OGE4OTU3MjE2MzRjNDEyOTA0MzZhNWFkOWY1OTg2ZGI0NGI1Y2E5N2VkOTk5NWEzYmIxYzUwODAxNTViODQwNGJhNzRhMDE2NmE3OGUyMzAwMDA0MDIzMDAwMWVjNzM3Mjk5NWQ1Y2M4NzMyMzk3ZmIwYWQzNWMwMTIxZTBlYWE5MGQyNmY4MjhhNTM0Y2FiNTQzOTFiM2E0ZjUwMDAwMDAwMDAwMGUwMTFhMjAwMTAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMjNjMzQ2MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwZmUxM2MzMjA4ZDAyNGQ0MDIwMjEyZDU1OGNmNjAzMDc0ZTk4YmYzMTRkZmRmMTI3Y2MwZTEzOGZiYzcxMmFhZGYwMDBmMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMCJ9",
        },
      },
    ],
  },
];
