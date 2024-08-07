import { evmRedeemedTransactionFoundMapper } from "../../../../src/infrastructure/mappers/evm/evmRedeemedTransactionFoundMapper";
import { HandleEvmTransactions } from "../../../../src/domain/actions";
import { describe, it, expect } from "@jest/globals";

const txHash = "0x612a35f6739f70a81dfc34448c68e99dbcfe8dafaf241edbaa204cf0e236494d";

let statsRepo = {
  count: () => {},
  measure: () => {},
  report: () => Promise.resolve(""),
};

const handler = new HandleEvmTransactions(
  {
    abi: "event Delivery(address indexed recipientContract, uint16 indexed sourceChain, uint64 indexed sequence, bytes32 deliveryVaaHash, uint8 status, uint256 gasUsed, uint8 refundStatus, bytes additionalStatusInfo, bytes overridesInfo)",
    environment: "testnet",
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
        gasUsed: "0x6efa0",
        effectiveGasPrice: "0x2fb1471cd",
      },
    ]);

    // Then
    expect(result).toBeUndefined;
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
        gasUsed: "0x6efa0",
        effectiveGasPrice: "0x2fb1471cd",
      },
    ]);

    // Then
    expect(result?.name).toBe("transfer-redeemed");
    expect(result?.chainId).toBe(1);
    expect(result?.txHash).toBe(txHash.substring(2)); // Remove 0x
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
        to: "0x0E24D17D7467467b39Bf64A9DFf88776Bd6c74d7",
        transactionIndex: "0x6f",
        type: "0x1",
        value: "0x0",
        environment: "testnet",
        chain: "arbitrum-sepolia",
        logs: [
          {
            address: "0x0E24D17D7467467b39Bf64A9DFf88776Bd6c74d7",
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
        gasUsed: "0x6efa0",
        effectiveGasPrice: "0x2fb1471cd",
      },
    ]);

    // Then
    expect(result?.name).toBe("transfer-redeemed");
    expect(result?.chainId).toBe(10003);
    expect(result?.txHash).toBe("cc63ff6948718c386158d8f6a678199575a3707b9d5014de4a984b5897d987f4");
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
            address: "0x0E24D17D7467467b39Bf64A9DFf88776Bd6c74d7",
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
        gasUsed: "0x6efa0",
        effectiveGasPrice: "0x2fb1471cd",
      },
    ]);

    // Then
    expect(result?.name).toBe("transfer-redeemed");
    expect(result?.chainId).toBe(421614);
    expect(result?.txHash).toBe("d8ff00d9dc3d9a0fa3d8a1b66ca4a6ff8a39ca940ec13609e66ae6959660765c");
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
        gasUsed: "0x6efa0",
        effectiveGasPrice: "0x2fb1471cd",
      },
    ]);

    // Then
    expect(result).toBeUndefined;
  });

  it("should be able to map log to evmRedeemedTransactionFoundMapper with vaaInformation from the log topics (e.g. PORTAL TOKEN BRIDGE)", async () => {
    // When
    const [result] = await handler.handle([
      {
        blockHash: "0xfe78ab4e96e70b70fa773283ada9e582a5506372757d570e8ac624ea7d23f602",
        blockNumber: 0x5584a3n,
        from: "0x6d225d88426737dbd56bbb959954cb787b5b63fe",
        gas: "0x2a0f1",
        gasPrice: "0x3b9acaed",
        maxPriorityFeePerGas: "0x3b9ac9ee",
        maxFeePerGas: "0x3b9acb26",
        hash: "0x350f1c1cd25ad3dffe6457ebec8432b861dd7e7884567ca3008ff28ab442cef7",
        input:
          "0xc687851900000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000100010000000001006578c0722c90b0bc493e6b743faf528a2a6a88ddf2604f913e36d92863953b7f742d639f3f96341a7e98ce08f9cc63a8aaa8cbd2f923153f8888a1645577276600660a5b17000103e200013b26409f8aaded3f5ddca184695aa6a0fa829b0c85caf84856324896d214ca980000000000006dc9200100000000000000000000000000000000000000000000000000000004b2cb597dd1507814aad5e65844574a570d123fe7d6eefadce5907471023f9e69d48a064e00010000000000000000000000006d225d88426737dbd56bbb959954cb787b5b63fe27120000000000000000000000000000000000000000000000000000000000000000",
        nonce: "0x12",
        to: "0xdb5492265f6038831e89f495670ff909ade94bd9",
        transactionIndex: "0x4d",
        value: "0x0",
        type: "0x2",
        chainId: 10002,
        v: "0x1",
        r: "0x6135b27ac924f0496534e12ecc904b3f26b2149505930bb9375be59c5de31b01",
        s: "0x2fa69d91ce7d738639b57e8ee7c2de77fc4dcb1921e66ddce811bcb9d1811fa8",
        status: "0x1",
        timestamp: 1711954956,
        environment: "testnet",
        chain: "ethereum-sepolia",
        logs: [
          {
            address: "0x6a90bff9a9fee43c3ed12869e0cfe4f6c8e000e7",
            topics: [
              "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
              "0x0000000000000000000000000000000000000000000000000000000000000000",
              "0x0000000000000000000000006d225d88426737dbd56bbb959954cb787b5b63fe",
            ],
            data: "0x00000000000000000000000000000000000000000000000000000004b2cb597d",
          },
        ],
        gasUsed: "0x6efa0",
        effectiveGasPrice: "0x2fb1471cd",
      },
    ]);

    // Then
    expect(result?.name).toBe("transfer-redeemed");
    expect(result?.chainId).toBe(10002);
    expect(result?.txHash).toBe("350f1c1cd25ad3dffe6457ebec8432b861dd7e7884567ca3008ff28ab442cef7"); // Remove 0x
    expect(result?.blockHeight).toBe(5604515n);
    expect(result?.attributes.blockNumber).toBe(5604515n);
    expect(result?.attributes.from).toBe("0x6d225d88426737dbd56bbb959954cb787b5b63fe");
    expect(result?.attributes.to).toBe("0xdb5492265f6038831e89f495670ff909ade94bd9");
    expect(result?.attributes.methodsByAddress).toBe("MethodCompleteTransfer");
    expect(result?.attributes.emitterChain).toBe(1);
    expect(result?.attributes.emitterAddress).toBe(
      "3b26409f8aaded3f5ddca184695aa6a0fa829b0c85caf84856324896d214ca98"
    );
    expect(result?.attributes.sequence).toBe(28105);
  });

  it("should be able to map log to evmRedeemedTransactionFoundMapper with vaaInformation from the log topics (e.g NTT for W token)", async () => {
    // When
    const [result] = await handler.handle([
      {
        blockHash: "0x2ed1e4699d1db88d7967a06a30e87783bfa2cd4cd3b6d452c4cbe8c125a021a6",
        blockNumber: 0xcc12d6n,
        from: "0x44a56b20e2f60f89a5711b819db4d866574bf010",
        gas: "0x659ef",
        gasPrice: "0x3747d0e",
        maxFeePerGas: "0xc09acfd8",
        maxPriorityFeePerGas: "0x30d40",
        hash: "0xf3e0110825f056831129d6b07430fd87c491999ee033617ebf5ee8d3da1fffd4",
        input:
          "0xf953cec70000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000046c01000000040d00289fec07a6f7232c5252cd54f0700939945a2c7990b26772dfdfb9b6658048ec53daeceb79fd6f8391ed69ed9ddecf55be0c52b4e425161627e73f06f35f94400101e1d2bd4e217e8d7adb4ca65f397f91ad8cedc3b4d7b81aaca4dcca97866072f45bb9f982d4fb5c54c5c714b213a75b76de79557f95dcef00b01d7d935e837b3e00049040420290c5ee47a29846661086e890dd8f191c7faf8c3085219f7e572b9edd003975d584b3e23654354e27b3a09bf18043298601afb9b588556daf25a363da0006dd6640d85edcb0330e1cc320c3528c0f5caa07f0f65f995874ccdad357a1c18343f59b8ab5f3efaf68cd30863c0b82b6b1d882c3a25d8c2d1be63635c8de89080007042035c81916e803ebe2d5acc02e3dabacbb34f407b90f31a075a9a34f06b063227d35e3c683164b900bcc9fc590c4d3e515dbb365c946ee251289b022a8ed1a00089b3e2c3bbaf46318d38fe0c501bdb325b4b605c00a2ace9b70f27077b221274d53874211dfa14a2a3987c260a5653b5ada9bb8df2f6320f310422d0c837c5f55000a7f57084cb5d1dde75c152403f88b3b864cc4f1b3efd7de819de90034e59204e62c6cfd2664b4ad0600f5085abb79c317081ddd83b530058ca91a49f611a1ebf4010b5c1dc054fbb5c11ed8df2aa5ba401692a0c10a861450a44d3c7a3bad84251031078f978b0f7ff9fa358564872b8d1259363edd7a93d9d790edab75e193bd5954000dac4b5201aee6055b49cf67e94414f01c75175fd7e3a8c236c5c432eee9e87c810764ec5aee22d6e4ac9e18d9969f0c1d11c9c09f60187b086082af38617725e7010f60850ce65a990fd64e26eecc7523d1e3fb5962b72bc7ce69058ed0b8429416050babfdae411cfe7eab7c76cf100b85d295f824ed40c40624dad7fc0cd1f5bf0300102bbde4c23983ed1dc30c832bc86a71d4a286f2769c3cd2edb40b7e647255c1b52b66c1aee289b9c1d15ace4db2e42037fd3fd67d48a76337bf73c4dca4a182ab011102ef207e5fd895b82e78339e5d810b279d6cb5196f87921c79732f6ac23f33b5528b2449f8d740fb10cc1c0b555cc4950e1f4e070098a2f8518f06122a4020f40012aa6bb3a3722c8ed10aedf00cc6aaa1d006f156c0882fc9ecb071d9d1ef70b66050755b634b8d86f8505ff2eab0dd0466548d189848cdfb15b5928a34d73c58420166228271000000000001cf5f3614e2cd9b374558f35c7618b25f0d306d5e749b7d29cc030a1a15686238000000000000001c209945ff10057f97be1c39478e57974f6cc9dbfbeebb0e5ce340c2efd52b8295e889a9ede40000000000000000000000005333d0aca64a450add6fef76d6d1375f726cb4840091f6584bf5ce12459598bbaf47ec38d42deec0d8234c826ea6940eb0e87038985767947ef13a158cb9bfcabea018b3f8d2e55b2281a76362624273971dbafa1e99004f994e54540600000000000027106927fdc01ea906f96d7137874cdd7adad00ca35764619310e54196c781d84d5b00000000000000000000000049887a216375fded17dc1aaad4920c3777265614001e00000000000000000000000000000000000000000000",
        nonce: "0x2",
        to: "0xd1a8ab69e00266e8b791a15bc47514153a5045a6",
        transactionIndex: "0x26",
        value: "0x0",
        type: "0x2",
        chainId: 30,
        v: "0x0",
        r: "0x6b4fbf9838f9a9bfe175b30245221558a9935b7c817959e780d1b6a673e2fdc7",
        s: "0x4f9c5abeb848486244df757ba16c2d2436225b2bf2c8adff56d4ecedb51b15b3",
        status: "0x1",
        timestamp: 1713537679,
        environment: "mainnet",
        chain: "base",
        logs: [
          {
            address: "0xd1a8ab69e00266e8b791a15bc47514153a5045a6",
            data: "0xf41341582ab18d4be58d1f914ac65a8e5b6932a41db1aad46a0bf141051065ec0000000000000000000000000000000000000000000000000000000000000001cf5f3614e2cd9b374558f35c7618b25f0d306d5e749b7d29cc030a1a15686238000000000000000000000000000000000000000000000000000000000000001c",
            topics: ["0xf6fc529540981400dc64edf649eb5e2e0eb5812a27f8c81bac2c1d317e71a5f0"],
          },
          {
            address: "0x5333d0aca64a450add6fef76d6d1375f726cb484",
            data: "0x1f715f6fc356db0fd7dfefec21c8023e06026036b304a7315e10b91dafc49990000000000000000000000000d1a8ab69e00266e8b791a15bc47514153a5045a60000000000000000000000000000000000000000000000000000000000000000",
            topics: ["0x35a2101eaac94b493e0dfca061f9a7f087913fde8678e7cde0aca9897edba0e5"],
          },
          {
            address: "0x5333d0aca64a450add6fef76d6d1375f726cb484",
            data: "0x",
            topics: [
              "0x504e6efe18ab9eed10dc6501a417f5b12a2f7f2b1593aed9b89f9bce3cf29a91",
              "0x1f715f6fc356db0fd7dfefec21c8023e06026036b304a7315e10b91dafc49990",
            ],
          },
          {
            address: "0xb0ffa8000886e57f86dd5264b9582b2ad87b2b91",
            data: "0x000000000000000000000000000000000000000000000000002386f26fc10000",
            topics: [
              "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
              "0x0000000000000000000000000000000000000000000000000000000000000000",
              "0x00000000000000000000000049887a216375fded17dc1aaad4920c3777265614",
            ],
          },
          {
            address: "0xb0ffa8000886e57f86dd5264b9582b2ad87b2b91",
            data: "0x0000000000000000000000000000000000000000000000000031bced02db0000000000000000000000000000000000000000000000000000005543df729c0000",
            topics: [
              "0xdec2bacdd2f05b59de34da9b523dff8be42e5e38e818c82fdb0bae774387a724",
              "0x00000000000000000000000049887a216375fded17dc1aaad4920c3777265614",
            ],
          },
        ],
        gasUsed: "0x6efa0",
        effectiveGasPrice: "0x2fb1471cd",
      },
    ]);

    // Then
    expect(result?.name).toBe("transfer-redeemed");
    expect(result?.chainId).toBe(30);
    expect(result?.txHash).toBe("f3e0110825f056831129d6b07430fd87c491999ee033617ebf5ee8d3da1fffd4"); // Remove 0x
    expect(result?.blockHeight).toBe(13374166n);
    expect(result?.attributes.blockNumber).toBe(13374166n);
    expect(result?.attributes.from).toBe("0x44a56b20e2f60f89a5711b819db4d866574bf010");
    expect(result?.attributes.to).toBe("0xd1a8ab69e00266e8b791a15bc47514153a5045a6");
    expect(result?.attributes.methodsByAddress).toBe("WormholeTransceiverReceiveMessage");
    expect(result?.attributes.emitterChain).toBe(1);
    expect(result?.attributes.emitterAddress).toBe(
      "CF5F3614E2CD9B374558F35C7618B25F0D306D5E749B7D29CC030A1A15686238"
    );
    expect(result?.attributes.sequence).toBe(28);
  });
});
