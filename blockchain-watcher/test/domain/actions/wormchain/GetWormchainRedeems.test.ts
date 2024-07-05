import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { thenWaitForAssertion } from "../../../wait-assertion";
import { WormchainTransaction } from "../../../../src/domain/entities/wormchain";
import {
  PollWormchainLogsMetadata,
  PollWormchainLogsConfig,
  PollWormchain,
} from "../../../../src/domain/actions/wormchain/PollWormchain";
import {
  WormchainRepository,
  MetadataRepository,
  StatRepository,
} from "../../../../src/domain/repositories";

let getBlockHeightSpy: jest.SpiedFunction<WormchainRepository["getBlockHeight"]>;
let getBlockLogsSpy: jest.SpiedFunction<WormchainRepository["getBlockLogs"]>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollWormchainLogsMetadata>["save"]>;

let handlerSpy: jest.SpiedFunction<(txs: WormchainTransaction[]) => Promise<void>>;

let metadataRepo: MetadataRepository<PollWormchainLogsMetadata>;
let wormchainRepo: WormchainRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (txs: WormchainTransaction[]) => Promise.resolve(),
  failing: (txs: WormchainTransaction[]) => Promise.reject(),
};
let pollWormchain: PollWormchain;

let props = {
  blockBatchSize: 100,
  from: 0n,
  limit: 0n,
  environment: "testnet",
  commitment: "immediate",
  addresses: ["wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j"],
  interval: 5000,
  topics: [],
  chainId: 3104,
  filter: {
    address: "wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j",
  },
  chain: "wormchain",
  id: "poll-log-message-published-wormchain",
};

let cfg = new PollWormchainLogsConfig(props);

