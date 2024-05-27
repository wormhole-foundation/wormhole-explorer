import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { HandleEvmLogs, HandleEvmLogsConfig } from "../../../../src/domain/actions";
import { EvmLog, LogFoundEvent } from "../../../../src/domain/entities";
import { StatRepository } from "../../../../src/domain/repositories";

let statsRepo: StatRepository;

const ABI =
  "event SendEvent(uint64 indexed sequence, uint256 deliveryQuote, uint256 paymentForExtraReceiverValue)";
const mapper = (log: EvmLog, args: ReadonlyArray<any>) => {
  return {
    name: "send-event",
    address: log.address,
    chainId: 1,
    txHash: "0x0",
    blockHeight: 0n,
    blockTime: 0,
    attributes: {
      sequence: args[0].toString(),
      deliveryQuote: args[1].toString(),
      paymentForExtraReceiverValue: args[2].toString(),
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

let targetRepoSpy: jest.SpiedFunction<(typeof targetRepo)["save"]>;

let evmLogs: EvmLog[];
let cfg: HandleEvmLogsConfig;
let handleEvmLogs: HandleEvmLogs<LogFoundEvent<Record<string, string>>>;

describe("HandleEvmLogs", () => {
  afterEach(async () => {});

  it("should be able to map logs", async () => {
    const expectedLength = 5;
    givenConfig(ABI);
    givenEvmLogs(expectedLength, expectedLength);
    givenStatsRepository();
    givenHandleEvmLogs();

    const result = await handleEvmLogs.handle(evmLogs);

    expect(result).toHaveLength(expectedLength);
    expect(result[0].attributes.sequence).toBe("3389");
    expect(result[0].attributes.deliveryQuote).toBe("75150000000000000");
    expect(result[0].attributes.paymentForExtraReceiverValue).toBe("0");
    expect(targetRepoSpy).toBeCalledWith(result, "ethereum");
  });
});

const givenHandleEvmLogs = (targetFn: "save" | "failingSave" = "save") => {
  targetRepoSpy = jest.spyOn(targetRepo, targetFn);
  handleEvmLogs = new HandleEvmLogs(cfg, mapper, targetRepo[targetFn], statsRepo);
};

const givenConfig = (abi: string) => {
  cfg = {
    filter: {
      addresses: ["0x28D8F1Be96f97C1387e94A53e00eCcFb4E75175a"],
      topics: ["0xda8540426b64ece7b164a9dce95448765f0a7263ef3ff85091c9c7361e485364"],
    },
    metricName: "process_source_ethereum_event",
    abi,
    commitment: "latest",
    chainId: 2,
    chain: "ethereum",
    id: "poll-log-message-published-ethereum",
  };
};

const givenStatsRepository = () => {
  statsRepo = {
    count: () => {},
    measure: () => {},
    report: () => Promise.resolve(""),
  };
};

const givenEvmLogs = (length: number, matchingFilterOnes: number) => {
  evmLogs = [];
  let matchingCount = 0;
  for (let i = 0; i < length; i++) {
    let address = "0x392f472048681816e91026cd768c60958b55352add2837adea9ea6249178b8a8";
    let topic: string | undefined = undefined;
    if (matchingCount < matchingFilterOnes) {
      address = cfg.filter.addresses![0].toUpperCase();
      topic = cfg.filter.topics![0];
      matchingCount++;
    }

    evmLogs.push({
      blockTime: 0,
      blockNumber: BigInt(i + 1),
      blockHash: "0x1a07d0bd31c84f0dab36eac31a2f3aa801852bf8240ffba19113c62463f694fa",
      address: address,
      removed: false,
      data: "0x000000000000000000000000000000000000000000000000010afc86dedee0000000000000000000000000000000000000000000000000000000000000000000",
      transactionHash: "0x2077dbd0c685c264dfa4e8e048ff15b03775043070216644258bf1bd3e419aa8",
      transactionIndex: "0x4",
      topics: topic
        ? [topic, "0x0000000000000000000000000000000000000000000000000000000000000d3d"]
        : [],
      logIndex: 0,
      chainId: 2,
      chain: "ethereum",
    });
  }
};
