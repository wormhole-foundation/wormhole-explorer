import { afterAll, afterEach, describe, it, expect, jest } from "@jest/globals";
import { GetEvmTransactions } from "../../../../src/domain/actions/evm/GetEvmTransactions";
import { EvmBlockRepository } from "../../../../src/domain/repositories";
import { EvmBlock, EvmLog, ReceiptTransaction } from "../../../../src/domain/entities/evm";

let getTransactionReceipt: jest.SpiedFunction<EvmBlockRepository["getTransactionReceipt"]>;
let getBlockSpy: jest.SpiedFunction<EvmBlockRepository["getBlock"]>;

let getEvmTransactions: GetEvmTransactions;
let evmBlockRepo: EvmBlockRepository;

describe("GetEvmTransactions", () => {
  afterAll(() => {
    jest.clearAllMocks();
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it("should be return empty array, because formBlock is higher than toBlock", async () => {
    // Given
    const range = {
      fromBlock: 10n,
      toBlock: 1n,
    };

    const opts = {
      addresses: [],
      topics: [],
      chain: "ethereum",
      chainId: 1,
      environment: "testnet",
    };

    givenPollEvmLogs();

    // When
    const result = getEvmTransactions.execute(range, opts);

    // Then
    result.then((response) => {
      expect(response).toEqual([]);
    });
  });

  it("should be return empty array, because do not match any contract address with transaction address", async () => {
    // Given
    const range = {
      fromBlock: 1n,
      toBlock: 1n,
    };

    const opts = {
      addresses: ["0x1ee18b2214aff97000d974cf647e7c545e8fa585"],
      topics: [],
      chain: "ethereum",
      chainId: 1,
      environment: "mainnet",
    };

    givenEvmBlockRepository(range.fromBlock, range.toBlock);
    givenPollEvmLogs();

    // When
    const result = getEvmTransactions.execute(range, opts);

    // Then
    result.then((response) => {
      expect(response).toEqual([]);
      expect(getBlockSpy).toHaveReturnedTimes(1);
    });
  });

  it("should be return array with one transaction filter and populated", async () => {
    // Given
    const range = {
      fromBlock: 1n,
      toBlock: 1n,
    };

    const opts = {
      addresses: ["0x3ee18b2214aff97000d974cf647e7c347e8fa585"],
      topics: ["0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9a"],
      chain: "ethereum",
      chainId: 1,
      environment: "mainnet",
    };

    givenEvmBlockRepository(range.fromBlock, range.toBlock);
    givenPollEvmLogs();

    // When
    const result = getEvmTransactions.execute(range, opts);

    // Then
    result.then((response) => {
      expect(response.length).toEqual(1);
      expect(response[0].chainId).toEqual(1);
      expect(response[0].status).toEqual("0x1");
      expect(response[0].from).toEqual("0x3ee123456786797000d974cf647e7c347e8fa585");
      expect(response[0].to).toEqual("0x3ee18b2214aff97000d974cf647e7c347e8fa585");
      expect(getTransactionReceipt).toHaveReturnedTimes(1);
      expect(getBlockSpy).toHaveReturnedTimes(1);
    });
  });

  it("should be return array with two transaction filter and populated with redeemed and MintAndWithdraw transaction log", async () => {
    // Given
    const range = {
      fromBlock: 1n,
      toBlock: 2n,
    };

    const opts = {
      addresses: [
        "0x4cb69fae7e7af841e44e1a1c30af640739378bb2",
        "0xBd3fa81B58Ba92a82136038B25aDec7066af3155",
      ],
      topics: [
        "0x1b2a7ff080b8cb6ff436ce0372e399692bbfb6d4ae5766fd8d58a7b8cc6142e6",
        "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e",
      ],
      chain: "ethereum",
      chainId: 1,
      environment: "mainnet",
    };

    const logs = [
      {
        address: "0xBd3fa81B58Ba92a82136038B25aDec7066af3155",
        topics: ["0x1b2a7ff080b8cb6ff436ce0372e399692bbfb6d4ae5766fd8d58a7b8cc6142e6"],
      },
      {
        address: "0xBd3fa81B58Ba92a82136038B25aDec7066af3155",
        topics: [
          "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e",
          "0x0000000000000000000000000000000000000000000000000000000000000017",
          "0x0000000000000000000000002703483b1a5a7c577e8680de9df8be03c6f30e3c",
          "0x000000000000000000000000000000000000000000000000000000000000250f",
        ],
      },
    ];

    givenEvmBlockRepository(
      range.fromBlock,
      range.toBlock,
      logs,
      "0x4cb69fae7e7af841e44e1a1c30af640739378bb2"
    );
    givenPollEvmLogs();

    // When
    const result = getEvmTransactions.execute(range, opts);

    // Then
    result.then((response) => {
      expect(response.length).toEqual(2);
      expect(response[0].chainId).toEqual(1);
      expect(response[0].status).toEqual("0x1");
      expect(response[0].from).toEqual("0x3ee123456786797000d974cf647e7c347e8fa585");
      expect(response[0].to).toEqual("0x4cb69fae7e7af841e44e1a1c30af640739378bb2");
      expect(getTransactionReceipt).toHaveReturnedTimes(2);
      expect(getBlockSpy).toHaveReturnedTimes(2);
    });
  });
});

const givenEvmBlockRepository = (
  height?: bigint,
  blocksAhead?: bigint,
  logs?: any,
  to?: string
) => {
  let logsMock = logs;
  const toMock = to ?? "0x3ee18b2214aff97000d974cf647e7c347e8fa585";

  if (!logs) {
    logsMock = [
      {
        address: "0xf890982f9310df57d00f659cf4fd87e65aded8d7",
        topics: ["0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e"],
      },
      {
        address: "0xf890982f9310df57d00f659cf4fd87e65aded8d7",
        topics: ["0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9a"],
      },
    ];
  }

  const logsResponse: EvmLog[] = [];
  const blocksResponse: Record<string, EvmBlock> = {};
  const receiptResponse: Record<string, ReceiptTransaction> = {};
  if (height) {
    for (let index = 0n; index <= (blocksAhead ?? 1n); index++) {
      logsResponse.push({
        blockNumber: height + index,
        blockHash: `0x0${index}`,
        blockTime: 0,
        address: "",
        removed: false,
        data: "",
        transactionHash: "",
        transactionIndex: "",
        topics: [],
        logIndex: 0,
        chainId: 2,
      });
      blocksResponse[`0x0${index}`] = {
        timestamp: 0,
        hash: `huohugigiyyff6677rr657s7xr8copi`,
        number: height + index,
        transactions: [
          {
            blockHash: "0xf5794b0970386d7951e45465ac2c9835537e5a9",
            hash: "dasdasfpialsfijlasfsahuf",
            blockNumber: 1n,
            chainId: 1,
            from: "0x3ee123456786797000d974cf647e7c347e8fa585",
            gas: "0x14485",
            gasPrice: "0xfc518561e",
            input: "0xc687851912312444wadadswadwd",
            maxFeePerGas: "0x1610f75b9a",
            maxPriorityFeePerGas: "0x5f5e100",
            nonce: "0x1",
            r: "0xf5794b0970386d73b693b17f147fae0427db278e951e45465ac2c9835537e5a9",
            s: "0x6dccc8cfee216bc43a9d66525fa94905da234ad32d6cc3220845bef78f25dd42",
            status: "0x1",
            timestamp: 12313123,
            to: toMock,
            transactionIndex: "0x6f",
            type: "0x2",
            v: "0x1",
            value: "0x5b09cd3e5e90000",
            environment: "testnet",
            chain: "ethereum",
            logs: logsMock,
          },
        ],
      };
      receiptResponse["dasdasfpialsfijlasfsahuf"] = {
        status: "0x1",
        transactionHash: "dasdasfpialsfijlasfsahuf",
        logs: logsMock,
      };
    }
  }

  evmBlockRepo = {
    getBlocks: () => Promise.resolve(blocksResponse),
    getBlockHeight: () => Promise.resolve(height ? height + (blocksAhead ?? 10n) : 10n),
    getFilteredLogs: () => Promise.resolve(logsResponse),
    getTransactionReceipt: () => Promise.resolve(receiptResponse),
    getBlock: () => Promise.resolve(blocksResponse[`0x01`]),
  };

  getBlockSpy = jest.spyOn(evmBlockRepo, "getBlock");
  getTransactionReceipt = jest.spyOn(evmBlockRepo, "getTransactionReceipt");
};

const givenPollEvmLogs = () => {
  getEvmTransactions = new GetEvmTransactions(evmBlockRepo);
};
