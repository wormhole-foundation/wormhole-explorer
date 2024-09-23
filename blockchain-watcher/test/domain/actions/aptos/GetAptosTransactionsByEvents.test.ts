import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { thenWaitForAssertion } from "../../../waitAssertion";
import { AptosTransaction } from "../../../../src/domain/entities/aptos";
import {
  PollAptosTransactionsMetadata,
  PollAptosTransactionsConfig,
  PollAptos,
} from "../../../../src/domain/actions/aptos/PollAptos";
import {
  MetadataRepository,
  AptosRepository,
  StatRepository,
} from "../../../../src/domain/repositories";

let getTransactionsByVersionSpy: jest.SpiedFunction<AptosRepository["getTransactionsByVersion"]>;
let getSequenceNumberSpy: jest.SpiedFunction<AptosRepository["getEventsByEventHandle"]>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollAptosTransactionsMetadata>["save"]>;

let handlerSpy: jest.SpiedFunction<(txs: AptosTransaction[]) => Promise<void>>;

let metadataRepo: MetadataRepository<PollAptosTransactionsMetadata>;
let aptosRepo: AptosRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (txs: AptosTransaction[]) => Promise.resolve(),
  failing: (txs: AptosTransaction[]) => Promise.reject(),
};
let pollAptos: PollAptos;

let props = {
  blockBatchSize: 100,
  from: 0n,
  limit: 0n,
  environment: "testnet",
  commitment: "finalized",
  addresses: ["0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625"],
  interval: 5000,
  topics: [],
  chainId: 22,
  filters: [
    {
      fieldName: "event",
      address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
      event:
        "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
      type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
    },
  ],
  chain: "aptos",
  id: "poll-log-message-published-aptos",
};

let cfg = new PollAptosTransactionsConfig(props);

describe("GetAptosTransactionsByEvents", () => {
  afterEach(async () => {
    await pollAptos.stop();
  });

  it("should be return an empty array and not to run the process because the newLastFrom is minor than lastFrom", async () => {
    // Given
    givenAptosBlockRepository("6040");
    givenMetadataRepository({ previousFrom: 146040n, lastFrom: 146140n });
    givenStatsRepository();
    givenPollAptosTx(cfg);

    // When
    await whenPollAptosLogsStarts();

    // Then
    // previousFrom: 146040n, lastFrom: 146140n, newLastFrom: 6040n from the rpc response
    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { from: 146140, limit: 101 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
            type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
          }
        )
    );
  });

  it("should be not generate range (from and limit) and search the latest from plus from batch size cfg", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository({ previousFrom: undefined, lastFrom: 146040n });
    givenStatsRepository();
    givenPollAptosTx(cfg);

    // When
    await whenPollAptosLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { from: 146040, limit: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
            type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
          }
        ),
      () => expect(getTransactionsByVersionSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be use from and batch size cfg from cfg", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository();
    givenStatsRepository();
    // Se from for cfg
    props.from = 146040n;
    givenPollAptosTx(cfg);

    // When
    await whenPollAptosLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { from: 146040, limit: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
            type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
          }
        ),
      () => expect(getTransactionsByVersionSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be return the same lastFrom and limit equal 100", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository({ previousFrom: 146040n, lastFrom: 146040n });
    givenStatsRepository();
    // Se from for cfg
    props.from = 0n;
    givenPollAptosTx(cfg);

    // When
    await whenPollAptosLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { from: 146040, limit: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
            type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
          }
        ),
      () => expect(getTransactionsByVersionSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be if return the lastFrom and limit equal the from batch size", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository({ previousFrom: undefined, lastFrom: 146040n });
    givenStatsRepository();
    // Se from for cfg
    props.from = 0n;
    givenPollAptosTx(cfg);

    // When
    await whenPollAptosLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { from: 146040, limit: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
            type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
          }
        ),
      () => expect(getTransactionsByVersionSpy).toHaveReturnedTimes(1)
    );
  });
});

const givenAptosBlockRepository = (sequenceNumber: string = "148985") => {
  const events = [
    {
      version: "481740133",
      guid: {
        creation_number: "2",
        account_address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
      },
      sequence_number: sequenceNumber,
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
      timestamp: 170963869344,
      blockTime: 170963869344,
      sequence: 3423n,
      version: "482649547",
      payload:
        "0x01000000000000000000000000000000000000000000000000000000000097d3650000000000000000000000003c499c542cef5e3811e1192ce70d8cc03d5c3359000500000000000000000000000081c1980abe8971e14865a629dd75b07621db1ae100050000000000000000000000000000000000000000000000000000000000001fe0",
      address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
      sender: "0x5aa807666de4dd9901c8f14a2f6021e85dc792890ece7f6bb929b46dba7671a2",
      status: true,
      to: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
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

  aptosRepo = {
    getEventsByEventHandle: () => Promise.resolve(events),
    getTransactionsByVersion: () => Promise.resolve(txs),
    getTransactions: () => Promise.resolve([]),
    healthCheck: () => Promise.resolve(),
  };

  getSequenceNumberSpy = jest.spyOn(aptosRepo, "getEventsByEventHandle");
  getTransactionsByVersionSpy = jest.spyOn(aptosRepo, "getTransactionsByVersion");
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
  pollAptos = new PollAptos(
    cfg,
    statsRepo,
    metadataRepo,
    aptosRepo,
    "GetAptosTransactionsByEvents"
  );
};

const whenPollAptosLogsStarts = async () => {
  pollAptos.run([handlers.working]);
};
