import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { describe, it, expect, afterEach, afterAll } from "@jest/globals";
import { MoonbeamEvmJsonRPCBlockRepository } from "../../../src/infrastructure/repositories";
import { InstrumentedHttpProvider } from "../../../src/infrastructure/rpc/http/InstrumentedHttpProvider";
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
    const hexadecimalblock = "0x12e3f5a";

    givenARepo();
    givenBlockHeightIs(block, "latest");
    givenGetBlockIs(block, hexadecimalblock);
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
      environment: "testnet",
    },
    {
      moonbeam: {
        get: () => new InstrumentedHttpProvider({ url: rpc, chain: "moonbeam" }),
      },
    } as any
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

const givenGetBlockIs = (height: bigint, hexadecimalblock: string) => {
  const objToString = JSON.parse(
    JSON.stringify({
      jsonrpc: "2.0",
      method: "eth_getBlockByNumber",
      params: [hexadecimalblock, false],
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
