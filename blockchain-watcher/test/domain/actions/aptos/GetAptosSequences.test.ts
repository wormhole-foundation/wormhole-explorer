import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { TransactionsByVersion } from "../../../../src/infrastructure/repositories/aptos/AptosJsonRPCBlockRepository";
import { thenWaitForAssertion } from "../../../wait-assertion";
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

let getTransactionsByVersionsForSourceEventSpy: jest.SpiedFunction<
  AptosRepository["getTransactionsByVersionForSourceEvent"]
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
let pollAptos: PollAptos;

let props = {
  blockBatchSize: 100,
  fromBlock: 0n,
  toBlock: 0n,
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
    type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
  },
  chain: "aptos",
  id: "poll-log-message-published-aptos",
};

let cfg = new PollAptosTransactionsConfig(props);

describe("GetAptosSequences", () => {
  afterEach(async () => {
    await pollAptos.stop();
  });

  it("should be not generate range (fromBlock and toBlock) and search the latest block plus block batch size cfg", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository();
    givenStatsRepository();
    givenPollAptosTx(cfg);

    // When
    await whenPollEvmLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { fromBlock: undefined, toBlock: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
            type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
          }
        ),
      () => expect(getTransactionsByVersionsForSourceEventSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be use fromBlock and batch size cfg from cfg", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository();
    givenStatsRepository();
    // Se fromBlock for cfg
    props.fromBlock = 146040n;
    givenPollAptosTx(cfg);

    // When
    await whenPollEvmLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { fromBlock: 146040, toBlock: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
            type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
          }
        ),
      () => expect(getTransactionsByVersionsForSourceEventSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be return the same lastBlock and toBlock equal 100", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository({ previousBlock: 146040n, lastBlock: 146040n });
    givenStatsRepository();
    // Se fromBlock for cfg
    props.fromBlock = 0n;
    givenPollAptosTx(cfg);

    // When
    await whenPollEvmLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { fromBlock: 146040, toBlock: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
            type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
          }
        ),
      () => expect(getTransactionsByVersionsForSourceEventSpy).toHaveReturnedTimes(1)
    );
  });

  it("should be if return the lastBlock and toBlock equal the block batch size", async () => {
    // Given
    givenAptosBlockRepository();
    givenMetadataRepository({ previousBlock: undefined, lastBlock: 146040n });
    givenStatsRepository();
    // Se fromBlock for cfg
    props.fromBlock = 0n;
    givenPollAptosTx(cfg);

    // When
    await whenPollEvmLogsStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getSequenceNumberSpy).toHaveReturnedTimes(1),
      () =>
        expect(getSequenceNumberSpy).toBeCalledWith(
          { fromBlock: 146040, toBlock: 100 },
          {
            address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
            event:
              "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessageHandle",
            fieldName: "event",
            type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
          }
        ),
      () => expect(getTransactionsByVersionsForSourceEventSpy).toHaveReturnedTimes(1)
    );
  });
});

const givenAptosBlockRepository = () => {
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

  aptosRepo = {
    getSequenceNumber: () => Promise.resolve(events),
    getTransactionsByVersionForSourceEvent: () => Promise.resolve(txs),
    getTransactionsByVersionForRedeemedEvent: () => Promise.resolve(txs),
    getTransactions: () => Promise.resolve(txs),
  };

  getSequenceNumberSpy = jest.spyOn(aptosRepo, "getSequenceNumber");
  getTransactionsByVersionsForSourceEventSpy = jest.spyOn(
    aptosRepo,
    "getTransactionsByVersionForSourceEvent"
  );
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
  pollAptos = new PollAptos(cfg, statsRepo, metadataRepo, aptosRepo, "GetAptosSequences");
};

const whenPollEvmLogsStarts = async () => {
  pollAptos.run([handlers.working]);
};
