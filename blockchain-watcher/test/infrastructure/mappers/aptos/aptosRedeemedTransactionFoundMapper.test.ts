import { aptosRedeemedTransactionFoundMapper } from "../../../../src/infrastructure/mappers/aptos/aptosRedeemedTransactionFoundMapper";
import { describe, it, expect } from "@jest/globals";
import { AptosTransaction } from "../../../../src/domain/entities/aptos";

describe("aptosRedeemedTransactionFoundMapper", () => {
  it("should be able to map log to aptosRedeemedTransactionFoundMapper", async () => {
    // When
    const result = aptosRedeemedTransactionFoundMapper(tx);

    if (result) {
      // Then
      expect(result.name).toBe("transfer-redeemed");
      expect(result.chainId).toBe(22);
      expect(result.txHash).toBe(
        "0xd297f46372ed734fd8f3b595a898f42ab328a4e3cc9ce0f810c6c66e64873e64"
      );
      expect(result.address).toBe(
        "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f"
      );
      expect(result.attributes.from).toBe(
        "0000000000000000000000005a58505a96d1dbf8df91cb21b54419fc36e93fde"
      );
      expect(result.attributes.emitterChain).toBe(5);
      expect(result.attributes.emitterAddress).toBe(
        "0000000000000000000000005a58505a96d1dbf8df91cb21b54419fc36e93fde"
      );
      expect(result.attributes.sequence).toBe(394768);
      expect(result.attributes.status).toBe("completed");
      expect(result.attributes.protocol).toBe("Token Bridge");
    }
  });

  it("should not be able to map log to aptosRedeemedTransactionFoundMapper", async () => {
    // Given
    const tx: AptosTransaction = {
      consistencyLevel: 15,
      emitterChain: 5,
      blockHeight: 154363203n,
      timestamp: 1709821894,
      blockTime: 1709821894,
      sequence: 394768n,
      version: "487572005",
      payload:
        "010000000000000000000000000000000000000000000000000000000000989680069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f0000000000100019cb6a1e8b0e7104e988b8d5928d58f79995b7d8832a873017bfc2038037768ea00160000000000000000000000000000000000000000000000000000000000000000",
      address: "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
      sender: "0000000000000000000000005a58505a96d1dbf8df91cb21b54419fc36e93fde",
      status: true,
      events: [],
      nonce: 302448640,
      hash: "0xd297f46372ed734fd8f3b595a898f42ab328a4e3cc9ce0f810c6c66e64873e64",
      type: "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f::complete_transfer::cancel",
    };
    // When
    const result = aptosRedeemedTransactionFoundMapper(tx);

    // Then
    expect(result).toBeUndefined();
  });
});

const tx: AptosTransaction = {
  consistencyLevel: 15,
  emitterChain: 5,
  blockHeight: 154363203n,
  timestamp: 1709821894,
  blockTime: 1709821894,
  sequence: 394768n,
  version: "487572005",
  payload:
    "010000000000000000000000000000000000000000000000000000000000989680069b8857feab8184fb687f634618c035dac439dc1aeb3b5598a0f0000000000100019cb6a1e8b0e7104e988b8d5928d58f79995b7d8832a873017bfc2038037768ea00160000000000000000000000000000000000000000000000000000000000000000",
  address: "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f",
  sender: "0000000000000000000000005a58505a96d1dbf8df91cb21b54419fc36e93fde",
  status: true,
  events: [
    {
      guid: {
        creation_number: "4",
        account_address: "0x9cb6a1e8b0e7104e988b8d5928d58f79995b7d8832a873017bfc2038037768ea",
      },
      sequence_number: "4",
      type: "0x1::coin::DepositEvent",
      data: { amount: "10000000" },
    },
    {
      guid: {
        creation_number: "4",
        account_address: "0x9cb6a1e8b0e7104e988b8d5928d58f79995b7d8832a873017bfc2038037768ea",
      },
      sequence_number: "5",
      type: "0x1::coin::DepositEvent",
      data: { amount: "0" },
    },
    {
      guid: { creation_number: "0", account_address: "0x0" },
      sequence_number: "0",
      type: "0x1::transaction_fee::FeeStatement",
      data: {
        execution_gas_units: "118",
        io_gas_units: "8",
        storage_fee_octas: "62820",
        storage_fee_refund_octas: "0",
        total_charge_gas_units: "753",
      },
    },
  ],
  nonce: 302448640,
  hash: "0xd297f46372ed734fd8f3b595a898f42ab328a4e3cc9ce0f810c6c66e64873e64",
  type: "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f::complete_transfer::submit_vaa_and_register_entry",
};
