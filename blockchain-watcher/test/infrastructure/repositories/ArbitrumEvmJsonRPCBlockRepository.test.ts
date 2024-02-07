import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { describe, it, expect, afterEach, afterAll, jest } from "@jest/globals";
import { ArbitrumEvmJsonRPCBlockRepository } from "../../../src/infrastructure/repositories";
import { MetadataRepository } from "../../../src/domain/repositories";
import { InstrumentedHttpProvider } from "../../../src/infrastructure/rpc/http/InstrumentedHttpProvider";
import { EvmTag } from "../../../src/domain/entities/evm";
import axios from "axios";
import nock from "nock";
import fs from "fs";
import { FirstProviderPool } from "@xlabs/rpc-pool";

const dirPath = "./metadata-repo";
axios.defaults.adapter = "http"; // needed by nock
const ethereum = "ethereum";
const arbitrum = "arbitrum";
const rpc = "http://localhost";

let repo: ArbitrumEvmJsonRPCBlockRepository;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PersistedBlock[]>["save"]>;
let metadataGetSpy: jest.SpiedFunction<MetadataRepository<PersistedBlock[]>["get"]>;
let metadataRepo: MetadataRepository<PersistedBlock[]>;

describe("ArbitrumEvmJsonRPCBlockRepository", () => {
  afterAll(() => {
    nock.restore();
    if (!fs.existsSync(dirPath)) {
      fs.mkdirSync(dirPath);
    }
  });

  afterEach(() => {
    nock.cleanAll();
    fs.rm(dirPath, () => {});
  });

  it("should be able to get block height with arbitrum latest commitment and eth finalized commitment", async () => {
    // Given
    const originalBlock = 19808090n;
    const expectedBlock = 157542621n;

    givenARepo();
    givenL2Block("latest");
    givenBlockHeightIs(originalBlock, "finalized");

    // When
    const result = await repo.getBlockHeight(arbitrum, "latest");

    // Then
    expect(result).toBe(expectedBlock);
  });

  it("should be throw error because unable to parse empty result for latest block", async () => {
    // Given
    givenARepo();
    nock(rpc)
      .post("/", {
        jsonrpc: "2.0",
        method: "eth_getBlockByNumber",
        params: ["latest", false],
        id: 1,
      })
      .reply(200, {
        jsonrpc: "2.0",
        id: 1,
        result: {},
      });

    try {
      // When
      await repo.getBlockHeight(arbitrum, "latest");
    } catch (e: Error | any) {
      // Then
      expect(e).toBeInstanceOf(Error);
    }
  });
});

const givenARepo = () => {
  repo = new ArbitrumEvmJsonRPCBlockRepository(
    {
      chains: {
        ethereum: { rpcs: [rpc], timeout: 100, name: ethereum, network: "mainnet", chainId: 2 },
        arbitrum: {
          rpcs: [rpc],
          timeout: 100,
          name: arbitrum,
          network: "mainnet",
          chainId: 23,
        },
      },
    },
    {
      ethereum: { get: () => new InstrumentedHttpProvider({ url: rpc, chain: "ethereum" }) },
      arbitrum: { get: () => new InstrumentedHttpProvider({ url: rpc, chain: "arbitrum" }) },
    } as any,
    givenMetadataRepository([{ associatedL1Block: 18764852, l2BlockNumber: 157542621 }])
  );
};

const givenL2Block = (commitment: EvmTag) => {
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
        baseFeePerGas: "0x5f5e100",
        difficulty: "0x1",
        extraData: "0xe24ff00c699874b42c1eb6a325ae6c672c502c529de9dbb24ed9ff51563ab5ec",
        gasLimit: "0x4000000000000",
        gasUsed: "0x44a679",
        hash: "0xff1461564bad17aa8b047fc819f2d167b48e172f9143725815b2ec57b5aef429",
        l1BlockNumber: "0x11dcb25",
        logsBloom:
          "0x00000002000000100000000000020000800000000002000040000000004100000000000000000000000000000000000000005000020020000000000080200000000000010000004980404008000000000202000000000000000000000000004000000020000400000000100400000000040000100000000240000010000800000000000000080000000000000000000400000000000000080000000000000110030000000202000000000400800000000000000000800001000200000080400000000002001000000080001020040000000000000000000080000000000000008010000000000000000000440000000000004000000000000000000000000000",
        miner: "0xa4b000000000000000000073657175656e636572",
        mixHash: "0x00000000000184c900000000011dcb25000000000000000a0000000000000000",
        nonce: "0x000000000012b586",
        number: "0x963e8dd",
        parentHash: "0x648eff1f5f81b60f3bd9031114d54af71a1787f3f73e62fdbcf90cc228e976e9",
        receiptsRoot: "0x25a3340b2150c6db9bab6f82b8885d7f08a0ddbf06b4a5ab765c882a869fc594",
        sendCount: "0x184c9",
        sendRoot: "0xe24ff00c699874b42c1eb6a325ae6c672c502c529de9dbb24ed9ff51563ab5ec",
        sha3Uncles: "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
        size: "0x51f",
        stateRoot: "0x3ab24e412a3954b82927e92294e42ad73051ecf95e561a57c5bf6fcbb8114cc2",
        timestamp: "0x6570deb5",
        totalDifficulty: "0x8110b95",
        transactions: [
          "0xbfd339af3216f5a029e3cb96b75f0a09ef81ad49458af64c6f621d94476de2da",
          "0x3fb701d004eae54917c73e2008ca69e96faa46e8bc9631923170a3c88b19150f",
          "0x538622dae61bdfd20ddad66a3c06e49c94cba6bdca848a9320b9486b5dc1c8ad",
        ],
        transactionsRoot: "0xc9f0d07e2b5d4f0d33e28917fc9e8af0ad5e0558c7d885fd6f48034bccafff06",
      },
    });
};

const givenMetadataRepository = (data: PersistedBlock[]) => {
  metadataRepo = {
    get: () => Promise.resolve(data),
    save: () => Promise.resolve(),
  };
  metadataGetSpy = jest.spyOn(metadataRepo, "get");
  metadataSaveSpy = jest.spyOn(metadataRepo, "save");
  return metadataRepo;
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

type PersistedBlock = {
  associatedL1Block: number;
  l2BlockNumber: number;
};

const blockHash = (blockNumber: bigint) => `0x${blockNumber.toString(16)}`;
