import { afterAll, afterEach, describe, expect, it } from "@jest/globals";
import nock from "nock";
import { SuiJsonRPCBlockRepository } from "../../../src/infrastructure/repositories";
import { getFullnodeUrl } from "@mysten/sui.js/client";

const rpc = "http://localhost";
let repo: SuiJsonRPCBlockRepository;

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
});

const givenARepo = () => {
  repo = new SuiJsonRPCBlockRepository({ rpc });
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
