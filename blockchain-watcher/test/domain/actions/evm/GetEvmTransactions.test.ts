import { describe, it, expect, jest } from "@jest/globals";
import { GetEvmTransactions } from "../../../../src/domain/actions/evm/GetEvmTransactions";
import { EvmBlockRepository } from "../../../../src/domain/repositories";
import { EvmBlock, EvmLog } from "../../../../src/domain/entities/evm";

let getTransactionReceipt: jest.SpiedFunction<EvmBlockRepository["getTransactionReceipt"]>;
let getBlockSpy: jest.SpiedFunction<EvmBlockRepository["getBlock"]>;

let getEvmTransactions: GetEvmTransactions;
let evmBlockRepo: EvmBlockRepository;

describe("GetEvmTransactions", () => {
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
      topics: [],
      chain: "ethereum",
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
      expect(response[0].methodsByAddress).toEqual("MethodCompleteTransfer");

      expect(getTransactionReceipt).toHaveReturnedTimes(1);
      expect(getBlockSpy).toHaveReturnedTimes(1);
    });
  });
});

const givenEvmBlockRepository = (height?: bigint, blocksAhead?: bigint) => {
  const logsResponse: EvmLog[] = [];
  const blocksResponse: Record<string, EvmBlock> = {};
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
            hash: "dasdasfpialsfijlasfsahuf",
            from: "0x3ee123456786797000d974cf647e7c347e8fa585",
            to: "0x3ee18b2214aff97000d974cf647e7c347e8fa585",
            blockNumber: 1n,
            topics: [],
            input: "0xc687851912312444wadadswadwd",
            data: "",
            chainId: 1,
          },
        ],
      };
    }
  }

  evmBlockRepo = {
    getBlocks: () => Promise.resolve(blocksResponse),
    getBlockHeight: () => Promise.resolve(height ? height + (blocksAhead ?? 10n) : 10n),
    getFilteredLogs: () => Promise.resolve(logsResponse),
    getTransactionReceipt: () => Promise.resolve("0x1"),
    getBlock: () => Promise.resolve(blocksResponse[`0x01`]),
  };

  getBlockSpy = jest.spyOn(evmBlockRepo, "getBlock");
  getTransactionReceipt = jest.spyOn(evmBlockRepo, "getTransactionReceipt");
};

const givenPollEvmLogs = () => {
  getEvmTransactions = new GetEvmTransactions(evmBlockRepo);
};
