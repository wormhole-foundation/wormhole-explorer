import { afterAll, afterEach, describe, expect, it, jest } from "@jest/globals";
import nock from "nock";
import { SuiJsonRPCBlockRepository } from "../../../src/infrastructure/repositories";
import { SuiClient, getFullnodeUrl } from "@mysten/sui.js/client";
import { count } from "console";
import { randomBytes } from "crypto";
import base58 from "bs58";

const rpc = "http://localhost";
let repo: SuiJsonRPCBlockRepository;

let txs: string[];
let checkpoints: string[];
let getTxsSpy: jest.SpiedFunction<SuiClient['multiGetTransactionBlocks']>;
let getCheckpointsSpy: jest.SpiedFunction<SuiClient['getCheckpoints']>;

const TX_BATCH_SIZE = 50;
const CHK_BATCH_SIZE = 100;

describe("SuiJsonRPCBlockRepository", () => {
  afterAll(() => {
    nock.restore();
  });

  afterEach(() => {
    nock.cleanAll();
  });

  it("should be able to get the latest checkpoint sequence", async () => {
    const expectedSeq = 23993824n;
    givenARepo();
    givenLastCheckpointIs(expectedSeq);

    const result = await repo.getLastCheckpoint();

    expect(result).toBe(expectedSeq);
  });

  it("should be able to fetch a set of transaction blocks", async () => {
    givenARepo();
    givenTransactions(2);

    const result = await repo.getTransactionBlockReceipts(txs);

    for (let i = 0; i < result.length; i++) {
      expect(result[i].digest).toBe(txs[i]);
      expect(result[i].timestampMs).toBe('1706107525474');
      expect(result[i].checkpoint).toBe('24408383');
      expect(result[i].events).toHaveLength(1);
      expect(result[i].errors).toHaveLength(0);
      expect(result[i].transaction).not.toBeFalsy();
    }
  });

  it("should fetch transactions above the rpc limit", async () => {
    givenARepo();
    givenTransactions(123);

    const result = await repo.getTransactionBlockReceipts(txs);

    expect(result.length).toBe(123)
    expect(getTxsSpy).toHaveBeenCalledTimes(3);
  });

  it("should fetch a range of checkpoints", async () => {
    const range = { from: 200000n, to: 200049n };
    givenARepo();
    givenCheckpoints(range.from, range.to);

    const result = await repo.getCheckpoints(range);

    expect(result).toHaveLength(50);
    expect(getCheckpointsSpy).toHaveBeenCalledTimes(1);
  });

  it("should fetch cehckpoints above the rpc limit", async () => {
    const range = { from: 200000n, to: 200249n };
    givenARepo();
    givenCheckpoints(range.from, range.to);

    const result = await repo.getCheckpoints(range);

    expect(result).toHaveLength(250);
    expect(getCheckpointsSpy).toHaveBeenCalledTimes(3);
  });
});

const givenARepo = () => {
  repo = new SuiJsonRPCBlockRepository({ rpc });

  getTxsSpy = jest.spyOn((repo as any).client as SuiClient, 'multiGetTransactionBlocks');
  getCheckpointsSpy = jest.spyOn((repo as any).client as SuiClient, 'getCheckpoints');
};

const givenLastCheckpointIs = (sequence: bigint) => {
  nock(rpc)
    .post("/", (body) => {
      // SuiClient inserts a GUID
      return body.method === "sui_getLatestCheckpointSequenceNumber" && body.params.length === 0;
    })
    .reply(200, {
      jsonrpc: "2.0",
      id: 1,
      result: sequence.toString(),
    });
};

const givenTransactions = (count: number) => {
  txs = [...new Array(count).keys()].map(_ => randomDigest());

  for (let i = 0; i < count / TX_BATCH_SIZE; i++) {
    nock(rpc)
      .post("/", (body) => {
        // SuiClient inserts a GUID
        return body.method === "sui_multiGetTransactionBlocks" && body.params.length === 2;
      })
      .reply(200, {
        jsonrpc: "2.0",
        id: 1,
        result: txs.slice(i * TX_BATCH_SIZE, (i+1) * TX_BATCH_SIZE).map(mapTx),
      });
  }
};

const givenCheckpoints = (from: bigint, to: bigint) => {
  checkpoints = [...new Array(Number(to - from + 1n)).keys()].map(t => (from + BigInt(t)).toString());

  for (let i = 0; i < checkpoints.length / CHK_BATCH_SIZE; i++) {
    nock(rpc)
      .post("/", (body) => {
        // SuiClient inserts a GUID
        return body.method === "sui_getCheckpoints";
      })
      .reply(200, {
        jsonrpc: "2.0",
        id: 1,
        result: {
          data: checkpoints.slice(i * CHK_BATCH_SIZE, (i+1) * CHK_BATCH_SIZE).map(mapCheckpoint),
          cursor: '',
          hasNext: false,
        },
      });
  }
};

const mapTx = (digest: string) => (
  {
    digest: digest,
    transaction: {
      data: {
        messageVersion: "v1",
        transaction: {
          kind: "ProgrammableTransaction",
          inputs: [],
          transactions: [],
        },
        sender: "0xfcda48b391b8a1c6a9e57f30247bc0d5a97595f4a61784078ad17e11c2a8d529",
        gasData: {
          payment: [
            {
              objectId: "0xf0405f2f6e2ef6f86762f973907faeb68e41f8a3fb00326bb58dea73701209f7",
              version: "63608463",
              digest: "CoLRg7oKMJ3T3p3X9yTWGpiNPJL3SvdkQJLbXetfVVYe",
            },
          ],
          owner: "0xfcda48b391b8a1c6a9e57f30247bc0d5a97595f4a61784078ad17e11c2a8d529",
          price: "750",
          budget: "7690416",
        },
      },
      txSignatures: [
        "AFEXiY7OGpEFtlU4YZ/K6IktslWzqRfSqP8e90V+Pqowv8emUBDg875hoDxLU+hRqPPShBDqfTy6IzOZeNSQpQrPyLwQgQtt1dF1BjINPA76mVZwoYzzB1KhblrBFvgl/Q==",
      ],
    },
    events: [{}],
    timestampMs: "1706107525474",
    checkpoint: "24408383",
  }
)

const randomDigest = () => base58.encode(randomBytes(32));

const mapCheckpoint = (height: string) => ({
  epoch: '285',
  sequenceNumber: height,
  digest: randomDigest(),
  networkTotalTransactions: '1063849022',
  previousDigest: randomDigest(),
  epochRollingGasCostSummary: {
    computationCost: '169602918500',
    storageCost: '1234564033600',
    storageRebate: '1198647002916',
    nonRefundableStorageFee: '12107545484'
  },
  timestampMs: '1705946928610',
  transactions: [],
  checkpointCommitments: [],
  validatorSignature: 'i9kUhaIs2SOZFMqYgAOf9rjZJs8lVWGUlxJwsiGcwW8Eg7u4EfWGHjJulg6ZwMWb'
})
