import { describe, it, expect } from "@jest/globals";
import { evmRedeemedTransactionFoundMapper } from "../../../../src/infrastructure/mappers/evm/evmRedeemedTransactionFoundMapper";
import { HandleEvmTransactions } from "../../../../src/domain/actions";

const address = "0x3ee18B2214AFF97000D974cf647E7C347E8fa585";
const topic = "0xbccc00b713f54173962e7de6098f643d8ebf53d488d71f4b2a5171496d038f9e";
const txHash = "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d";

let statsRepo = {
  count: () => {},
  measure: () => {},
  report: () => Promise.resolve(""),
};

const handler = new HandleEvmTransactions(
  {
    abi: "event Delivery(address indexed recipientContract, uint16 indexed sourceChain, uint64 indexed sequence, bytes32 deliveryVaaHash, uint8 status, uint256 gasUsed, uint8 refundStatus, bytes additionalStatusInfo, bytes overridesInfo)",
    metricName: "process_vaa_ethereum_event",
    commitment: "latest",
    chainId: 2,
    chain: "ethereum",
    id: "poll-log-message-published-ethereum",
  },
  evmRedeemedTransactionFoundMapper,
  async () => {},
  statsRepo
);

describe("evmRedeemedTransactionFoundMapper", () => {
  it("should be able to map log to evmRedeemedTransactionFoundMapper without vaaInformation", async () => {
    // When
    const [result] = await handler.handle([
      {
        blockHash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        blockNumber: 0x11ec2bcn,
        chainId: 1,
        from: "0xfb070adcd21361a3946a0584dc84a7b89faa68e3",
        gas: "0x14485",
        gasPrice: "0xfc518561e",
        hash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        input:
          "0xc68785190000000000000000000000000000000000000000000000000000000000000001637651ef71f834be28b8fab1dce9c228c2fe1813831bbc3673cfd3abde6dbb3d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080420000",
        maxFeePerGas: "0x1610f75b9a",
        maxPriorityFeePerGas: "0x5f5e100",
        nonce: "0x1",
        r: "0xf5794b0970386d73b693b17f147fae0427db278e951e45465ac2c9835537e5a9",
        s: "0x6dccc8cfee216bc43a9d66525fa94905da234ad32d6cc3220845bef78f25dd42",
        status: "0x1",
        timestamp: 1702663079,
        to: "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
        transactionIndex: "0x6f",
        type: "0x2",
        v: "0x1",
        value: "0x5b09cd3e5e90000",
        environment: "testnet",
        chain: "ethereum",
        logs: [
          {
            address: "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
            topics: [
              "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759b7e0c00a",
              "0x0000000000000000000000000000000000000000000000000000000000000017",
              "0x0000000000000000000000002703483b1a5a7c577e8680de9df8be03c6f30e3c",
              "0x000000000000000000000000000000000000000000000000000000000000250f",
            ],
            data: "0x",
          },
        ],
      },
    ]);

    // Then
    expect(result?.name).toBe("transfer-redeemed");
    expect(result?.chainId).toBe(1);
    expect(result?.txHash).toBe(txHash);
    expect(result?.blockHeight).toBe(18793148n);
    expect(result?.attributes.blockNumber).toBe(18793148n);
    expect(result?.attributes.from).toBe("0xfb070adcd21361a3946a0584dc84a7b89faa68e3");
    expect(result?.attributes.to).toBe("0x3ee18B2214AFF97000D974cf647E7C347E8fa585");
    expect(result?.attributes.methodsByAddress).toBe("MethodCompleteTransfer");
    expect(result?.attributes.emitterChain).toBe(undefined);
    expect(result?.attributes.emitterAddress).toBe(undefined);
    expect(result?.attributes.sequence).toBe(undefined);
  });

  it("should be able to map log to evmRedeemedTransactionFoundMapper with vaaInformation from the log topics", async () => {
    // When
    const [result] = await handler.handle([
      {
        blockHash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        blockNumber: 0x11ec2bcn,
        chainId: 1,
        from: "0xfb070adcd21361a3946a0584dc84a7b89faa68e3",
        gas: "0x14485",
        gasPrice: "0xfc518561e",
        hash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        input:
          "0xc68785190000000000000000000000000000000000000000000000000000000000000001637651ef71f834be28b8fab1dce9c228c2fe1813831bbc3673cfd3abde6dbb3d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080420000",
        maxFeePerGas: "0x1610f75b9a",
        maxPriorityFeePerGas: "0x5f5e100",
        nonce: "0x1",
        r: "0xf5794b0970386d73b693b17f147fae0427db278e951e45465ac2c9835537e5a9",
        s: "0x6dccc8cfee216bc43a9d66525fa94905da234ad32d6cc3220845bef78f25dd42",
        status: "0x1",
        timestamp: 1702663079,
        to: "0x3ee18b2214aFF97000d974Cf647e7C347e8fa585",
        transactionIndex: "0x6f",
        type: "0x2",
        v: "0x1",
        value: "0x5b09cd3e5e90000",
        environment: "testnet",
        chain: "ethereum",
        logs: [
          {
            address: "0x3ee18B2214AFF97000D974cf647E7C347E8fa585",
            topics: [
              "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e",
              "0x0000000000000000000000000000000000000000000000000000000000000017",
              "0x0000000000000000000000002703483b1a5a7c577e8680de9df8be03c6f30e3c",
              "0x000000000000000000000000000000000000000000000000000000000000250f",
            ],
            data: "0x",
          },
        ],
      },
    ]);

    // Then
    expect(result?.name).toBe("transfer-redeemed");
    expect(result?.chainId).toBe(1);
    expect(result?.txHash).toBe(txHash);
    expect(result?.blockHeight).toBe(18793148n);
    expect(result?.attributes.blockNumber).toBe(18793148n);
    expect(result?.attributes.from).toBe("0xfb070adcd21361a3946a0584dc84a7b89faa68e3");
    expect(result?.attributes.to).toBe("0x3ee18b2214aFF97000d974Cf647e7C347e8fa585");
    expect(result?.attributes.methodsByAddress).toBe("MethodCompleteTransfer");
    expect(result?.attributes.emitterChain).toBe(23);
    expect(result?.attributes.emitterAddress).toBe(
      "0000000000000000000000002703483B1A5A7C577E8680DE9DF8BE03C6F30E3C"
    );
    expect(result?.attributes.sequence).toBe(9487);
  });

  it("should be able to map log to evmRedeemedTransactionFoundMapper with vaaInformation from the log data (e.g. manual NTT)", async () => {
    // When
    const [result] = await handler.handle([
      {
        blockHash: "0x3cc1804f1ffb64a8a62383e20a493fb6a4ca7cfbc17a21382787b228e4906ca0",
        blockNumber: 19957644n,
        chainId: 10003,
        from: "0xe6990c7e206d418d62b9e50c8e61f59dc360183b",
        gas: "0x4f60c",
        gasPrice: "0x5f5e1000",
        hash: "0xcc63ff6948718c386158d8f6a678199575a3707b9d5014de4a984b5897d987f4",
        input:
          "0xf953cec700000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000154010000000001000bfc8dc23fa1933433ae7fcbcb24c0a20672131bbcf46a3b4faeb07072757dec7ef2a1cd1453b7146fe8f73c181e0c91e944d362324c3c4e61271501caaf1ed40165e656f000000000271200000000000000000000000055aaf4d9399c472b252e7c0b49408b5bc7d7328e0000000000000016c89945ff10000000000000000000000000459b4d6df31c1c1f8b6fda0f8ad77e1eff832bcf000000000000000000000000cc1ebd7a6661c0f6e19d2bbdb881b11f3b3f40ff00910000000000000000000000000000000000000000000000000000000000000019000000000000000000000000e6990c7e206d418d62b9e50c8e61f59dc360183b004f994e54540800000000001e8480000000000000000000000000ce0bd78b496bc8ddd25c8a192771e4537f0794c8000000000000000000000000e6990c7e206d418d62b9e50c8e61f59dc360183b27130000000000000000000000000000",
        maxFeePerGas: "0x608f3d00",
        maxPriorityFeePerGas: "0x59682f00",
        nonce: "0x9",
        v: "0x1",
        r: "0x3b3a98aee936de6490a73dcf7653e75619f5189c6ad589286dedabd368a0e71b",
        s: "0x61335d02972125552e7f0c85b3333bae262e9751532fd45654c0dc764015f658",
        status: "0x1",
        timestamp: 1702663079,
        to: "0xC9a478f97ad763052AD4F00c4d7fC5d187DFFb1B",
        transactionIndex: "0x6f",
        type: "0x1",
        value: "0x0",
        environment: "testnet",
        chain: "arbitrum-sepolia",
        logs: [
          {
            address: "0xC9a478f97ad763052AD4F00c4d7fC5d187DFFb1B",
            topics: ["0xf6fc529540981400dc64edf649eb5e2e0eb5812a27f8c81bac2c1d317e71a5f0"],
            data: "0xd6e26148b9040a4a017db65958f5c5d30e6b53cf53a34299d8efa45c42cd2c8e000000000000000000000000000000000000000000000000000000000000271200000000000000000000000055aaf4d9399c472b252e7c0b49408b5bc7d7328e0000000000000000000000000000000000000000000000000000000000000016",
          },
          {
            address: "0xcc1ebd7a6661c0f6e19d2bbdb881b11f3b3f40ff",
            topics: ["0x35a2101eaac94b493e0dfca061f9a7f087913fde8678e7cde0aca9897edba0e5"],
            data: "0x18f7512db4006dc0d2ee49923afaa07dbf530d7feda528969e7253236aa3c66d00000000000000000000000000ac6efc189140b50a043b5e43c108cf571586d10000000000000000000000000000000000000000000000000000000000000000",
          },
          {
            address: "0xcc1ebd7a6661c0f6e19d2bbdb881b11f3b3f40ff",
            topics: [
              "0x504e6efe18ab9eed10dc6501a417f5b12a2f7f2b1593aed9b89f9bce3cf29a91",
              "0x18f7512db4006dc0d2ee49923afaa07dbf530d7feda528969e7253236aa3c66d",
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
            data: "0x00000000000000000000000000000000000000000000000000470de4df820000",
          },
        ],
      },
    ]);

    // Then
    expect(result?.name).toBe("transfer-redeemed");
    expect(result?.chainId).toBe(10003);
    expect(result?.txHash).toBe(
      "0xcc63ff6948718c386158d8f6a678199575a3707b9d5014de4a984b5897d987f4"
    );
    expect(result?.attributes.methodsByAddress).toBe("WormholeTransceiverReceiveMessage");
    expect(result?.attributes.emitterChain).toBe(10002);
    expect(result?.attributes.emitterAddress).toBe(
      "00000000000000000000000055AAF4D9399C472B252E7C0B49408B5BC7D7328E"
    );
    expect(result?.attributes.sequence).toBe(22);
  });

  it("should be able to map log to evmRedeemedTransactionFoundMapper with vaaInformation from the log data (e.g. relayed NTT)", async () => {
    // When
    const [result] = await handler.handle([
      {
        hash: "0xd8ff00d9dc3d9a0fa3d8a1b66ca4a6ff8a39ca940ec13609e66ae6959660765c",
        transactionIndex: "0x1",
        blockHash: "0xcfb3d3bd9ee0f4df69acebc6bc7edffc679452a29459f8bea618da41a2f0e6e3",
        blockNumber: 19955591n,
        from: "0x734d539a7efee15714a2755caa4280e12ef3d7e4",
        to: "0x7b1bd7a6b4e61c2a123ac6bc2cbfc614437d0470",
        logs: [
          {
            address: "0xC9a478f97ad763052AD4F00c4d7fC5d187DFFb1B",
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
        ],
        status: "0x1",
        type: "0x2",
        nonce: "0x162",
        value: "0x21d5dab38e83e0",
        gasPrice: "0x5f8ee40",
        gas: "0x3d0900",
        input:
          "0xa60eb4c8000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000a0000000000000000000000000734d539a7efee15714a2755caa4280e12ef3d7e40000000000000000000000000000000000000000000000000000000000000380000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002a2010000000001000648a2069233ff8ef1a7a07d62ede39bfef7b50cd5d40dfc9c295d42c66f0f096a5c358ad589646b1d61e95f56602ab771c4b6bdcae49890a40d32171b58549d0165e6518c0000000027120000000000000000000000007b1bd7a6b4e61c2a123ac6bc2cbfc614437d047000000000000013ac0f01271300000000000000000000000000ac6efc189140b50a043b5e43c108cf571586d1000000d99945ff10000000000000000000000000459b4d6df31c1c1f8b6fda0f8ad77e1eff832bcf000000000000000000000000cc1ebd7a6661c0f6e19d2bbdb881b11f3b3f40ff00910000000000000000000000000000000000000000000000000000000000000018000000000000000000000000e6990c7e206d418d62b9e50c8e61f59dc360183b004f994e545408000000000016e360000000000000000000000000ce0bd78b496bc8ddd25c8a192771e4537f0794c8000000000000000000000000e6990c7e206d418d62b9e50c8e61f59dc360183b2713000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000600000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007a120000000000000000000000000000000000000000000000000000000046f5399e7271300000000000000000000000000000000000000000000000000000000000000000000000000000000000000007a0a53847776f7e94cc35742971acb2217b0db810000000000000000000000007a0a53847776f7e94cc35742971acb2217b0db8100000000000000000000000055aaf4d9399c472b252e7c0b49408b5bc7d7328e000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
        v: "0x0",
        r: "0xf30b39c8ded6f12c89bc84447fabc57efa33ce14301be27bcc8daba7d1554619",
        s: "0x24f3411550177e821f2f87ef7a07d50b15f04c51b5c977ef207e2639a9296a90",
        maxPriorityFeePerGas: "0x30d40",
        maxFeePerGas: "0x6553f100",
        chainId: 421614,
        timestamp: Date.now(),
        environment: "testnet",
        chain: "arbitrum-sepolia",
      },
    ]);

    // Then
    expect(result?.name).toBe("transfer-redeemed");
    expect(result?.chainId).toBe(421614);
    expect(result?.txHash).toBe(
      "0xd8ff00d9dc3d9a0fa3d8a1b66ca4a6ff8a39ca940ec13609e66ae6959660765c"
    );
    expect(result?.attributes.methodsByAddress).toBe("StandardRelayerDelivery");
    expect(result?.attributes.emitterChain).toBe(10002);
    expect(result?.attributes.emitterAddress).toBe(
      "0000000000000000000000007B1BD7A6B4E61C2A123AC6BC2CBFC614437D0470"
    );
    expect(result?.attributes.sequence).toBe(5036);
  });

  it("should be remove all events because is not possible map protocol values in mapper", async () => {
    // When
    const [result] = await handler.handle([
      {
        blockHash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        blockNumber: 0x11ec2bcn,
        chainId: 1,
        from: "0xfb070adcd21361a3946a0584dc84a7b89faa68e3",
        gas: "0x14485",
        gasPrice: "0xfc518561e",
        hash: "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d",
        input:
          "0xc68785190000000000000000000000000000000000000000000000000000000000000001637651ef71f834be28b8fab1dce9c228c2fe1813831bbc3673cfd3abde6dbb3d00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000080420000",
        maxFeePerGas: "0x1610f75b9a",
        maxPriorityFeePerGas: "0x5f5e100",
        nonce: "0x1",
        r: "0xf5794b0970386d73b693b17f147fae0427db278e951e45465ac2c9835537e5a9",
        s: "0x6dccc8cfee216bc43a9d66525fa94905da234ad32d6cc3220845bef78f25dd42",
        status: "0x1",
        timestamp: 1702663079,
        to: "0x3ee18B2214AFF97000D974cf646E7C347E8fa585",
        transactionIndex: "0x6f",
        type: "0x2",
        v: "0x1",
        value: "0x5b09cd3e5e90000",
        environment: "testnet",
        chain: "ethereum",
        logs: [
          {
            address: "0x3ee18B2214AFF97000D974cf646E7C347E8fa585",
            topics: [
              "0xf02867db6908ee5f81fd178573ae9385837f0a0a72553f8c08306759a7e0f00e",
              "0x0000000000000000000000000000000000000000000000000000000000000017",
              "0x0000000000000000000000002703483b1a5a7c577e8680de9df8be03c6f30e3c",
              "0x000000000000000000000000000000000000000000000000000000000000250f",
            ],
            data: "0x",
          },
        ],
      },
    ]);

    // Then
    expect(result).toBeUndefined;
  });
});
