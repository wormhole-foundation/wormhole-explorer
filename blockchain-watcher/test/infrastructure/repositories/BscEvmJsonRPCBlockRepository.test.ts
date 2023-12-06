import { describe, it, expect, afterEach, afterAll } from "@jest/globals";
import { BscEvmJsonRPCBlockRepository } from "../../../src/infrastructure/repositories";
import { HttpClient } from "../../../src/infrastructure/rpc/http/HttpClient";
import { EvmTag } from "../../../src/domain/entities/evm";
import axios from "axios";
import nock from "nock";

axios.defaults.adapter = "http"; // needed by nock
const bsc = "bsc";
const rpc = "http://localhost";

let repo: BscEvmJsonRPCBlockRepository;

describe("BscEvmJsonRPCBlockRepository", () => {
  afterAll(() => {
    nock.restore();
  });

  afterEach(() => {
    nock.cleanAll();
  });

  it("should be able to get block height", async () => {
    // Given
    const originalBlock = 1980809n;
    const expectedBlock = 1980794n;

    givenARepo();
    givenBlockHeightIs(originalBlock, "latest");

    // When
    const result = await repo.getBlockHeight(bsc, "latest");

    // Then
    expect(result).toBe(expectedBlock);
  });
});

const givenARepo = () => {
  repo = new BscEvmJsonRPCBlockRepository(
    {
      chains: {
        bsc: { rpcs: [rpc], timeout: 100, name: bsc, network: "mainnet", chainId: 4 },
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
        number: `0x${height.toString(16)}`,
        hash: blockHash(height),
        timestamp: "0x654a892f",
      },
    });
};

const blockHash = (blockNumber: bigint) => `0x${blockNumber.toString(16)}`;
