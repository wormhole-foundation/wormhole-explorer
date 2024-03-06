import { afterAll, afterEach, describe, it, expect, jest } from "@jest/globals";
import { GetEvmTransactions } from "../../../../src/domain/actions/evm/GetEvmTransactions";
import { EvmBlockRepository } from "../../../../src/domain/repositories";
import {
  EvmBlock,
  EvmLog,
  EvmTransaction,
  ReceiptTransaction,
} from "../../../../src/domain/entities/evm";
import { randomBytes } from "crypto";

let getTransactionReceipt: jest.SpiedFunction<EvmBlockRepository["getTransactionReceipt"]>;
let getBlocksSpy: jest.SpiedFunction<EvmBlockRepository["getBlocks"]>;

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
    const result = await getEvmTransactions.execute(range, opts);

    // Then
    expect(result).toEqual([]);
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

    const blocks = {
      "0x01": new BlockBuilder()
        .number(1n)
        .txs([new TxBuilder().logs([]).to("0x3ee18b2214aff97000d974cf647e7c347e8fa585").create()])
        .create(),
    };

    givenEvmBlockRepository(range.fromBlock, range.toBlock, blocks);
    givenPollEvmLogs();

    // When
    const result = await getEvmTransactions.execute(range, opts);

    // Then
    expect(result).toEqual([]);
    expect(getBlocksSpy).toHaveReturnedTimes(1);
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

    const blocks = {
      "0x01": new BlockBuilder().number(1n).txs([new TxBuilder().create()]).create(),
    };

    givenEvmBlockRepository(range.fromBlock, range.toBlock, blocks);
    givenPollEvmLogs();

    // When
    const result = await getEvmTransactions.execute(range, opts);

    // Then
    expect(result.length).toEqual(1);
    expect(result[0].chainId).toEqual(1);
    expect(result[0].status).toEqual("0x1");
    expect(result[0].from).toEqual("0x3ee123456786797000d974cf647e7c347e8fa585");
    expect(result[0].to).toEqual("0x3ee18b2214aff97000d974cf647e7c347e8fa585");
    expect(getTransactionReceipt).toHaveReturnedTimes(1);
    expect(getBlocksSpy).toHaveReturnedTimes(1);
  });

  it("should return one transaction from a block with multiple", async () => {
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

    const blocks = {
      "0x01": new BlockBuilder()
        .number(1n)
        .txs([
          // different topic
          new TxBuilder()
            .hash("0x01")
            .logs([
              {
                address: "0x3ee18b2214aff97000d974cf647e7c347e8fa585",
                topics: ["0x1b2a7ff080b8cb6ff436ce0372e399692bbfb6d4ae5766fd8d58a7b8cc6142e6"],
                data: "0x0",
              },
            ])
            .create(),
          // matches filters
          new TxBuilder()
            .hash("0x02")
            .logs([
              {
                address: "0x3ee18b2214aff97000d974cf647e7c347e8fa585",
                topics: ["0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9a"],
                data: "0x0",
              },
            ])
            .create(),
          // different to address
          new TxBuilder()
            .hash("0x03")
            .to("0x4cb69fae7e7af841e44e1a1c30af640739378bb2")
            .logs([
              {
                address: "0x4cb69fae7e7af841e44e1a1c30af640739378bb2",
                topics: ["0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9a"],
                data: "0x0",
              },
            ])
            .create(),
          // different to address, but same log emitter
          new TxBuilder()
            .hash("0x04")
            .to("0x4cb69fae7e7af841e44e1a1c30af640739378bb2")
            .logs([
              {
                address: "0x3ee18b2214aff97000d974cf647e7c347e8fa585",
                topics: ["0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9a"],
                data: "0x0",
              },
            ])
            .create(),
        ])
        .create(),
    };

    givenEvmBlockRepository(range.fromBlock, range.toBlock, blocks);
    givenPollEvmLogs();

    // When
    const result = await getEvmTransactions.execute(range, opts);

    // Then
    expect(result.length).toEqual(1);
    expect(result[0].hash).toEqual("0x02");
    expect(result[0].to).toEqual("0x3ee18b2214aff97000d974cf647e7c347e8fa585");
    expect(getTransactionReceipt).toHaveReturnedTimes(1);
    expect(getBlocksSpy).toHaveReturnedTimes(1);
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

    const blocks = {
      "0x01": new BlockBuilder()
        .number(1n)
        .txs([new TxBuilder().logs(logs).to("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").create()])
        .create(),
      "0x02": new BlockBuilder()
        .number(2n)
        .txs([new TxBuilder().logs(logs).to("0x4cb69fae7e7af841e44e1a1c30af640739378bb2").create()])
        .create(),
    };

    givenEvmBlockRepository(range.fromBlock, range.toBlock, blocks);
    givenPollEvmLogs();

    // When
    const result = await getEvmTransactions.execute(range, opts);

    // Then
    expect(result.length).toEqual(2);
    expect(result[0].chainId).toEqual(1);
    expect(result[0].status).toEqual("0x1");
    expect(result[0].from).toEqual("0x3ee123456786797000d974cf647e7c347e8fa585");
    expect(result[0].to).toEqual("0x4cb69fae7e7af841e44e1a1c30af640739378bb2");
    expect(getTransactionReceipt).toHaveReturnedTimes(2);
    expect(getBlocksSpy).toHaveReturnedTimes(1);
  });

  it("should apply a multiple topics filter", async () => {
    // Given
    const range = {
      fromBlock: 1n,
      toBlock: 2n,
    };

    const opts = {
      addresses: ["0x7b1bd7a6b4e61c2a123ac6bc2cbfc614437d0470"],
      topics: [
        [
          "0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e",
          "0x504e6efe18ab9eed10dc6501a417f5b12a2f7f2b1593aed9b89f9bce3cf29a91",
        ],
      ],
      chain: "ethereum",
      chainId: 1,
      environment: "mainnet",
    };

    const logs = [
      {
        address: "0x00ac6efc189140b50a043b5e43c108cf571586d1",
        topics: ["0xf557dbbb087662f52c815f6c7ee350628a37a51eae9608ff840d996b65f87475"],
        data: "0xe9d6f4dbc1d568640ce3f6111b2d082e8282461feb9812135b30c9f7c1dcf300000000000000000000000000000000000000000000000000000000000000271200000000000000000000000055aaf4d9399c472b252e7c0b49408b5bc7d7328e",
      },
      {
        address: "0xcc1ebd7a6661c0f6e19d2bbdb881b11f3b3f40ff",
        topics: ["0x35a2101eaac94b493e0dfca061f9a7f087913fde8678e7cde0aca9897edba0e5"],
        data: "0xb20b3f32244182844595b9670c53aa82303829fe827af22c460458be9bbae85700000000000000000000000000ac6efc189140b50a043b5e43c108cf571586d10000000000000000000000000000000000000000000000000000000000000000",
      },
      {
        address: "0xcc1ebd7a6661c0f6e19d2bbdb881b11f3b3f40ff",
        topics: [
          "0x504e6efe18ab9eed10dc6501a417f5b12a2f7f2b1593aed9b89f9bce3cf29a91",
          "0xb20b3f32244182844595b9670c53aa82303829fe827af22c460458be9bbae857",
        ],
        data: "0x",
      },
      {
        address: "0xb12c77938c09d81f1e9797d48501b5c4e338b45b",
        topics: [
          "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
          "0x0000000000000000000000000000000000000000000000000000000000000000",
          "0x000000000000000000000000e6990c7e206d418d62b9e50c8e61f59dc360183b",
        ],
        data: "0x00000000000000000000000000000000000000000000000000354a6ba7a18000",
      },
      {
        address: "0x7b1bd7a6b4e61c2a123ac6bc2cbfc614437d0470",
        topics: [
          "0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e",
          "0x00000000000000000000000000ac6efc189140b50a043b5e43c108cf571586d1",
          "0x0000000000000000000000000000000000000000000000000000000000002712",
          "0x00000000000000000000000000000000000000000000000000000000000013ac",
        ],
        data: "0xe9d6f4dbc1d568640ce3f6111b2d082e8282461feb9812135b30c9f7c1dcf3000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002dd05000000000000000000000000000000000000000000000000000000000000000500000000000000000000000000000000000000000000000000000000000000c000000000000000000000000000000000000000000000000000000000000000e000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
      },
    ];

    const blocks = {
      "0x01": new BlockBuilder()
        .number(1n)
        .txs([new TxBuilder().logs(logs).to("0x7b1bd7a6b4e61c2a123ac6bc2cbfc614437d0470").create()])
        .create(),
      "0x02": new BlockBuilder()
        .number(2n)
        .txs([
          new TxBuilder().logs([logs[4]]).to("0x7b1bd7a6b4e61c2a123ac6bc2cbfc614437d0470").create(),
        ])
        .create(),
    };

    givenEvmBlockRepository(range.fromBlock, range.toBlock, blocks);
    givenPollEvmLogs();

    // When
    const result = await getEvmTransactions.execute(range, opts);

    // Then
    expect(result.length).toEqual(1);
    expect(result[0].chainId).toEqual(1);
    expect(result[0].status).toEqual("0x1");
    expect(result[0].from).toEqual("0x3ee123456786797000d974cf647e7c347e8fa585");
    expect(result[0].to).toEqual("0x7b1bd7a6b4e61c2a123ac6bc2cbfc614437d0470");
    expect(getTransactionReceipt).toHaveReturnedTimes(2);
    expect(getBlocksSpy).toHaveReturnedTimes(1);
  });
});

