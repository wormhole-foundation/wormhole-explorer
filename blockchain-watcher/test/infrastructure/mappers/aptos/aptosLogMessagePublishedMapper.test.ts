import { describe, it, expect } from "@jest/globals";
import { aptosLogMessagePublishedMapper } from "../../../../src/infrastructure/mappers/aptos/aptosLogMessagePublishedMapper";
import { TransactionsByVersion } from "../../../../src/infrastructure/repositories/aptos/AptosJsonRPCBlockRepository";

describe("aptosLogMessagePublishedMapper", () => {
  it("should be able to map log to aptosLogMessagePublishedMapper", async () => {
    const result = aptosLogMessagePublishedMapper(txs);
    if (result) {
      expect(result.name).toBe("log-message-published");
      expect(result.chainId).toBe(22);
      expect(result.txHash).toBe(
        "0x99f9cd1ea181d568ba4d89e414dcf1b129968b1c805388f29821599a447b7741"
      );
      expect(result.address).toBe(
        "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625"
      );
      expect(result.attributes.consistencyLevel).toBe(0);
      expect(result.attributes.nonce).toBe(76704);
      expect(result.attributes.payload).toBe(
        "0x01000000000000000000000000000000000000000000000000000000000097d3650000000000000000000000003c499c542cef5e3811e1192ce70d8cc03d5c3359000500000000000000000000000081c1980abe8971e14865a629dd75b07621db1ae100050000000000000000000000000000000000000000000000000000000000001fe0"
      );
      expect(result.attributes.sender).toBe(
        "0x5aa807666de4dd9901c8f14a2f6021e85dc792890ece7f6bb929b46dba7671a2"
      );
      expect(result.attributes.sequence).toBe(34);
    }
  });
});

const txs: TransactionsByVersion = {
  consistencyLevel: 0,
  blockHeight: 153517771n,
  timestamp: "1709638693443328",
  blockTime: 1709638693443328,
  sequence: "34",
  version: "482649547",
  payload:
    "0x01000000000000000000000000000000000000000000000000000000000097d3650000000000000000000000003c499c542cef5e3811e1192ce70d8cc03d5c3359000500000000000000000000000081c1980abe8971e14865a629dd75b07621db1ae100050000000000000000000000000000000000000000000000000000000000001fe0",
  address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
  sender: "0x5aa807666de4dd9901c8f14a2f6021e85dc792890ece7f6bb929b46dba7671a2",
  status: true,
  events: [
    {
      guid: {
        creation_number: "11",
        account_address: "0x5aa807666de4dd9901c8f14a2f6021e85dc792890ece7f6bb929b46dba7671a2",
      },
      sequence_number: "0",
      type: "0x1::coin::WithdrawEvent",
      data: { amount: "9950053" },
    },
    {
      guid: {
        creation_number: "3",
        account_address: "0x5aa807666de4dd9901c8f14a2f6021e85dc792890ece7f6bb929b46dba7671a2",
      },
      sequence_number: "16",
      type: "0x1::coin::WithdrawEvent",
      data: { amount: "0" },
    },
    {
      guid: {
        creation_number: "4",
        account_address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
      },
      sequence_number: "149041",
      type: "0x1::coin::DepositEvent",
      data: { amount: "0" },
    },
    {
      guid: {
        creation_number: "2",
        account_address: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625",
      },
      sequence_number: "149040",
      type: "0x5bc11445584a763c1fa7ed39081f1b920954da14e04b32440cba863d03e19625::state::WormholeMessage",
      data: {
        consistency_level: 0,
        nonce: "76704",
        payload:
          "0x01000000000000000000000000000000000000000000000000000000000097d3650000000000000000000000003c499c542cef5e3811e1192ce70d8cc03d5c3359000500000000000000000000000081c1980abe8971e14865a629dd75b07621db1ae100050000000000000000000000000000000000000000000000000000000000001fe0",
        sender: "1",
        sequence: "146094",
        timestamp: "1709638693",
      },
    },
    {
      guid: { creation_number: "0", account_address: "0x0" },
      sequence_number: "0",
      type: "0x1::transaction_fee::FeeStatement",
      data: {
        execution_gas_units: "7",
        io_gas_units: "12",
        storage_fee_octas: "0",
        storage_fee_refund_octas: "0",
        total_charge_gas_units: "19",
      },
    },
  ],
  nonce: "76704",
  hash: "0xb2fa774485ce02c5786475dd2d689c3e3c2d0df0c5e09a1c8d1d0e249d96d76e",
};
