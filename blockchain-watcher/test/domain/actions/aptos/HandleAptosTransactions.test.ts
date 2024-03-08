import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { TransactionsByVersion } from "../../../../src/infrastructure/repositories/aptos/AptosJsonRPCBlockRepository";
import { StatRepository } from "../../../../src/domain/repositories";
import { LogFoundEvent } from "../../../../src/domain/entities";
import {
  HandleAptosTransactionsOptions,
  HandleAptosTransactions,
} from "../../../../src/domain/actions/aptos/HandleAptosTransactions";

let targetRepoSpy: jest.SpiedFunction<(typeof targetRepo)["save"]>;
let statsRepo: StatRepository;

let handleAptosTransactions: HandleAptosTransactions;
let txs: TransactionsByVersion[];
let cfg: HandleAptosTransactionsOptions;

describe("HandleAptosTransactions", () => {
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
    expect(result[0].chainId).toBe(22);
    expect(result[0].txHash).toBe(
      "0x99f9cd1ea181d568ba4d89e414dcf1b129968b1c805388f29821599a447b7741"
    );
    expect(result[0].address).toBe(
      "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625"
    );
  });
});

const mapper = (tx: TransactionsByVersion) => {
  return {
    name: "log-message-published",
    address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
    chainId: 22,
    txHash: "0x99f9cd1ea181d568ba4d89e414dcf1b129968b1c805388f29821599a447b7741",
    blockHeight: 153549311n,
    blockTime: 1709645685704036,
    attributes: {
      sender: "0xa216910c9a74291aa3e26135486399b7a04771977687c3da57a7498e77103658",
      sequence: 203,
      payload:
        "0x7b2274797065223a226d657373616765222c226d657373616765223a2268656c6c6f222c227369676e6174757265223a226d657373616765222c2276616c7565223a7b226d657373616765223a2268656c6c6f222c2274797065223a226d657373616765222c227369676e6174757265223a226d657373616765227d7d",
      nonce: 75952,
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
  handleAptosTransactions = new HandleAptosTransactions(
    cfg,
    mapper,
    () => Promise.resolve(),
    statsRepo
  );
};

const givenConfig = () => {
  cfg = {
    id: "poll-log-message-published-aptos",
    metricName: "process_source_event",
    metricLabels: {
      job: "poll-log-message-published-aptos",
      chain: "aptos",
      commitment: "immediate",
    },
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
    consistencyLevel: 0,
    blockHeight: 153517771n,
    timestamp: 170963869344,
    blockTime: 170963869344,
    sequence: 3423n,
    version: "482649547",
    payload:
      "0x01000000000000000000000000000000000000000000000000000000000097d3650000000000000000000000003c499c542cef5e3811e1192ce70d8cc03d5c3359000500000000000000000000000081c1980abe8971e14865a629dd75b07621db1ae100050000000000000000000000000000000000000000000000000000000000001fe0",
    address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
    sender: "0x5aa807666de4dd9901c8f14a2f6021e85dc792890ece7f6bb929b46dba7671a2",
    status: true,
    events: [
      {
        guid: {
          creation_number: "11",
          account_address: "0x5aa807666de4dd9901c8f14a2f6021e85dc792890ece7f6bb929b46dba7671a2",
        },
        sequence_number: "0",
        type: "0x1::coin::WithdrawEvent",
        data: { amount: "9950053" },
      },
      {
        guid: {
          creation_number: "3",
          account_address: "0x5aa807666de4dd9901c8f14a2f6021e85dc792890ece7f6bb929b46dba7671a2",
        },
        sequence_number: "16",
        type: "0x1::coin::WithdrawEvent",
        data: { amount: "0" },
      },
      {
        guid: {
          creation_number: "4",
          account_address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
        },
        sequence_number: "149041",
        type: "0x1::coin::DepositEvent",
        data: { amount: "0" },
      },
      {
        guid: {
          creation_number: "2",
          account_address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
        },
        sequence_number: "149040",
        type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
        data: {
          consistency_level: 0,
          nonce: "76704",
          payload:
            "0x01000000000000000000000000000000000000000000000000000000000097d3650000000000000000000000003c499c542cef5e3811e1192ce70d8cc03d5c3359000500000000000000000000000081c1980abe8971e14865a629dd75b07621db1ae100050000000000000000000000000000000000000000000000000000000000001fe0",
          sender: "1",
          sequence: "146094",
          timestamp: "1709638693",
        },
      },
      {
        guid: { creation_number: "0", account_address: "0x0" },
        sequence_number: "0",
        type: "0x1::transaction_fee::FeeStatement",
        data: {
          execution_gas_units: "7",
          io_gas_units: "12",
          storage_fee_octas: "0",
          storage_fee_refund_octas: "0",
          total_charge_gas_units: "19",
        },
      },
    ],
    nonce: 76704,
    hash: "0xb2fa774485ce02c5786475dd2d689c3e3c2d0df0c5e09a1c8d1d0e249d96d76e",
  },
];