const givenEvmBlockRepository = (
  height?: bigint,
  blocksAhead?: bigint,
  blocks?: Record<string, EvmBlock>
) => {
  const logsResponse: EvmLog[] = [];
  const receiptResponse: Record<string, ReceiptTransaction> = Object.values(blocks || {})
    .map((b) => b.transactions || [])
    .flat()
    .reduce((acc, tx) => {
      acc[tx.hash] = {
        status: "0x1",
        transactionHash: tx.hash,
        logs: tx.logs,
      };
      return acc;
    }, {} as Record<string, ReceiptTransaction>);

  if (height) {
    for (let index = height; index <= (blocksAhead ?? 1n); index++) {
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
    }
  }

  evmBlockRepo = {
    getBlocks: () => Promise.resolve(blocks || {}),
    getBlockHeight: () => Promise.resolve(height ? height + (blocksAhead ?? 10n) : 10n),
    getFilteredLogs: () => Promise.resolve(logsResponse),
    getTransactionReceipt: () => Promise.resolve(receiptResponse),
    getBlock: () => Promise.resolve(blocks ? blocks[`0x01`] : new BlockBuilder().create()),
  };

  getBlocksSpy = jest.spyOn(evmBlockRepo, "getBlocks");
  getTransactionReceipt = jest.spyOn(evmBlockRepo, "getTransactionReceipt");
};

