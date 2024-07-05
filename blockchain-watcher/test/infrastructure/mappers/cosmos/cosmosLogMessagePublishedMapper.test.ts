import { cosmosLogMessagePublishedMapper } from "../../../../src/infrastructure/mappers/cosmos/cosmosLogMessagePublishedMapper";
import { CosmosTransaction } from "../../../../src/domain/entities/Cosmos";
import { describe, it, expect } from "@jest/globals";

describe("cosmosLogMessagePublishedMapper", () => {
  it("should be able to map tx to cosmosLogMessagePublishedMapper", async () => {
    // When
    const result = cosmosLogMessagePublishedMapper(
      ["terra12mrnzvhx3rpej6843uge2yyfppfyd3u9c3uq223q8sl48huz9juqffcnhp"],
      tx
    ) as any;

    // Then
    expect(result.name).toBe("log-message-published");
    expect(result.chainId).toBe(18);
    expect(result.txHash).toBe("F860EA95AC350D50354290349B2027EA01DE64AAA144381E22FBB97EE6FFA215");
    expect(result.address).toBe("terra12mrnzvhx3rpej6843uge2yyfppfyd3u9c3uq223q8sl48huz9juqffcnhp");
    expect(result.attributes.consistencyLevel).toBe(0);
    expect(result.attributes.nonce).toBe(66846);
    expect(result.attributes.payload).toBe(
      "0100000000000000000000000000000000000000000000000000000000000f4240069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f0000000000100012e27f04bb682de440fa03920e5823953b3ec86a57bae563afcc41f753ef8fd1800010000000000000000000000000000000000000000000000000000000000000000"
    );
    expect(result.attributes.sender).toBe(
      "a463ad028fb79679cfc8ce1efba35ac0e77b35080a1abe9bebe83461f176b0a3"
    );
    expect(result.attributes.sequence).toBe(3417);
  });

  const tx: CosmosTransaction = {
    chainId: 18,
    events: [
      {
        type: "coin_spent",
        attributes: [
          { key: "spender", value: "terra1kgs8ld9wyvyedjavy62kd05vrf6dkurehgl2vg", index: true },
          { key: "amount", value: "11350uluna", index: true },
        ],
      },
      {
        type: "coin_received",
        attributes: [
          { key: "receiver", value: "terra17xpfvakm2amg962yls6f84z3kell8c5lkaeqfa", index: true },
          { key: "amount", value: "11350uluna", index: true },
        ],
      },
      {
        type: "transfer",
        attributes: [
          { key: "recipient", value: "terra17xpfvakm2amg962yls6f84z3kell8c5lkaeqfa", index: true },
          { key: "sender", value: "terra1kgs8ld9wyvyedjavy62kd05vrf6dkurehgl2vg", index: true },
          { key: "amount", value: "11350uluna", index: true },
        ],
      },
      {
        type: "message",
        attributes: [
          { key: "sender", value: "terra1kgs8ld9wyvyedjavy62kd05vrf6dkurehgl2vg", index: true },
        ],
      },
      {
        type: "tx",
        attributes: [
          { key: "fee", value: "11350uluna", index: true },
          { key: "fee_payer", value: "terra1kgs8ld9wyvyedjavy62kd05vrf6dkurehgl2vg", index: true },
        ],
      },
      {
        type: "tx",
        attributes: [
          { key: "acc_seq", value: "terra1kgs8ld9wyvyedjavy62kd05vrf6dkurehgl2vg/13", index: true },
        ],
      },
      {
        type: "tx",
        attributes: [
          {
            key: "signature",
            value:
              "73EqTWK2i8+4Gl6gjRpepoqnUKwfyMqB4WFXoE4eiikZuZulXDGgnY7CZV9CGA11qa8F/oudaOXZJ/8WzKczaA==",
            index: true,
          },
        ],
      },
      {
        type: "message",
        attributes: [
          { key: "action", value: "/cosmwasm.wasm.v1.MsgExecuteContract", index: true },
          { key: "sender", value: "terra1kgs8ld9wyvyedjavy62kd05vrf6dkurehgl2vg", index: true },
          { key: "module", value: "wasm", index: true },
        ],
      },
      {
        type: "execute",
        attributes: [
          {
            key: "_contract_address",
            value: "terra1ctelwayk6t2zu30a8v9kdg3u2gr0slpjdfny5pjp7m3tuquk32ysugyjdg",
            index: true,
          },
        ],
      },
      {
        type: "wasm",
        attributes: [
          {
            key: "_contract_address",
            value: "terra1ctelwayk6t2zu30a8v9kdg3u2gr0slpjdfny5pjp7m3tuquk32ysugyjdg",
            index: true,
          },
          { key: "action", value: "increase_allowance", index: true },
          { key: "owner", value: "terra1kgs8ld9wyvyedjavy62kd05vrf6dkurehgl2vg", index: true },
          {
            key: "spender",
            value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
            index: true,
          },
          { key: "amount", value: "1000000", index: true },
        ],
      },
      {
        type: "message",
        attributes: [
          { key: "action", value: "/cosmwasm.wasm.v1.MsgExecuteContract", index: true },
          { key: "sender", value: "terra1kgs8ld9wyvyedjavy62kd05vrf6dkurehgl2vg", index: true },
          { key: "module", value: "wasm", index: true },
        ],
      },
      {
        type: "execute",
        attributes: [
          {
            key: "_contract_address",
            value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
            index: true,
          },
        ],
      },
      {
        type: "wasm",
        attributes: [
          {
            key: "_contract_address",
            value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
            index: true,
          },
          { key: "transfer.token_chain", value: "1", index: true },
          {
            key: "transfer.token",
            value: "069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f00000000001",
            index: true,
          },
          {
            key: "transfer.sender",
            value: "000000000000000000000000b2207fb4ae230996cbac269566be8c1a74db7079",
            index: true,
          },
          { key: "transfer.recipient_chain", value: "1", index: true },
          {
            key: "transfer.recipient",
            value: "2e27f04bb682de440fa03920e5823953b3ec86a57bae563afcc41f753ef8fd18",
            index: true,
          },
          { key: "transfer.amount", value: "1000000", index: true },
          { key: "transfer.nonce", value: "66846", index: true },
          { key: "transfer.block_time", value: "1716388503", index: true },
        ],
      },
      {
        type: "execute",
        attributes: [
          {
            key: "_contract_address",
            value: "terra1ctelwayk6t2zu30a8v9kdg3u2gr0slpjdfny5pjp7m3tuquk32ysugyjdg",
            index: true,
          },
        ],
      },
      {
        type: "wasm",
        attributes: [
          {
            key: "_contract_address",
            value: "terra1ctelwayk6t2zu30a8v9kdg3u2gr0slpjdfny5pjp7m3tuquk32ysugyjdg",
            index: true,
          },
          { key: "action", value: "burn_from", index: true },
          { key: "amount", value: "1000000", index: true },
          {
            key: "by",
            value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
            index: true,
          },
          { key: "from", value: "terra1kgs8ld9wyvyedjavy62kd05vrf6dkurehgl2vg", index: true },
        ],
      },
      {
        type: "execute",
        attributes: [
          {
            key: "_contract_address",
            value: "terra12mrnzvhx3rpej6843uge2yyfppfyd3u9c3uq223q8sl48huz9juqffcnhp",
            index: true,
          },
        ],
      },
      {
        type: "wasm",
        attributes: [
          {
            key: "_contract_address",
            value: "terra12mrnzvhx3rpej6843uge2yyfppfyd3u9c3uq223q8sl48huz9juqffcnhp",
            index: true,
          },
          { key: "message.block_time", value: "1716388503", index: true },
          { key: "message.chain_id", value: "18", index: true },
          {
            key: "message.message",
            value:
              "0100000000000000000000000000000000000000000000000000000000000f4240069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f0000000000100012e27f04bb682de440fa03920e5823953b3ec86a57bae563afcc41f753ef8fd1800010000000000000000000000000000000000000000000000000000000000000000",
            index: true,
          },
          { key: "message.nonce", value: "66846", index: true },
          {
            key: "message.sender",
            value: "a463ad028fb79679cfc8ce1efba35ac0e77b35080a1abe9bebe83461f176b0a3",
            index: true,
          },
          { key: "message.sequence", value: "3417", index: true },
        ],
      },
    ],
    height: 10436798n,
    chain: "terra2",
    data: "Ei4KLC9jb3Ntd2FzbS53YXNtLnYxLk1zZ0V4ZWN1dGVDb250cmFjdFJlc3BvbnNlEi4KLC9jb3Ntd2FzbS53YXNtLnYxLk1zZ0V4ZWN1dGVDb250cmFjdFJlc3BvbnNl",
    hash: "F860EA95AC350D50354290349B2027EA01DE64AAA144381E22FBB97EE6FFA215",
    tx: Buffer.from([
      10, 236, 5, 10, 171, 2, 10, 36, 47, 99, 111, 115, 109, 119, 97, 115, 109, 46, 119, 97, 115,
      109, 46, 118, 49, 46, 77, 115, 103, 69, 120, 101, 99, 117, 116, 101, 67, 111, 110, 116, 114,
      97, 99, 116, 18, 130, 2, 10, 44, 116, 101, 114, 114, 97, 49, 107, 103, 115, 56, 108, 100, 57,
      119, 121, 118, 121, 101, 100, 106, 97, 118, 121, 54, 50, 107, 100, 48, 53, 118, 114, 102, 54,
      100, 107, 117, 114, 101, 104, 103, 108, 50, 118, 103, 18, 64, 116, 101, 114, 114, 97, 49, 99,
      116, 101, 108, 119, 97, 121, 107, 54, 116, 50, 122, 117, 51, 48, 97, 56, 118, 57, 107, 100,
      103, 51, 117, 50, 103, 114, 48, 115, 108, 112, 106, 100, 102, 110, 121, 53, 112, 106, 112, 55,
      109, 51, 116, 117, 113, 117, 107, 51, 50, 121, 115, 117, 103, 121, 106, 100, 103, 26, 143, 1,
      123, 34, 105, 110, 99, 114, 101, 97, 115, 101, 95, 97, 108, 108, 111, 119, 97, 110, 99, 101,
      34, 58, 123, 34, 97, 109, 111, 117, 110, 116, 34, 58, 34, 49, 48, 48, 48, 48, 48, 48, 34, 44,
      34, 101, 120, 112, 105, 114, 101, 115, 34, 58, 123, 34, 110, 101, 118, 101, 114, 34, 58, 123,
      125, 125, 44, 34, 115, 112, 101, 110, 100, 101, 114, 34, 58, 34, 116, 101, 114, 114, 97, 49,
      53, 51, 51, 54, 54, 113, 53, 48, 107, 55, 116, 56, 110, 110, 55, 103, 101, 99, 48, 48, 104,
      103, 54, 54, 99, 114, 110, 104, 107, 100, 103, 103, 112, 103, 100, 116, 97, 120, 108, 116, 97,
      113, 54, 120, 114, 117, 116, 107, 107, 122, 51, 115, 57, 57, 50, 102, 119, 57, 34, 125, 125,
      10, 157, 3, 10, 36, 47, 99, 111, 115, 109, 119, 97, 115, 109, 46, 119, 97, 115, 109, 46, 118,
      49, 46, 77, 115, 103, 69, 120, 101, 99, 117, 116, 101, 67, 111, 110, 116, 114, 97, 99, 116,
      18, 244, 2, 10, 44, 116, 101, 114, 114, 97, 49, 107, 103, 115, 56, 108, 100, 57, 119, 121,
      118, 121, 101, 100, 106, 97, 118, 121, 54, 50, 107, 100, 48, 53, 118, 114, 102, 54, 100, 107,
      117, 114, 101, 104, 103, 108, 50, 118, 103, 18, 64, 116, 101, 114, 114, 97, 49, 53, 51, 51,
      54, 54, 113, 53, 48, 107, 55, 116, 56, 110, 110, 55, 103, 101, 99, 48, 48, 104, 103, 54, 54,
      99, 114, 110, 104, 107, 100, 103, 103, 112, 103, 100, 116, 97, 120, 108, 116, 97, 113, 54,
      120, 114, 117, 116, 107, 107, 122, 51, 115, 57, 57, 50, 102, 119, 57, 26, 129, 2, 123, 34,
      105, 110, 105, 116, 105, 97, 116, 101, 95, 116, 114, 97, 110, 115, 102, 101, 114, 34, 58, 123,
      34, 97, 115, 115, 101, 116, 34, 58, 123, 34, 97, 109, 111, 117, 110, 116, 34, 58, 34, 49, 48,
      48, 48, 48, 48, 48, 34, 44, 34, 105, 110, 102, 111, 34, 58, 123, 34, 116, 111, 107, 101, 110,
      34, 58, 123, 34, 99, 111, 110, 116, 114, 97, 99, 116, 95, 97, 100, 100, 114, 34, 58, 34, 116,
      101, 114, 114, 97, 49, 99, 116, 101, 108, 119, 97, 121, 107, 54, 116, 50, 122, 117, 51, 48,
      97, 56, 118, 57, 107, 100, 103, 51, 117, 50, 103, 114, 48, 115, 108, 112, 106, 100, 102, 110,
      121, 53, 112, 106, 112, 55, 109, 51, 116, 117, 113, 117, 107, 51, 50, 121, 115, 117, 103, 121,
      106, 100, 103, 34, 125, 125, 125, 44, 34, 102, 101, 101, 34, 58, 34, 48, 34, 44, 34, 110, 111,
      110, 99, 101, 34, 58, 54, 54, 56, 52, 54, 44, 34, 114, 101, 99, 105, 112, 105, 101, 110, 116,
      34, 58, 34, 76, 105, 102, 119, 83, 55, 97, 67, 51, 107, 81, 80, 111, 68, 107, 103, 53, 89, 73,
      53, 85, 55, 80, 115, 104, 113, 86, 55, 114, 108, 89, 54, 47, 77, 81, 102, 100, 84, 55, 52, 47,
      82, 103, 61, 34, 44, 34, 114, 101, 99, 105, 112, 105, 101, 110, 116, 95, 99, 104, 97, 105,
      110, 34, 58, 49, 125, 125, 18, 28, 87, 111, 114, 109, 104, 111, 108, 101, 32, 45, 32, 73, 110,
      105, 116, 105, 97, 116, 101, 32, 84, 114, 97, 110, 115, 102, 101, 114, 18, 104, 10, 80, 10,
      70, 10, 31, 47, 99, 111, 115, 109, 111, 115, 46, 99, 114, 121, 112, 116, 111, 46, 115, 101,
      99, 112, 50, 53, 54, 107, 49, 46, 80, 117, 98, 75, 101, 121, 18, 35, 10, 33, 2, 249, 152, 102,
      64, 163, 68, 87, 197, 197, 20, 195, 154, 182, 195, 142, 82, 48, 25, 31, 214, 249, 122, 129,
      39, 78, 157, 117, 126, 105, 234, 166, 249, 18, 4, 10, 2, 8, 1, 24, 13, 18, 20, 10, 14, 10, 5,
      117, 108, 117, 110, 97, 18, 5, 49, 49, 51, 53, 48, 16, 176, 151, 46, 26, 64, 239, 113, 42, 77,
      98, 182, 139, 207, 184, 26, 94, 160, 141, 26, 94, 166, 138, 167, 80, 172, 31, 200, 202, 129,
      225, 97, 87, 160, 78, 30, 138, 41, 25, 185, 155, 165, 92, 49, 160, 157, 142, 194, 101, 95, 66,
      24, 13, 117, 169, 175, 5, 254, 139, 157, 104, 229, 217, 39, 255, 22, 204, 167, 51, 104,
    ]),
    timestamp: "1716388503458",
  };
});
