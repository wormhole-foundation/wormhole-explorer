import { describe, it, expect, afterEach, afterAll } from "@jest/globals";
import { MoonbeamEvmJsonRPCBlockRepository } from "../../../src/infrastructure/repositories";
import { HttpClient } from "../../../src/infrastructure/rpc/http/HttpClient";
import { EvmTag } from "../../../src/domain/entities";
import axios from "axios";
import nock from "nock";

let repo: MoonbeamEvmJsonRPCBlockRepository;

axios.defaults.adapter = "http"; // needed by nock
const moonbeam = "moonbeam";
const rpc = "http://localhost";

describe("MoonbeamEvmJsonRPCBlockRepository", () => {
  afterAll(() => {
    nock.restore();
  });

  afterEach(() => {
    nock.cleanAll();
  });

  it("should be able to get finalized block height", async () => {
    // Given
    const block = 19808090n;

    givenARepo();
    givenBlockHeightIs(block, "latest");
    givenGetBlockIs(block, block);
    givenFinalizedBlock(block);

    // When
    const result = await repo.getBlockHeight(moonbeam, "latest");

    // Then
    expect(result).toBe(block);
  });
});

const givenARepo = () => {
  repo = new MoonbeamEvmJsonRPCBlockRepository(
    {
      chains: {
        moonbeam: { rpcs: [rpc], timeout: 100, name: moonbeam, network: "mainnet", chainId: 16 },
      },
    },
    new HttpClient()
  );
};

const givenBlockHeightIs = (height: bigint, commitment: EvmTag) => {
  nock(rpc)
    .post("/", {
      jsonrpc: "2.0",
      method: "eth_getBlockByNumber",
      params: [commitment, false],
      id: 1,
    })
    .reply(200, {
      jsonrpc: "2.0",
      id: 1,
      result: {
        number: height,
        hash: blockHash(height),
        timestamp: "0x654a892f",
      },
    });
};

const givenGetBlockIs = (height: bigint, commitment: bigint) => {
  const objToString = JSON.parse(
    JSON.stringify({
      jsonrpc: "2.0",
      method: "eth_getBlockByNumber",
      params: [commitment, false],
      id: 1,
    })
  );

  nock(rpc)
    .post("/", objToString)
    .reply(200, {
      jsonrpc: "2.0",
      id: 1,
      result: {
        number: `0x${height.toString(16)}`,
        hash: blockHash(height),
        timestamp: "0x654a892f",
      },
    });
};

const givenFinalizedBlock = (height: bigint) => {
  nock(rpc)
    .post("/", {
      jsonrpc: "2.0",
      method: "moon_isBlockFinalized",
      params: [blockHash(height)],
      id: 1,
    })
    .reply(200, {
      result: true,
    });
};

const blockHash = (blockNumber: bigint) => `0x${blockNumber.toString(16)}`;
