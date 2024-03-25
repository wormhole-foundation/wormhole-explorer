import { wormchainLogMessagePublishedMapper } from "../../../../src/infrastructure/mappers/wormchain/wormchainLogMessagePublishedMapper";
import { describe, it, expect } from "@jest/globals";
import { WormchainLog } from "../../../../src/domain/entities/wormchain";

describe("wormchainLogMessagePublishedMapper", () => {
  it("should be able to map log to aptosLogMessagePublishedMapper", async () => {
    // When
    const result = wormchainLogMessagePublishedMapper(log);

    if (result) {
      // Then
      expect(result.name).toBe("log-message-published");
      expect(result.chainId).toBe(3104);
      expect(result.txHash).toBe(
        "0xa08b0ac6ee67e21d3dd89f48f60cc907fc867288f4439bcf72731b0884d8aff2"
      );
      expect(result.address).toBe(
        "wormhole1ufs3tlq4umljk0qfe8k5ya0x6hpavn897u2cnf9k0en9jr7qarqqaqfk2j"
      );
      expect(result.attributes.consistencyLevel).toBe(0);
      expect(result.attributes.nonce).toBe(7671);
      expect(result.attributes.payload).toBe(
        "0100000000000000000000000000000000000000000000000000000000555643a3f5edec8471c75624ebc4079a634326d96a689e6157d79abe8f5a6f94472853bc00018622b98735cb870ae0cb22bd4ea58cfb512bd4002247ccd0b250eb6d0c5032fc00010000000000000000000000000000000000000000000000000000000000000000"
      );
      expect(result.attributes.sender).toBe(
        "aeb534c45c3049d380b9d9b966f9895f53abd4301bfaff407fa09dea8ae7a924"
      );
      expect(result.attributes.sequence).toBe(28603);
    }
  });
});