describe("GetWormchainRedeems", () => {
  afterEach(async () => {
    await pollWormchain.stop();
  });

  it("should be skip the transations blocks, because the transactions will be undefined", async () => {
    // Given
    givenWormchainBlockRepository(8418529n);
    givenMetadataRepository({ lastBlock: 8418528n });
    givenStatsRepository();
    givenPollWormchainTx(cfg);

    // When
    await whenPollWormchainLogsStarts();

    // Then
    await thenWaitForAssertion(() =>
      expect(getBlockLogsSpy).toBeCalledWith(3104, 8418529n, ["wasm", "send_packet"])
    );
  });

  it("should be process the log because it contains wasm and send_packet transactions", async () => {
    // Given
    const log = {
      transactions: [
        {
          hash: "0xd84a9c85170c28b12a1436082e99c1ea2598cbf36f9e263bfc0b7fb79a972dfe",
          type: "wasm",
          attributes: [
            {
              key: "X2NvbnRyYWN0X2FkZHJlc3M=",
              value:
                "d29ybWhvbGUxNGhqMnRhdnE4ZnBlc2R3eHhjdTQ0cnR5M2hoOTB2aHVqcnZjbXN0bDR6cjN0eG1mdnc5c3JyZzQ2NQ==",
              index: true,
            },
            { key: "YWN0aW9u", value: "c3VibWl0X29ic2VydmF0aW9ucw==", index: true },
            {
              key: "b3duZXI=",
              value: "d29ybWhvbGUxcWdqZmRmNWczMnN4d2pqanduNHZwOGZnZzJmdzBtOHh4aDdheGM=",
              index: true,
            },
          ],
        },
        {
          hash: "0x9042d7f656f2292e8a4bfa9468ee8215fd6de9ff23b447e20f96f6a70559df68",
          type: "wasm",
          attributes: [
            {
              key: "X2NvbnRyYWN0X2FkZHJlc3M=",
              value:
                "d29ybWhvbGUxNGhqMnRhdnE4ZnBlc2R3eHhjdTQ0cnR5M2hoOTB2aHVqcnZjbXN0bDR6cjN0eG1mdnc5c3JyZzQ2NQ==",
              index: true,
            },
            { key: "YWN0aW9u", value: "c3VibWl0X29ic2VydmF0aW9ucw==", index: true },
            {
              key: "b3duZXI=",
              value: "d29ybWhvbGUxNW5rbTdhdnB4eHNuY3I0Z2c4dTJxbDdnY2tsbTJrcmt6d2U3N20=",
              index: true,
            },
          ],
        },
        {
          type: "send_packet",
          attributes: [
            {
              key: "packet_data",
              value:
                '{"amount":"200000000","denom":"factory/wormhole14ejqjyq8um4p3xfqj74yld5waqljf88fz25yxnma0cngspxe3les00fpjx/8sYgCzLRJC3J7qPn2bNbx6PiGcarhyx8rBhVaNnfvHCA","receiver":"osmo1r6f5tfxdx2pw5p94f2v5n96xd4nglz5q30mczj","sender":"wormhole14ejqjyq8um4p3xfqj74yld5waqljf88fz25yxnma0cngspxe3les00fpjx"}',
            },
            {
              key: "packet_data_hex",
              value:
                "7b22616d6f756e74223a22323030303030303030222c2264656e6f6d223a22666163746f72792f776f726d686f6c653134656a716a797138756d3470337866716a3734796c64357761716c6a663838667a323579786e6d6130636e6773707865336c6573303066706a782f38735967437a4c524a43334a3771506e32624e6278365069476361726879783872426856614e6e6676484341222c227265636569766572223a226f736d6f3172366635746678647832707735703934663276356e39367864346e676c7a357133306d637a6a222c2273656e646572223a22776f726d686f6c653134656a716a797138756d3470337866716a3734796c64357761716c6a663838667a323579786e6d6130636e6773707865336c6573303066706a78227d",
            },
            {
              key: "packet_timeout_height",
              value: "0-0",
            },
            {
              key: "packet_timeout_timestamp",
              value: "1716857658737721878",
            },
            {
              key: "packet_sequence",
              value: "37063",
            },
            {
              key: "packet_src_port",
              value: "transfer",
            },
            {
              key: "packet_src_channel",
              value: "channel-3",
            },
            {
              key: "packet_dst_port",
              value: "transfer",
            },
            {
              key: "packet_dst_channel",
              value: "channel-2186",
            },
            {
              key: "packet_channel_ordering",
              value: "ORDER_UNORDERED",
            },
            {
              key: "packet_connection",
              value: "connection-4",
            },
          ],
        },
      ],
      blockHeight: "7606615",
      timestamp: 1711025902418,
    };

    givenWormchainBlockRepository(7606615n, log);
    givenMetadataRepository({ lastBlock: 7606614n });
    givenStatsRepository();
    givenPollWormchainTx(cfg);

    // When
    await whenPollWormchainLogsStarts();

    // Then
    await thenWaitForAssertion(() =>
      expect(getBlockLogsSpy).toBeCalledWith(3104, 7606615n, ["wasm", "send_packet"])
    );
  });
});

const givenWormchainBlockRepository = (blockHeigh: bigint, log: any = {}) => {
  wormchainRepo = {
    getBlockHeight: () => Promise.resolve(blockHeigh),
    getBlockLogs: () => Promise.resolve(log),
    getRedeems: () => Promise.resolve([]),
  };

  getBlockHeightSpy = jest.spyOn(wormchainRepo, "getBlockHeight");
  getBlockLogsSpy = jest.spyOn(wormchainRepo, "getBlockLogs");
};

const givenMetadataRepository = (data?: PollWormchainLogsMetadata) => {
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

const givenPollWormchainTx = (cfg: PollWormchainLogsConfig) => {
  pollWormchain = new PollWormchain(
    wormchainRepo,
    metadataRepo,
    statsRepo,
    cfg,
    "GetWormchainRedeems"
  );
};

const whenPollWormchainLogsStarts = async () => {
  pollWormchain.run([handlers.working]);
};