const givenPollEvmLogs = () => {
  getEvmTransactions = new GetEvmTransactions(evmBlockRepo);
};

class BlockBuilder {
  private block: EvmBlock;

  constructor() {
    this.block = this.default();
  }

  number(n: bigint) {
    this.block.number = n;
    return this;
  }

  txs(transactions: any) {
    this.block.transactions = transactions;
    return this;
  }

  create() {
    return this.block;
  }

  default() {
    return {
      timestamp: 0,
      hash: "1n",
      number: 1n,
    };
  }
}

class TxBuilder {
  private tx: EvmTransaction;

  constructor() {
    this.tx = this.default();
  }

  logs(logs: any) {
    this.tx.logs = logs;
    return this;
  }

  to(to: string) {
    this.tx.to = to;
    return this;
  }

  create() {
    return this.tx;
  }

  hash(hash: string) {
    this.tx.hash = hash;
    return this;
  }

  default() {
    return {
      blockHash: "0xf5794b0970386d7951e45465ac2c9835537e5a9",
      hash: "0x" + randomBytes(32).toString("hex"),
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
      to: "0x3ee18b2214aff97000d974cf647e7c347e8fa585",
      transactionIndex: "0x6f",
      type: "0x2",
      v: "0x1",
      value: "0x5b09cd3e5e90000",
      environment: "testnet",
      chain: "ethereum",
      logs: [
        {
          address: "0xf890982f9310df57d00f659cf4fd87e65aded8d7",
          topics: ["0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e"],
          data: "0x0",
        },
        {
          address: "0xf890982f9310df57d00f659cf4fd87e65aded8d7",
          topics: ["0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9a"],
          data: "0x0",
        },
      ],
    };
  }
}