const log: WormchainLog = {
  transactions: [
    {
      hash: "0x987e77d2d8cf8b9c0b3998dc62dc94fad9de47c4e3b50ad9bfd3083d7ab958ff",
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
          value: "d29ybWhvbGUxODc4a3h6M3VnZXN2YTRoNGtmeng2Y3F0ZHk5NmN3d2RqajBwaHc=",
          index: true,
        },
      ],
    },
    {
      hash: "0xa08b0ac6ee67e21d3dd89f48f60cc907fc867288f4439bcf72731b0884d8aff2",
      type: "wasm",
      attributes: [
        {
          key: "X2NvbnRyYWN0X2FkZHJlc3M=",
          value:
            "d29ybWhvbGUxajYydGt5cWhqeWpscXN5MzB1bnVhcW5uZDhkdDV3cXFucndqemYwZms3bnc0dzdkeHBzcWhheWFuZw==",
          index: true,
        },
        { key: "YWN0aW9u", value: "aW5jcmVhc2VfYWxsb3dhbmNl", index: true },
        { key: "YW1vdW50", value: "MTQzMTcxNjc3MQ==", index: true },
        {
          key: "b3duZXI=",
          value:
            "d29ybWhvbGUxNGVqcWp5cTh1bTRwM3hmcWo3NHlsZDV3YXFsamY4OGZ6MjV5eG5tYTBjbmdzcHhlM2xlczAwZnBqeA==",
          index: true,
        },
        {
          key: "c3BlbmRlcg==",
          value:
            "d29ybWhvbGUxNDY2bmYzenV4cHlhOHE5ZW14dWtkN3ZmdGFmNmg0cHNyMGEwN3NybDV6dzc0emg4NHlqcTRseWptaA==",
          index: true,
        },
      ],
    },
    {
      hash: "0xa08b0ac6ee67e21d3dd89f48f60cc907fc867288f4439bcf72731b0884d8aff2",
      type: "wasm",
      attributes: [
        {
          key: "X2NvbnRyYWN0X2FkZHJlc3M=",
          value:
            "d29ybWhvbGUxNDY2bmYzenV4cHlhOHE5ZW14dWtkN3ZmdGFmNmg0cHNyMGEwN3NybDV6dzc0emg4NHlqcTRseWptaA==",
          index: true,
        },
        { key: "dHJhbnNmZXIuYW1vdW50", value: "MTQzMTcxNjc3MQ==", index: true },
        { key: "dHJhbnNmZXIuYmxvY2tfdGltZQ==", value: "MTcxMTE0MzIyMg==", index: true },
        { key: "dHJhbnNmZXIubm9uY2U=", value: "NzY3MQ==", index: true },
        {
          key: "dHJhbnNmZXIucmVjaXBpZW50",
          value:
            "ODYyMmI5ODczNWNiODcwYWUwY2IyMmJkNGVhNThjZmI1MTJiZDQwMDIyNDdjY2QwYjI1MGViNmQwYzUwMzJmYw==",
          index: true,
        },
        { key: "dHJhbnNmZXIucmVjaXBpZW50X2NoYWlu", value: "MQ==", index: true },
        {
          key: "dHJhbnNmZXIuc2VuZGVy",
          value:
            "YWU2NDA5MTAwN2U2ZWExODk5MjA5N2FhNGZiNjhlZTgzZjI0OWNlOTEyYTg0MzRmN2Q3ZTI2ODgwNGQ5OGZmMw==",
          index: true,
        },
        {
          key: "dHJhbnNmZXIudG9rZW4=",
          value:
            "ZjVlZGVjODQ3MWM3NTYyNGViYzQwNzlhNjM0MzI2ZDk2YTY4OWU2MTU3ZDc5YWJlOGY1YTZmOTQ0NzI4NTNiYw==",
          index: true,
        },
        { key: "dHJhbnNmZXIudG9rZW5fY2hhaW4=", value: "MQ==", index: true },
      ],
    },
    {
      hash: "0xa08b0ac6ee67e21d3dd89f48f60cc907fc867288f4439bcf72731b0884d8aff2",
      type: "wasm",
      attributes: [
        {
          key: "X2NvbnRyYWN0X2FkZHJlc3M=",
          value:
            "d29ybWhvbGUxajYydGt5cWhqeWpscXN5MzB1bnVhcW5uZDhkdDV3cXFucndqemYwZms3bnc0dzdkeHBzcWhheWFuZw==",
          index: true,
        },
        { key: "YWN0aW9u", value: "YnVybl9mcm9t", index: true },
        { key: "YW1vdW50", value: "MTQzMTcxNjc3MQ==", index: true },
        {
          key: "Ynk=",
          value:
            "d29ybWhvbGUxNDY2bmYzenV4cHlhOHE5ZW14dWtkN3ZmdGFmNmg0cHNyMGEwN3NybDV6dzc0emg4NHlqcTRseWptaA==",
          index: true,
        },
        {
          key: "ZnJvbQ==",
          value:
            "d29ybWhvbGUxNGVqcWp5cTh1bTRwM3hmcWo3NHlsZDV3YXFsamY4OGZ6MjV5eG5tYTBjbmdzcHhlM2xlczAwZnBqeA==",
          index: true,
        },
      ],
    },
    {
      hash: "0xa08b0ac6ee67e21d3dd89f48f60cc907fc867288f4439bcf72731b0884d8aff2",
      type: "wasm",
      attributes: [
        {
          key: "X2NvbnRyYWN0X2FkZHJlc3M=",
          value:
            "d29ybWhvbGUxdWZzM3RscTR1bWxqazBxZmU4azV5YTB4NmhwYXZuODk3dTJjbmY5azBlbjlqcjdxYXJxcWFxZmsyag==",
          index: true,
        },
        { key: "bWVzc2FnZS5ibG9ja190aW1l", value: "MTcxMTE0MzIyMg==", index: true },
        { key: "bWVzc2FnZS5jaGFpbl9pZA==", value: "MzEwNA==", index: true },
        {
          key: "bWVzc2FnZS5tZXNzYWdl",
          value:
            "MDEwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDU1NTY0M2EzZjVlZGVjODQ3MWM3NTYyNGViYzQwNzlhNjM0MzI2ZDk2YTY4OWU2MTU3ZDc5YWJlOGY1YTZmOTQ0NzI4NTNiYzAwMDE4NjIyYjk4NzM1Y2I4NzBhZTBjYjIyYmQ0ZWE1OGNmYjUxMmJkNDAwMjI0N2NjZDBiMjUwZWI2ZDBjNTAzMmZjMDAwMTAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDA=",
          index: true,
        },
        { key: "bWVzc2FnZS5ub25jZQ==", value: "NzY3MQ==", index: true },
        {
          key: "bWVzc2FnZS5zZW5kZXI=",
          value:
            "YWViNTM0YzQ1YzMwNDlkMzgwYjlkOWI5NjZmOTg5NWY1M2FiZDQzMDFiZmFmZjQwN2ZhMDlkZWE4YWU3YTkyNA==",
          index: true,
        },
        { key: "bWVzc2FnZS5zZXF1ZW5jZQ==", value: "Mjg2MDM=", index: true },
      ],
    },
  ],
  blockHeight: 7626736n,
  timestamp: 1711143222043,
};
