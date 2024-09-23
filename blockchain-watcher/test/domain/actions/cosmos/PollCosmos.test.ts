import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { thenWaitForAssertion } from "../../../waitAssertion";
import { CosmosTransaction } from "../../../../src/domain/entities/cosmos";
import {
  MetadataRepository,
  CosmosRepository,
  StatRepository,
} from "../../../../src/domain/repositories";
import {
  PollCosmosMetadata,
  PollCosmosConfig,
  PollCosmos,
} from "../../../../src/domain/actions/cosmos/PollCosmos";

let props = {
  blockBatchSize: 20,
  environment: "testnet",
  commitment: "latest",
  addresses: ["sei1smzlm9t79kur392nu9egl8p8je9j92q4gzguewj56a05kyxxra0qy0nuf3"],
  interval: 5000,
  topics: [],
  chainId: 32,
  filter: {
    addresses: ["sei1smzlm9t79kur392nu9egl8p8je9j92q4gzguewj56a05kyxxra0qy0nuf3"],
    query: "wasm.action='complete_transfer_wrapped'",
  },
  chain: "sei",
  id: "poll-redeemed-transactions-sei",
};

let cfg = new PollCosmosConfig(props);

let getTransactionsSpy: jest.SpiedFunction<CosmosRepository["getTransactions"]>;
let getBlockTimestampSpy: jest.SpiedFunction<CosmosRepository["getBlockTimestamp"]>;
let handlerSpy: jest.SpiedFunction<(txs: CosmosTransaction[]) => Promise<void>>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollCosmosMetadata>["save"]>;

let metadataRepo: MetadataRepository<PollCosmosMetadata>;
let cosmosRepo: CosmosRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (txs: CosmosRepository[]) => Promise.resolve(),
  failing: (txs: CosmosRepository[]) => Promise.reject(),
};
let pollCosmos: PollCosmos;

