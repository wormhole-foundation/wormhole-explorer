import {
  PollAptosTransactions,
  PollAptosTransactionsConfig,
  PollAptosTransactionsMetadata,
} from "../../../../src/domain/actions/aptos/PollAptosTransactions";
import { afterEach, describe, it, expect, jest } from "@jest/globals";
import {
  AptosRepository,
  MetadataRepository,
  StatRepository,
} from "../../../../src/domain/repositories";
import { thenWaitForAssertion } from "../../../wait-assertion";
import { TransactionsByVersion } from "../../../../src/infrastructure/repositories/aptos/AptosJsonRPCBlockRepository";

let getTransactionsForVersionsSpy: jest.SpiedFunction<
  AptosRepository["getTransactionsForVersions"]
>;
let getSequenceNumberSpy: jest.SpiedFunction<AptosRepository["getSequenceNumber"]>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollAptosTransactionsMetadata>["save"]>;

let handlerSpy: jest.SpiedFunction<(txs: TransactionsByVersion[]) => Promise<void>>;

let metadataRepo: MetadataRepository<PollAptosTransactionsMetadata>;
let aptosRepo: AptosRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (txs: TransactionsByVersion[]) => Promise.resolve(),
  failing: (txs: TransactionsByVersion[]) => Promise.reject(),
};
let pollAptosTransactions: PollAptosTransactions;

let props = {
  blockBatchSize: 100,
  fromSequence: 0n,
  toSequence: 0n,
  environment: "testnet",
  commitment: "finalized",
  addresses: ["0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625"],
  interval: 5000,
  topics: [],
  chainId: 22,
  filter: {
    fieldName: "event",
    address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
    event:
      "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
  },
  chain: "aptos",
  id: "poll-log-message-published-aptos",
};

let cfg = new PollAptosTransactionsConfig(props);

describe("pollAptosTransactions", () => {
  afterEach(async () => {
    await pollAptosTransactions.stop();
  });

  it("should be not generate range (from and to sequence) and search the latest sequence plus block batch size cfg", async () => {
    givenEvmBlockRepository();
    givenMetadataRepository();
    givenStatsRepository();
    givenPollAptosTx(cfg);

    await whenPollEvmLogsStarts();

    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { fromSequence: undefined, toSequence: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
          }
        ),
      () => expect(getTransactionsForVersionsSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be use fromSequence and batch size cfg from cfg", async () => {
    givenEvmBlockRepository();
    givenMetadataRepository();
    givenStatsRepository();
    // Se fromSequence for cfg
    props.fromSequence = 146040n;
    givenPollAptosTx(cfg);

    await whenPollEvmLogsStarts();

    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { fromSequence: 146040, toSequence: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
          }
        ),
      () => expect(getTransactionsForVersionsSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be return the same last sequence and the to sequence equal 1", async () => {
    givenEvmBlockRepository();
    givenMetadataRepository({ previousSequence: 146040n, lastSequence: 146040n });
    givenStatsRepository();
    // Se fromSequence for cfg
    props.fromSequence = 0n;
    givenPollAptosTx(cfg);

    await whenPollEvmLogsStarts();

    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { fromSequence: 146040, toSequence: 1 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
          }
        ),
      () => expect(getTransactionsForVersionsSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be return the difference between the last sequence and the previous sequence plus 1", async () => {
    givenEvmBlockRepository();
    givenMetadataRepository({ previousSequence: 146000n, lastSequence: 146040n });
    givenStatsRepository();
    // Se fromSequence for cfg
    props.fromSequence = 0n;
    givenPollAptosTx(cfg);

    await whenPollEvmLogsStarts();

    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { fromSequence: 146040, toSequence: 41 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
          }
        ),
      () => expect(getTransactionsForVersionsSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be if return the last sequence and the to sequence equal the block batch size", async () => {
    givenEvmBlockRepository();
    givenMetadataRepository({ previousSequence: undefined, lastSequence: 146040n });
    givenStatsRepository();
    // Se fromSequence for cfg
    props.fromSequence = 0n;
    givenPollAptosTx(cfg);

    await whenPollEvmLogsStarts();

    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { fromSequence: 146040, toSequence: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
          }
        ),
      () => expect(getTransactionsForVersionsSpy).toHaveReturnedTimes(1)
    );
  });
});

const givenEvmBlockRepository = () => {
  const events = [
    {
      version: "481740133",
      guid: {
        creation_number: "2",
        account_address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
      },
      sequence_number: "148985",
      type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
      data: {
        consistency_level: 0,
        nonce: "41611",
        payload:
          "0x0100000000000000000000000000000000000000000000000000000000003b826000000000000000000000000082af49447d8a07e3bd95bd0d56f35241523fbab10017000000000000000000000000451febd0f01b9d6bda1a5b9d0b6ef88026e4a79100170000000000000000000000000000000000000000000000000000000000011c1e",
        sender: "1",
        sequence: "146040",
        timestamp: "1709585379",
      },
    },
  ];

  const txs = [
    {
      consistencyLevel: 0,
      blockHeight: 153517771n,
      timestamp: "1709638693443328",
      blockTime: 1709638693443328,
      sequence: "34",
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
      nonce: "76704",
      hash: "0xb2fa774485ce02c5786475dd2d689c3e3c2d0df0c5e09a1c8d1d0e249d96d76e",
    },
  ];

  aptosRepo = {
    getSequenceNumber: () => Promise.resolve(events),
    getTransactionsForVersions: () => Promise.resolve(txs),
  };

  getSequenceNumberSpy = jest.spyOn(aptosRepo, "getSequenceNumber");
  getTransactionsForVersionsSpy = jest.spyOn(aptosRepo, "getTransactionsForVersions");
  handlerSpy = jest.spyOn(handlers, "working");
};

const givenMetadataRepository = (data?: PollAptosTransactionsMetadata) => {
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

const givenPollAptosTx = (cfg: PollAptosTransactionsConfig) => {
  pollAptosTransactions = new PollAptosTransactions(cfg, statsRepo, metadataRepo, aptosRepo);
};

const whenPollEvmLogsStarts = async () => {
  pollAptosTransactions.run([handlers.working]);
};
