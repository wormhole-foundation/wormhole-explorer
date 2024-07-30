import { evmLogMessageSentMapper } from "../../../../src/infrastructure/mappers/evm/evmLogMessageSentMapper";
import { HandleEvmTransactions } from "../../../../src/domain/actions";
import { describe, it, expect } from "@jest/globals";

const topic = "0x6eb224fb001ed210e379b335e35efe88672a8ce935d981a6896b27ffdf52a3b2";
const txHash = "0xcbdefc83080a8f60cbde7785eb2978548fd5c1f7d0ea2c024cce537845d339c7";

let statsRepo = {
  count: () => {},
  measure: () => {},
  report: () => Promise.resolve(""),
};

const handler = new HandleEvmTransactions(
  {
    abi: "event MessageSent (bytes message)",
    metricName: "process_message_sent_event",
    commitment: "latest",
    environment: "testnet",
    chainId: 2,
    chain: "ethereum",
    id: "poll-log-message-sent-ethereum",
  },
  evmLogMessageSentMapper,
  async () => {},
  statsRepo
);

describe("evmLogMessageSentMapper", () => {
  it("should be able to map log to messageSent", async () => {
    const [result] = await handler.handle([
      {
        blockHash: "0xb4499131ea4de775e1ac604003d18edbcb757ebc8b147d5b781d70e6ae05ab8f",
        from: "0x07ae8551be970cb1cca11dd7a11f47ae82e70e67",
        gas: "0x2f89b",
        blockNumber: 18521386n,
        gasPrice: "0x21cfa6db7",
        maxPriorityFeePerGas: "0x6b49d200",
        maxFeePerGas: "0x7686d9329",
        hash: "0x7c9c7866df17ce30bc2086498752912d5e0ea1d1fac32d567509a0aef555a3b7",
        input:
          "0x6fd3504e00000000000000000000000000000000000000000000000000000001b774c276000000000000000000000000000000000000000000000000000000000000000200000000000000000000000007ae8551be970cb1cca11dd7a11f47ae82e70e67000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
        nonce: "0xaf6",
        to: "0xbd3fa81b58ba92a82136038b25adec7066af3155",
        transactionIndex: "0x39",
        value: "0x0",
        type: "0x2",
        chainId: 2,
        v: "0x0",
        r: "0x9bc730ed55c2d008945381ab65fb51575837237fdb21cc12e04c13e11709601e",
        s: "0x5a5ba869e0908fc2def6e7782b52cb23f5268384386f4c4752c62680db116989",
        effectiveGasPrice: "0x21cfa6db7",
        gasUsed: "0x1a81a",
        timestamp: 1722340847,
        status: "0x1",
        logs: [
          {
            address: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
            topics: [
              "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
              "0x00000000000000000000000007ae8551be970cb1cca11dd7a11f47ae82e70e67",
              "0x000000000000000000000000c4922d64a24675e16e1586e3e3aa56c06fabe907",
            ],
            data: "0x00000000000000000000000000000000000000000000000000000001b774c276",
          },
          {
            address: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
            topics: [
              "0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5",
              "0x000000000000000000000000c4922d64a24675e16e1586e3e3aa56c06fabe907",
            ],
            data: "0x00000000000000000000000000000000000000000000000000000001b774c276",
          },
          {
            address: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
            topics: [
              "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
              "0x000000000000000000000000c4922d64a24675e16e1586e3e3aa56c06fabe907",
              "0x0000000000000000000000000000000000000000000000000000000000000000",
            ],
            data: "0x00000000000000000000000000000000000000000000000000000001b774c276",
          },
          {
            address: "0x0a992d191deec32afe36203ad87d7d289a738f81",
            topics: ["0x8c5261668696ce22758910d05bab8f186d6eb247ceac2af2e82c7dc17669b036"],
            data: "0x000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000f80000000000000000000000020000000000015bda000000000000000000000000bd3fa81b58ba92a82136038b25adec7066af31550000000000000000000000002b4069517957735be00cee0fadae88a26365528f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb4800000000000000000000000007ae8551be970cb1cca11dd7a11f47ae82e70e6700000000000000000000000000000000000000000000000000000001b774c27600000000000000000000000007ae8551be970cb1cca11dd7a11f47ae82e70e670000000000000000",
          },
          {
            address: "0xbd3fa81b58ba92a82136038b25adec7066af3155",
            topics: [
              "0x2fa9ca894982930190727e75500a97d8dc500233a5065e0f3126c48fbe0343c0",
              "0x0000000000000000000000000000000000000000000000000000000000015bda",
              "0x000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
              "0x00000000000000000000000007ae8551be970cb1cca11dd7a11f47ae82e70e67",
            ],
            data: "0x00000000000000000000000000000000000000000000000000000001b774c27600000000000000000000000007ae8551be970cb1cca11dd7a11f47ae82e70e6700000000000000000000000000000000000000000000000000000000000000020000000000000000000000002b4069517957735be00cee0fadae88a26365528f0000000000000000000000000000000000000000000000000000000000000000",
          },
        ],
        environment: "mainnet",
        chain: "ethereum",
      },
    ]);

    expect(result!.blockHeight).toBe(18521386n);
    expect(result!.chainId).toBe(2);
    expect(result!.txHash).toBe(
      "0x7c9c7866df17ce30bc2086498752912d5e0ea1d1fac32d567509a0aef555a3b7"
    );
    expect(result!.name).toBe("message-sent");
  });
});