describe("PollWCosmos", () => {
  afterEach(async () => {
    await pollCosmos.stop();
  });

  it("should be process the txs from sei blockchain and update the lastFrom height", async () => {
    const txs = [
      {
        chainId: 32,
        events: [
          {
            type: "coin_spent",
            attributes: [
              {
                key: "c3BlbmRlcg==",
                value: "c2VpMWczNjh1YTlxeWtyZGNwaGVrYTY0ZWFyNGw3M2F6NXhlZzMwM3pl",
                index: true,
              },
              { key: "YW1vdW50", value: "NjcyODZ1c2Vp", index: true },
            ],
          },
          {
            type: "tx",
            attributes: [
              { key: "ZmVl", value: "NjcyODZ1c2Vp", index: true },
              {
                key: "ZmVlX3BheWVy",
                value: "c2VpMWczNjh1YTlxeWtyZGNwaGVrYTY0ZWFyNGw3M2F6NXhlZzMwM3pl",
                index: true,
              },
            ],
          },
          {
            type: "tx",
            attributes: [
              {
                key: "YWNjX3NlcQ==",
                value: "c2VpMWczNjh1YTlxeWtyZGNwaGVrYTY0ZWFyNGw3M2F6NXhlZzMwM3plLzM2NjU=",
                index: true,
              },
            ],
          },
          {
            type: "tx",
            attributes: [
              {
                key: "c2lnbmF0dXJl",
                value:
                  "eWROMjlWM1lhWHNjUUlidUNDcUR5UFBGYjZvWWtJMVVvc2Z5SVBMTVhMa09La1ZSZWdZNTYzb3lwVzFzSTdZUXZCdkxKV1pzQlNtQmFKdGZvRjh2cVE9PQ==",
                index: true,
              },
            ],
          },
          {
            type: "signer",
            attributes: [
              {
                key: "ZXZtX2FkZHI=",
                value: "MHgzNUM4NEUzN2RkOTM2OTFFQzZFZUY1MTdmODc1MGI3NjVlMjQ5OTkz",
                index: true,
              },
              {
                key: "c2VpX2FkZHI=",
                value: "c2VpMWczNjh1YTlxeWtyZGNwaGVrYTY0ZWFyNGw3M2F6NXhlZzMwM3pl",
                index: true,
              },
            ],
          },
          {
            type: "message",
            attributes: [
              {
                key: "YWN0aW9u",
                value: "L2Nvc213YXNtLndhc20udjEuTXNnRXhlY3V0ZUNvbnRyYWN0",
                index: true,
              },
            ],
          },
          {
            type: "message",
            attributes: [
              { key: "bW9kdWxl", value: "d2FzbQ==", index: true },
              {
                key: "c2VuZGVy",
                value: "c2VpMWczNjh1YTlxeWtyZGNwaGVrYTY0ZWFyNGw3M2F6NXhlZzMwM3pl",
                index: true,
              },
            ],
          },
          {
            type: "execute",
            attributes: [
              {
                key: "X2NvbnRyYWN0X2FkZHJlc3M=",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
            ],
          },
          {
            type: "wasm",
            attributes: [
              {
                key: "X2NvbnRyYWN0X2FkZHJlc3M=",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
              { key: "YWN0aW9u", value: "Y29tcGxldGVfdHJhbnNmZXJfd2l0aF9wYXlsb2Fk", index: true },
              {
                key: "dHJhbnNmZXJfcGF5bG9hZA==",
                value:
                  "ZXlKaVlYTnBZMTl5WldOcGNHbGxiblFpT25zaWNtVmphWEJwWlc1MElqb2lZekpXY0UxWFkzcE9hbWd4V1ZSc2VHVlhkSGxhUjA1M1lVZFdjbGxVV1RCYVYwWjVUa2QzTTAweVJqWk9XR2hzV25wTmQwMHpjR3dpZlgwPQ==",
                index: true,
              },
            ],
          },
          {
            type: "execute",
            attributes: [
              {
                key: "X2NvbnRyYWN0X2FkZHJlc3M=",
                value:
                  "c2VpMXNtemxtOXQ3OWt1cjM5Mm51OWVnbDhwOGplOWo5MnE0Z3pndWV3ajU2YTA1a3l4eHJhMHF5MG51ZjM=",
                index: true,
              },
            ],
          },
          {
            type: "wasm",
            attributes: [
              {
                key: "X2NvbnRyYWN0X2FkZHJlc3M=",
                value:
                  "c2VpMXNtemxtOXQ3OWt1cjM5Mm51OWVnbDhwOGplOWo5MnE0Z3pndWV3ajU2YTA1a3l4eHJhMHF5MG51ZjM=",
                index: true,
              },
              { key: "YWN0aW9u", value: "Y29tcGxldGVfdHJhbnNmZXJfd3JhcHBlZA==", index: true },
              {
                key: "Y29udHJhY3Q=",
                value:
                  "c2VpMWxnOTV2enFrMzUzbXM5YXBwem5oZDZyMzM2c3NjcXJjNTA3NDM5bjc0d3I3cmZkdno3ZHFxYTlzbGw=",
                index: true,
              },
              {
                key: "cmVjaXBpZW50",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
              { key: "YW1vdW50", value: "NTEyOTcwODY=", index: true },
              {
                key: "cmVsYXllcg==",
                value: "c2VpMWczNjh1YTlxeWtyZGNwaGVrYTY0ZWFyNGw3M2F6NXhlZzMwM3pl",
                index: true,
              },
              { key: "ZmVl", value: "MA==", index: true },
            ],
          },
          {
            type: "execute",
            attributes: [
              {
                key: "X2NvbnRyYWN0X2FkZHJlc3M=",
                value:
                  "c2VpMWxnOTV2enFrMzUzbXM5YXBwem5oZDZyMzM2c3NjcXJjNTA3NDM5bjc0d3I3cmZkdno3ZHFxYTlzbGw=",
                index: true,
              },
            ],
          },
          {
            type: "wasm",
            attributes: [
              {
                key: "X2NvbnRyYWN0X2FkZHJlc3M=",
                value:
                  "c2VpMWxnOTV2enFrMzUzbXM5YXBwem5oZDZyMzM2c3NjcXJjNTA3NDM5bjc0d3I3cmZkdno3ZHFxYTlzbGw=",
                index: true,
              },
              { key: "YWN0aW9u", value: "bWludA==", index: true },
              {
                key: "dG8=",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
              { key: "YW1vdW50", value: "NTEyOTcwODY=", index: true },
            ],
          },
          {
            type: "reply",
            attributes: [
              {
                key: "X2NvbnRyYWN0X2FkZHJlc3M=",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
            ],
          },
          {
            type: "coin_received",
            attributes: [
              {
                key: "cmVjZWl2ZXI=",
                value: "c2VpMTllank4bjlxc2VjdHJmNHNlbWRwOWNwa25mbGxkMGo2c3Z2bXRx",
                index: true,
              },
              {
                key: "YW1vdW50",
                value:
                  "NTEyOTcwODZmYWN0b3J5L3NlaTE4OWFkZ3Vhd3VnazNlNTV6bjYzejhyOWxsMjl4cmp3Y2E2MzZyYTd2N2d4dXpuOThzeHlxd3p0NDdsL0hxNHR1RHpoUkJueHczdEZBNW42TTUyTlZNVmNDMTlYZ2dieURpSktDRDZI",
                index: true,
              },
            ],
          },
          {
            type: "coinbase",
            attributes: [
              {
                key: "bWludGVy",
                value: "c2VpMTllank4bjlxc2VjdHJmNHNlbWRwOWNwa25mbGxkMGo2c3Z2bXRx",
                index: true,
              },
              {
                key: "YW1vdW50",
                value:
                  "NTEyOTcwODZmYWN0b3J5L3NlaTE4OWFkZ3Vhd3VnazNlNTV6bjYzejhyOWxsMjl4cmp3Y2E2MzZyYTd2N2d4dXpuOThzeHlxd3p0NDdsL0hxNHR1RHpoUkJueHczdEZBNW42TTUyTlZNVmNDMTlYZ2dieURpSktDRDZI",
                index: true,
              },
            ],
          },
          {
            type: "coin_spent",
            attributes: [
              {
                key: "c3BlbmRlcg==",
                value: "c2VpMTllank4bjlxc2VjdHJmNHNlbWRwOWNwa25mbGxkMGo2c3Z2bXRx",
                index: true,
              },
              {
                key: "YW1vdW50",
                value:
                  "NTEyOTcwODZmYWN0b3J5L3NlaTE4OWFkZ3Vhd3VnazNlNTV6bjYzejhyOWxsMjl4cmp3Y2E2MzZyYTd2N2d4dXpuOThzeHlxd3p0NDdsL0hxNHR1RHpoUkJueHczdEZBNW42TTUyTlZNVmNDMTlYZ2dieURpSktDRDZI",
                index: true,
              },
            ],
          },
          {
            type: "coin_received",
            attributes: [
              {
                key: "cmVjZWl2ZXI=",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
              {
                key: "YW1vdW50",
                value:
                  "NTEyOTcwODZmYWN0b3J5L3NlaTE4OWFkZ3Vhd3VnazNlNTV6bjYzejhyOWxsMjl4cmp3Y2E2MzZyYTd2N2d4dXpuOThzeHlxd3p0NDdsL0hxNHR1RHpoUkJueHczdEZBNW42TTUyTlZNVmNDMTlYZ2dieURpSktDRDZI",
                index: true,
              },
            ],
          },
          {
            type: "transfer",
            attributes: [
              {
                key: "cmVjaXBpZW50",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
              {
                key: "c2VuZGVy",
                value: "c2VpMTllank4bjlxc2VjdHJmNHNlbWRwOWNwa25mbGxkMGo2c3Z2bXRx",
                index: true,
              },
              {
                key: "YW1vdW50",
                value:
                  "NTEyOTcwODZmYWN0b3J5L3NlaTE4OWFkZ3Vhd3VnazNlNTV6bjYzejhyOWxsMjl4cmp3Y2E2MzZyYTd2N2d4dXpuOThzeHlxd3p0NDdsL0hxNHR1RHpoUkJueHczdEZBNW42TTUyTlZNVmNDMTlYZ2dieURpSktDRDZI",
                index: true,
              },
            ],
          },
          {
            type: "mint",
            attributes: [
              {
                key: "bWludF90b19hZGRyZXNz",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
              {
                key: "YW1vdW50",
                value:
                  "NTEyOTcwODZmYWN0b3J5L3NlaTE4OWFkZ3Vhd3VnazNlNTV6bjYzejhyOWxsMjl4cmp3Y2E2MzZyYTd2N2d4dXpuOThzeHlxd3p0NDdsL0hxNHR1RHpoUkJueHczdEZBNW42TTUyTlZNVmNDMTlYZ2dieURpSktDRDZI",
                index: true,
              },
            ],
          },
          {
            type: "coin_spent",
            attributes: [
              {
                key: "c3BlbmRlcg==",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
              {
                key: "YW1vdW50",
                value:
                  "NTEyOTcwODZmYWN0b3J5L3NlaTE4OWFkZ3Vhd3VnazNlNTV6bjYzejhyOWxsMjl4cmp3Y2E2MzZyYTd2N2d4dXpuOThzeHlxd3p0NDdsL0hxNHR1RHpoUkJueHczdEZBNW42TTUyTlZNVmNDMTlYZ2dieURpSktDRDZI",
                index: true,
              },
            ],
          },
          {
            type: "coin_received",
            attributes: [
              {
                key: "cmVjZWl2ZXI=",
                value: "c2VpMWczNjh1YTlxeWtyZGNwaGVrYTY0ZWFyNGw3M2F6NXhlZzMwM3pl",
                index: true,
              },
              {
                key: "YW1vdW50",
                value:
                  "NTEyOTcwODZmYWN0b3J5L3NlaTE4OWFkZ3Vhd3VnazNlNTV6bjYzejhyOWxsMjl4cmp3Y2E2MzZyYTd2N2d4dXpuOThzeHlxd3p0NDdsL0hxNHR1RHpoUkJueHczdEZBNW42TTUyTlZNVmNDMTlYZ2dieURpSktDRDZI",
                index: true,
              },
            ],
          },
          {
            type: "transfer",
            attributes: [
              {
                key: "cmVjaXBpZW50",
                value: "c2VpMWczNjh1YTlxeWtyZGNwaGVrYTY0ZWFyNGw3M2F6NXhlZzMwM3pl",
                index: true,
              },
              {
                key: "c2VuZGVy",
                value:
                  "c2VpMTg5YWRndWF3dWdrM2U1NXpuNjN6OHI5bGwyOXhyandjYTYzNnJhN3Y3Z3h1em45OHN4eXF3enQ0N2w=",
                index: true,
              },
              {
                key: "YW1vdW50",
                value:
                  "NTEyOTcwODZmYWN0b3J5L3NlaTE4OWFkZ3Vhd3VnazNlNTV6bjYzejhyOWxsMjl4cmp3Y2E2MzZyYTd2N2d4dXpuOThzeHlxd3p0NDdsL0hxNHR1RHpoUkJueHczdEZBNW42TTUyTlZNVmNDMTlYZ2dieURpSktDRDZI",
                index: true,
              },
            ],
          },
        ],
        height: 80542798n,
        data: "CiYKJC9jb3Ntd2FzbS53YXNtLnYxLk1zZ0V4ZWN1dGVDb250cmFjdA==",
        hash: "C94451E3FCDC064E9E7FCAB6FBBFADD9C909910817BB185FF651580A72B3034B",
        tx: {
          type: Buffer.from([
            10, 230, 13, 10, 180, 13, 10, 36, 47, 99, 111, 115, 109, 119, 97, 115, 109, 46, 119, 97,
            115, 109, 46, 118, 49, 46, 77, 115, 103, 69, 120, 101, 99, 117, 116, 101, 67, 111, 110,
            116, 114, 97, 99, 116, 18, 139, 13, 10, 42, 115, 101, 105, 49, 113, 51, 114, 50, 118,
            57, 106, 55, 112, 102, 48, 114, 109, 106, 56, 52, 54, 104, 51, 108, 97, 107, 107, 56,
            53, 114, 118, 54, 114, 50, 53, 116, 108, 112, 103, 53, 57, 122, 18, 62, 115, 101, 105,
            49, 56, 57, 97, 100, 103, 117, 97, 119, 117, 103, 107, 51, 101, 53, 53, 122, 110, 54,
            51, 122, 56, 114, 57, 108, 108, 50, 57, 120, 114, 106, 119, 99, 97, 54, 51, 54, 114, 97,
            55, 118, 55, 103, 120, 117, 122, 110, 57, 56, 115, 120, 121, 113, 119, 122, 116, 52, 55,
            108, 26, 156, 12, 123, 34, 99, 111, 109, 112, 108, 101, 116, 101, 95, 116, 114, 97, 110,
            115, 102, 101, 114, 95, 97, 110, 100, 95, 99, 111, 110, 118, 101, 114, 116, 34, 58, 123,
            34, 118, 97, 97, 34, 58, 34, 65, 81, 65, 65, 65, 65, 81, 78, 65, 80, 115, 111, 57, 97,
            73, 43, 99, 65, 48, 100, 56, 70, 114, 74, 98, 74, 71, 101, 97, 71, 70, 68, 113, 104, 49,
            98, 81, 98, 87, 75, 79, 102, 68, 77, 70, 105, 111, 102, 69, 76, 102, 111, 79, 103, 106,
            102, 73, 118, 111, 87, 89, 83, 104, 55, 110, 53, 43, 48, 88, 82, 74, 68, 84, 43, 113,
            48, 102, 120, 121, 90, 66, 101, 115, 68, 73, 71, 85, 106, 50, 52, 80, 114, 107, 88, 73,
            65, 65, 85, 113, 121, 87, 122, 52, 110, 54, 104, 112, 77, 89, 75, 85, 122, 97, 103, 53,
            49, 56, 120, 121, 99, 66, 86, 76, 69, 50, 48, 51, 47, 88, 74, 85, 56, 65, 75, 70, 76,
            101, 118, 120, 101, 76, 71, 98, 55, 72, 48, 43, 113, 54, 53, 73, 105, 112, 53, 118, 56,
            109, 116, 70, 78, 78, 71, 88, 116, 104, 104, 119, 49, 81, 113, 83, 99, 114, 122, 119,
            121, 77, 84, 67, 107, 79, 85, 56, 65, 66, 74, 99, 106, 82, 78, 87, 70, 109, 105, 100,
            112, 88, 84, 51, 75, 118, 55, 68, 79, 103, 78, 120, 81, 57, 75, 50, 84, 78, 97, 106, 86,
            121, 97, 111, 118, 81, 51, 120, 78, 71, 86, 80, 82, 80, 86, 74, 117, 110, 68, 118, 71,
            47, 89, 68, 90, 112, 106, 102, 110, 66, 78, 67, 111, 99, 52, 106, 80, 102, 55, 79, 90,
            52, 111, 119, 83, 102, 74, 65, 76, 111, 119, 113, 43, 73, 70, 89, 65, 66, 113, 87, 67,
            87, 65, 72, 67, 76, 82, 83, 122, 107, 65, 116, 112, 106, 106, 82, 98, 73, 47, 56, 97,
            110, 117, 97, 74, 85, 89, 114, 52, 47, 119, 66, 98, 72, 107, 98, 71, 109, 89, 113, 87,
            86, 53, 119, 117, 67, 104, 121, 49, 84, 105, 80, 103, 102, 118, 73, 118, 76, 66, 78,
            106, 71, 80, 99, 88, 65, 84, 49, 65, 48, 108, 111, 97, 79, 55, 110, 78, 52, 47, 116, 52,
            53, 83, 111, 66, 66, 55, 55, 99, 88, 113, 99, 117, 106, 49, 109, 118, 67, 121, 99, 100,
            121, 70, 65, 108, 80, 75, 69, 69, 78, 56, 55, 101, 101, 52, 50, 109, 112, 109, 50, 56,
            97, 98, 84, 80, 72, 72, 86, 83, 66, 100, 82, 80, 52, 77, 112, 57, 104, 111, 114, 53, 86,
            121, 120, 87, 121, 66, 47, 73, 48, 76, 84, 68, 88, 54, 82, 49, 71, 55, 102, 122, 77, 57,
            52, 71, 116, 70, 48, 121, 104, 73, 69, 65, 67, 69, 50, 102, 75, 86, 118, 118, 104, 117,
            66, 49, 118, 103, 84, 119, 86, 55, 82, 71, 78, 109, 114, 121, 76, 75, 89, 105, 52, 110,
            74, 103, 65, 119, 87, 116, 56, 71, 88, 113, 51, 55, 65, 43, 87, 99, 79, 117, 101, 102,
            101, 70, 53, 53, 75, 114, 90, 86, 101, 118, 71, 101, 99, 80, 115, 74, 114, 74, 72, 110,
            90, 71, 114, 86, 99, 54, 122, 86, 75, 110, 43, 105, 50, 118, 73, 43, 89, 65, 67, 98, 48,
            88, 108, 100, 51, 67, 84, 49, 65, 81, 83, 55, 51, 109, 101, 52, 79, 80, 119, 109, 72,
            113, 52, 119, 76, 75, 71, 114, 77, 52, 115, 89, 76, 78, 48, 87, 111, 70, 85, 49, 53, 51,
            86, 103, 112, 90, 53, 110, 113, 110, 88, 112, 117, 122, 122, 70, 71, 84, 90, 57, 88, 66,
            116, 113, 51, 81, 111, 53, 102, 52, 116, 119, 108, 116, 71, 57, 72, 104, 90, 101, 121,
            47, 103, 80, 81, 66, 67, 110, 102, 112, 66, 99, 83, 100, 107, 75, 69, 83, 50, 74, 80,
            105, 109, 77, 107, 110, 84, 43, 112, 116, 68, 113, 50, 117, 68, 100, 88, 50, 120, 77,
            117, 114, 56, 85, 119, 83, 71, 50, 111, 82, 81, 122, 110, 77, 77, 98, 81, 122, 114, 53,
            115, 111, 53, 82, 72, 113, 71, 55, 48, 43, 74, 80, 89, 72, 104, 47, 105, 108, 101, 79,
            49, 86, 54, 75, 107, 75, 52, 69, 65, 57, 120, 120, 99, 65, 68, 88, 70, 88, 70, 77, 104,
            111, 69, 108, 116, 90, 86, 76, 47, 78, 108, 66, 97, 49, 115, 69, 50, 74, 117, 75, 120,
            67, 70, 81, 55, 56, 88, 54, 55, 81, 107, 87, 119, 119, 53, 108, 102, 69, 88, 111, 65,
            119, 47, 98, 118, 110, 117, 47, 115, 77, 74, 48, 99, 43, 86, 71, 84, 113, 83, 49, 99,
            89, 77, 69, 55, 98, 112, 111, 51, 56, 107, 51, 67, 52, 79, 115, 53, 114, 117, 117, 89,
            66, 68, 107, 76, 82, 90, 89, 52, 116, 85, 99, 120, 118, 75, 43, 106, 104, 65, 47, 51,
            87, 86, 87, 110, 109, 106, 103, 107, 98, 108, 68, 97, 112, 74, 109, 81, 50, 65, 69, 57,
            88, 56, 100, 90, 88, 84, 90, 50, 103, 111, 49, 54, 110, 71, 89, 76, 80, 97, 100, 43,
            122, 43, 90, 74, 53, 107, 50, 57, 72, 51, 102, 117, 69, 86, 71, 110, 99, 66, 87, 47, 65,
            117, 69, 78, 98, 109, 114, 115, 66, 68, 43, 99, 78, 109, 70, 97, 43, 49, 80, 87, 71, 51,
            97, 105, 77, 102, 110, 105, 107, 101, 111, 98, 111, 78, 122, 79, 102, 70, 102, 113, 43,
            111, 76, 114, 118, 75, 72, 55, 116, 103, 114, 51, 105, 68, 69, 119, 87, 69, 99, 101,
            119, 83, 98, 55, 112, 79, 105, 111, 121, 88, 50, 69, 108, 54, 78, 52, 114, 118, 85, 113,
            112, 102, 82, 66, 84, 105, 103, 56, 109, 118, 86, 84, 99, 105, 84, 77, 65, 69, 74, 90,
            56, 115, 110, 90, 121, 113, 111, 122, 53, 83, 56, 85, 79, 76, 116, 67, 121, 111, 102,
            57, 100, 101, 83, 89, 72, 51, 48, 111, 105, 81, 98, 68, 84, 111, 90, 75, 97, 72, 114,
            83, 80, 70, 81, 55, 105, 47, 78, 107, 52, 122, 103, 98, 54, 50, 113, 104, 43, 120, 98,
            108, 88, 112, 87, 102, 52, 75, 112, 68, 68, 105, 87, 43, 53, 117, 54, 72, 88, 66, 89,
            51, 88, 74, 47, 115, 65, 69, 106, 115, 102, 75, 79, 113, 55, 68, 103, 109, 89, 116, 118,
            73, 120, 68, 100, 56, 66, 50, 98, 57, 122, 49, 116, 89, 68, 57, 122, 102, 100, 70, 49,
            98, 107, 52, 85, 112, 71, 86, 119, 83, 122, 70, 111, 80, 54, 55, 57, 81, 43, 103, 84,
            75, 43, 120, 67, 49, 55, 57, 51, 118, 80, 78, 65, 86, 116, 109, 76, 113, 81, 98, 84, 84,
            54, 54, 115, 79, 81, 75, 111, 79, 102, 112, 99, 107, 66, 90, 108, 86, 121, 116, 81, 65,
            65, 103, 74, 115, 65, 65, 101, 120, 122, 99, 112, 108, 100, 88, 77, 104, 122, 73, 53,
            102, 55, 67, 116, 78, 99, 65, 83, 72, 103, 54, 113, 107, 78, 74, 118, 103, 111, 112, 84,
            84, 75, 116, 85, 79, 82, 115, 54, 84, 49, 65, 65, 65, 65, 65, 65, 65, 78, 99, 109, 81,
            103, 65, 119, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65,
            65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 79, 53, 114,
            75, 65, 120, 118, 112, 54, 56, 55, 55, 98, 114, 84, 111, 57, 90, 102, 78, 113, 113, 56,
            108, 48, 77, 98, 71, 55, 53, 77, 76, 83, 57, 117, 68, 107, 102, 75, 89, 67, 65, 48, 85,
            118, 88, 87, 69, 65, 65, 84, 108, 54, 49, 72, 79, 117, 52, 105, 48, 99, 48, 111, 75,
            101, 111, 105, 79, 77, 118, 47, 113, 75, 89, 99, 110, 89, 55, 113, 79, 104, 57, 56, 122,
            121, 68, 99, 70, 77, 112, 52, 71, 73, 65, 67, 67, 57, 109, 108, 100, 84, 83, 52, 79, 83,
            86, 118, 117, 56, 74, 107, 121, 70, 70, 104, 79, 117, 104, 79, 43, 47, 100, 119, 87, 80,
            51, 99, 115, 56, 73, 84, 110, 103, 113, 89, 113, 82, 69, 110, 115, 105, 89, 109, 70,
            122, 97, 87, 78, 102, 99, 109, 86, 106, 97, 88, 66, 112, 90, 87, 53, 48, 73, 106, 112,
            55, 73, 110, 74, 108, 89, 50, 108, 119, 97, 87, 86, 117, 100, 67, 73, 54, 73, 109, 77,
            121, 86, 110, 66, 78, 86, 122, 69, 49, 84, 84, 78, 119, 78, 107, 53, 85, 85, 110, 104,
            104, 98, 108, 111, 48, 89, 48, 100, 107, 101, 70, 111, 121, 86, 88, 112, 108, 86, 51,
            104, 116, 90, 71, 112, 115, 101, 108, 108, 88, 77, 72, 108, 107, 87, 69, 112, 117, 89,
            50, 112, 97, 97, 108, 108, 85, 81, 88, 112, 79, 86, 48, 48, 48, 73, 110, 49, 57, 34,
            125, 125, 18, 45, 87, 111, 114, 109, 104, 111, 108, 101, 32, 45, 32, 67, 111, 109, 112,
            108, 101, 116, 101, 32, 84, 111, 107, 101, 110, 32, 84, 114, 97, 110, 115, 108, 97, 116,
            111, 114, 32, 84, 114, 97, 110, 115, 102, 101, 114, 18, 151, 1, 10, 82, 10, 70, 10, 31,
            47, 99, 111, 115, 109, 111, 115, 46, 99, 114, 121, 112, 116, 111, 46, 115, 101, 99, 112,
            50, 53, 54, 107, 49, 46, 80, 117, 98, 75, 101, 121, 18, 35, 10, 33, 3, 185, 116, 9, 72,
            116, 44, 202, 69, 71, 52, 114, 33, 78, 210, 187, 51, 206, 29, 125, 231, 178, 228, 13,
            138, 230, 163, 33, 114, 14, 214, 28, 212, 18, 4, 10, 2, 8, 1, 24, 184, 215, 1, 18, 65,
            10, 14, 10, 4, 117, 115, 101, 105, 18, 6, 51, 48, 48, 48, 48, 48, 16, 192, 141, 183, 1,
            34, 42, 115, 101, 105, 49, 122, 51, 114, 48, 99, 99, 115, 115, 115, 110, 118, 117, 97,
            104, 101, 117, 97, 107, 117, 108, 53, 56, 122, 108, 117, 54, 53, 114, 110, 103, 119, 55,
            110, 106, 114, 106, 99, 122, 26, 64, 149, 61, 175, 131, 175, 91, 106, 106, 66, 225, 99,
            170, 24, 61, 32, 138, 116, 119, 199, 204, 144, 193, 231, 139, 236, 226, 189, 39, 160,
            72, 104, 183, 108, 9, 253, 45, 17, 28, 158, 91, 188, 252, 20, 75, 13, 48, 8, 253, 45,
            67, 193, 181, 32, 24, 126, 152, 15, 242, 109, 18, 211, 225, 99, 16,
          ]),
        },
      },
    ];
    givenCosmosBlockRepository(12312312312, txs);
    givenMetadataRepository();
    givenStatsRepository();
    givenPollWormchainLogs();

    await whenPollWormchainLogsStarts();

    await thenWaitForAssertion(
      () => expect(getTransactionsSpy).toHaveReturnedTimes(1),
      () =>
        expect(getTransactionsSpy).toHaveBeenCalledWith(
          {
            addresses: ["sei1smzlm9t79kur392nu9egl8p8je9j92q4gzguewj56a05kyxxra0qy0nuf3"],
            query: "wasm.action='complete_transfer_wrapped'",
          },
          20,
          "sei"
        ),
      () => expect(getBlockTimestampSpy).toHaveBeenCalledWith(80542798n, "sei")
    );
  });
});

const givenCosmosBlockRepository = (timestamp?: number, txs: any = []) => {
  cosmosRepo = {
    getTransactions: () => Promise.resolve(txs),
    getBlockTimestamp: () => Promise.resolve(timestamp),
    healthCheck: () => Promise.resolve(),
  };

  getTransactionsSpy = jest.spyOn(cosmosRepo, "getTransactions");
  getBlockTimestampSpy = jest.spyOn(cosmosRepo, "getBlockTimestamp");
};

const givenMetadataRepository = (data?: PollCosmosMetadata) => {
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

const givenPollWormchainLogs = (from?: bigint) => {
  cfg.setFromBlock(from);
  pollCosmos = new PollCosmos(cosmosRepo, metadataRepo, statsRepo, cfg);
};

const whenPollWormchainLogsStarts = async () => {
  pollCosmos.run([handlers.working]);
};
