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
        "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f"
      );
      expect(result.attributes.emitterChain).toBe(21);
      expect(result.attributes.emitterAddress).toBe(
        "ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5"
      );
      expect(result.attributes.sequence).toBe(111883);
      expect(result.attributes.status).toBe("completed");
      expect(result.attributes.protocol).toBe("Token Bridge");
    }
  });

  it("should not be able to map log to aptosRedeemedTransactionFoundMapper", async () => {
    // Given
    const tx: AptosTransaction = {
      blockHeight: 154363203n,
      timestamp: 1709821894,
      blockTime: 1709821894,
      version: "487572005",
      payload: {
        function:
          "0xb1421c3d524a353411aa4e3cce0f0ce7f404a12da91a2889e1bc3cea6ffb17da::cancel_and_place::cancel_and_place",
        type_arguments: [
          "0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::WETH",
          "0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::USDC",
        ],
        arguments: [
          "0x01000000030d010f845d9923156821372114d23c633a86d8651cf79524764555dc9711631de5a4623f6cec4127b1e6872532bd24c148e326d324c09e413dc0ca222ec3dbc56abd00035fe16475cbf0d4f1a07f76dffd9690dc56fdeee3dce8f15f0dc4c71e32964ef72cb8b3b41c042912622dfb0de69e92dc77387f51659fdf85d7783b9cc8c0662d0104f982da1feb8b0b6a33ab69e395bda21dc1125fa1bae8386174c989f921f2bd71794ebaca92fed4c92be13cc031af8c27678b65ae2a2bf891b689ab15050bf62b000696f6ec7f68b7a0bf659f541a02e98a8c6e054b9ed7c3b654d97e3db678efa0f500465d833d307e23da40cbbc52e42457bb1d52e36a5b44b557cc84a970ae4ee40107097002812bea42577b5ba3c570a4b7f580a89144d2941aa699787eeaebaffb027d7ba89fd01ce2d635356ba63a7b25fd880b6e033a69717641ceabeb0fd2df2a0008e0faeafb1e8b2bd0dcaf807e7185ec2c425725e53cf933f713ac3e1461a712400b004d913c07d97d5fa2e7aed8a42cf3f3a7bca1fa74978653dcb235a5ba8eb60109610dd1508c204ff306584dba43a8bd86bc7c9fabefae661f367d4a01fd82bdfb5e5ed9a72cc4b0a5d12f6a47de5e6b66c3498afbb3b242012355ae5a1658a0ab010a7bb4707ee2af704ac0de725bc7c3ba9f776a595244d21203a05211cdfbfc550e5281a16d3f9f5d691a4b26a11af22a682b8f317c2cda55bde9272d2d68f9ed2e010c85d6087988bed609680b2047820bf14fca38001d2702607fe5f50a05a2b9533f2460be246d1c249e90aac396b06aaecb90304b9c4748d3ff6280a48338f7f4fc000df65098fbae3333d0dbaef9b10c50f2ec0d395e899d939c1a91363fa570d711e955d20eec3c7fba59f8198ad2ef5746a9a21ff69441c2b460ad4e8a55f3fb21450010493b06f48838dcf965e08bd266473d4275c5b3b73e70ca03fbeb198f5d610e955c6d11f16b9e592e0f85fe6d1f7f82118e6933541bb90dd20ad41c884a0308a4011123515dc5e8222adec3c31130bd245e43a4a7b435bb531027f7632fcd23607a92592d0769557558686f5cc8ba606096ea181829bc0d54e36a4d249a59f74034e50012db6bf0b9aad11fbb5959128a1103dc5d948d6ede23359626700d1b2c6dff9be1137888f72bd0b79fd7c36e511c05c1475a837954948e313b981e6c56123d1d270165f0be060000410f0015ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5000000000001b50b00010000000000000000000000000000000000000000000000000000000004d13720000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48000234226de67ee40839dff5760c2725bd2d5d7817fea499126c533e69f5702c5a7d00160000000000000000000000000000000000000000000000000000000000000000",
          ["42645072269700382266349876", "42644869357063623703648693"],
          ["42645090718696431883245683", "42644906249285249691670581"],
          ["2517", "10041"],
          ["383120", "383216"],
          ["2539", "10008"],
          ["382928", "382832"],
          "0x50bc83f01d48ab3b9c00048542332201ab9cbbea61bda5f48bf81dc506caa78f",
          3,
          0,
        ],
        type: "entry_function_payload",
      },
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
  blockHeight: 154363203n,
  timestamp: 1709821894,
  blockTime: 1709821894,
  version: "487572005",
  payload: {
    function:
      "0x576410486a2da45eee6c949c995670112ddf2fbeedab20350d506328eefc9d4f::complete_transfer::submit_vaa_and_register_entry",
    type_arguments: ["0x5e156f1207d0ebfa19a9eeff00d62a282278fb8719f4fab3a586a0a2c0fffbea::coin::T"],
    arguments: [
      "0x01000000030d010f845d9923156821372114d23c633a86d8651cf79524764555dc9711631de5a4623f6cec4127b1e6872532bd24c148e326d324c09e413dc0ca222ec3dbc56abd00035fe16475cbf0d4f1a07f76dffd9690dc56fdeee3dce8f15f0dc4c71e32964ef72cb8b3b41c042912622dfb0de69e92dc77387f51659fdf85d7783b9cc8c0662d0104f982da1feb8b0b6a33ab69e395bda21dc1125fa1bae8386174c989f921f2bd71794ebaca92fed4c92be13cc031af8c27678b65ae2a2bf891b689ab15050bf62b000696f6ec7f68b7a0bf659f541a02e98a8c6e054b9ed7c3b654d97e3db678efa0f500465d833d307e23da40cbbc52e42457bb1d52e36a5b44b557cc84a970ae4ee40107097002812bea42577b5ba3c570a4b7f580a89144d2941aa699787eeaebaffb027d7ba89fd01ce2d635356ba63a7b25fd880b6e033a69717641ceabeb0fd2df2a0008e0faeafb1e8b2bd0dcaf807e7185ec2c425725e53cf933f713ac3e1461a712400b004d913c07d97d5fa2e7aed8a42cf3f3a7bca1fa74978653dcb235a5ba8eb60109610dd1508c204ff306584dba43a8bd86bc7c9fabefae661f367d4a01fd82bdfb5e5ed9a72cc4b0a5d12f6a47de5e6b66c3498afbb3b242012355ae5a1658a0ab010a7bb4707ee2af704ac0de725bc7c3ba9f776a595244d21203a05211cdfbfc550e5281a16d3f9f5d691a4b26a11af22a682b8f317c2cda55bde9272d2d68f9ed2e010c85d6087988bed609680b2047820bf14fca38001d2702607fe5f50a05a2b9533f2460be246d1c249e90aac396b06aaecb90304b9c4748d3ff6280a48338f7f4fc000df65098fbae3333d0dbaef9b10c50f2ec0d395e899d939c1a91363fa570d711e955d20eec3c7fba59f8198ad2ef5746a9a21ff69441c2b460ad4e8a55f3fb21450010493b06f48838dcf965e08bd266473d4275c5b3b73e70ca03fbeb198f5d610e955c6d11f16b9e592e0f85fe6d1f7f82118e6933541bb90dd20ad41c884a0308a4011123515dc5e8222adec3c31130bd245e43a4a7b435bb531027f7632fcd23607a92592d0769557558686f5cc8ba606096ea181829bc0d54e36a4d249a59f74034e50012db6bf0b9aad11fbb5959128a1103dc5d948d6ede23359626700d1b2c6dff9be1137888f72bd0b79fd7c36e511c05c1475a837954948e313b981e6c56123d1d270165f0be060000410f0015ccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5000000000001b50b00010000000000000000000000000000000000000000000000000000000004d13720000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48000234226de67ee40839dff5760c2725bd2d5d7817fea499126c533e69f5702c5a7d00160000000000000000000000000000000000000000000000000000000000000000",
    ],
    type: "entry_function_payload",
  },

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
