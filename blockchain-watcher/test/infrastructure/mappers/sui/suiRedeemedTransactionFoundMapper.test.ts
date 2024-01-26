import { describe, expect, it } from "@jest/globals";
import { SuiTransactionBlockReceipt } from "../../../../src/domain/entities/sui";
import { suiRedeemedTransactionFoundMapper } from "../../../../src/infrastructure/mappers/sui/suiRedeemedTransactionFoundMapper";

let redeemTx: SuiTransactionBlockReceipt;

describe("suiRedeemedTransactionFoundMapper", () => {
  it("should map a transaction block receipt", () => {
    givenARedeemTransaction();

    const result = suiRedeemedTransactionFoundMapper(redeemTx);

    expect(result).not.toBeFalsy();
    expect(result?.name).toEqual("transfer-redeemed");
    expect(result?.chainId).toEqual(21);
    expect(result?.address).toEqual("0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d");
    expect(result?.txHash).toEqual("C6T5fBM636j4y34kt4Rowor3qvMaNZ8QmT9n1TRzkHX1");
    expect(result?.blockHeight).toEqual(24408383n);
    expect(result?.blockTime).toEqual(1706107525474);
    expect(result?.attributes.from).toEqual("0xfcda48b391b8a1c6a9e57f30247bc0d5a97595f4a61784078ad17e11c2a8d529");
    expect(result?.attributes.emitterChain).toEqual(1);
    expect(result?.attributes.emitterAddress).toEqual('ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5');
    expect(result?.attributes.sequence).toEqual(610687);
    expect(result?.attributes.status).toEqual('completed');
  });
});

const givenARedeemTransaction = () => {
  redeemTx = {
    digest: "C6T5fBM636j4y34kt4Rowor3qvMaNZ8QmT9n1TRzkHX1",
    transaction: {
      data: {
        messageVersion: "v1",
        transaction: {
          kind: "ProgrammableTransaction",
          inputs: [
            {
              type: "object",
              objectType: "sharedObject",
              objectId: "0xaeab97f96cf9877fee2883315d459552b2b921edc16d7ceac6eab944dd88919c",
              initialSharedVersion: "64",
              mutable: false,
            },
            {
              type: "pure",
              valueType: "vector<u8>",
              value: [
                1, 0, 0, 0, 3, 13, 1, 165, 59, 175, 174, 115, 81, 195, 166, 230, 125, 166, 74, 224,
                174, 203, 174, 198, 43, 219, 165, 56, 248, 74, 140, 175, 28, 252, 153, 66, 89, 192,
                224, 54, 49, 55, 108, 236, 221, 132, 4, 225, 157, 193, 157, 166, 203, 177, 162, 214,
                119, 115, 96, 213, 179, 131, 115, 193, 110, 85, 131, 3, 44, 232, 207, 0, 2, 17, 215,
                240, 139, 132, 12, 110, 218, 66, 120, 190, 57, 118, 97, 252, 140, 54, 154, 16, 110,
                175, 71, 124, 53, 60, 33, 78, 58, 24, 14, 20, 72, 111, 93, 225, 28, 246, 198, 72, 245,
                232, 93, 254, 249, 103, 30, 246, 185, 52, 12, 138, 234, 15, 136, 30, 126, 118, 167,
                75, 196, 101, 152, 61, 206, 0, 4, 112, 16, 206, 142, 7, 212, 94, 50, 7, 116, 139, 103,
                192, 238, 221, 227, 86, 158, 200, 127, 148, 39, 148, 213, 241, 95, 151, 155, 107, 239,
                254, 55, 20, 136, 65, 145, 96, 99, 70, 5, 98, 139, 242, 13, 44, 21, 27, 136, 113, 245,
                120, 70, 244, 155, 41, 255, 127, 75, 54, 153, 239, 92, 77, 235, 1, 6, 45, 225, 36, 74,
                90, 77, 109, 139, 143, 93, 136, 55, 213, 102, 85, 235, 197, 122, 54, 218, 80, 20, 133,
                121, 239, 125, 13, 170, 143, 123, 11, 51, 77, 99, 32, 53, 106, 76, 133, 84, 206, 183,
                194, 146, 151, 122, 94, 20, 76, 182, 154, 183, 219, 25, 101, 233, 253, 143, 183, 94,
                178, 110, 242, 213, 0, 8, 95, 239, 146, 102, 26, 246, 37, 252, 95, 238, 250, 117, 119,
                223, 193, 66, 208, 161, 173, 1, 110, 52, 249, 104, 120, 98, 30, 4, 113, 162, 185, 99,
                98, 190, 161, 110, 55, 224, 187, 156, 201, 254, 66, 1, 52, 159, 106, 219, 215, 53,
                105, 210, 21, 35, 37, 162, 47, 61, 231, 148, 220, 59, 79, 78, 1, 9, 89, 13, 110, 19,
                159, 42, 37, 55, 207, 74, 154, 167, 25, 15, 207, 194, 225, 190, 54, 253, 172, 199,
                216, 22, 42, 224, 94, 178, 125, 97, 105, 164, 88, 171, 235, 182, 82, 103, 145, 12,
                171, 168, 25, 191, 27, 97, 1, 66, 77, 158, 64, 92, 42, 150, 66, 121, 52, 10, 77, 181,
                138, 23, 152, 123, 1, 10, 247, 91, 93, 221, 128, 223, 179, 5, 60, 51, 182, 138, 124,
                40, 233, 63, 122, 117, 201, 37, 12, 189, 26, 82, 197, 157, 175, 183, 232, 128, 169,
                122, 56, 53, 252, 248, 103, 152, 71, 35, 199, 121, 14, 15, 207, 186, 210, 232, 31,
                242, 14, 143, 212, 190, 9, 3, 70, 76, 66, 145, 237, 177, 7, 174, 1, 11, 137, 186, 191,
                33, 229, 12, 105, 7, 105, 243, 136, 17, 133, 56, 88, 19, 10, 255, 223, 107, 101, 1, 0,
                242, 24, 85, 131, 151, 41, 75, 101, 120, 55, 177, 99, 239, 13, 247, 236, 52, 190, 121,
                126, 64, 47, 57, 26, 78, 2, 55, 182, 42, 224, 74, 95, 113, 236, 158, 243, 29, 54, 129,
                214, 111, 0, 14, 227, 138, 75, 76, 226, 86, 165, 142, 188, 140, 93, 69, 15, 140, 200,
                83, 196, 4, 76, 3, 239, 21, 108, 82, 200, 12, 180, 207, 40, 10, 172, 110, 73, 197,
                217, 149, 218, 51, 139, 129, 169, 146, 116, 155, 219, 215, 223, 34, 238, 218, 123,
                219, 95, 14, 248, 124, 140, 151, 237, 11, 151, 223, 170, 20, 1, 15, 183, 68, 194, 204,
                245, 184, 63, 166, 176, 31, 1, 240, 162, 121, 227, 2, 170, 161, 22, 193, 171, 87, 97,
                177, 11, 204, 133, 129, 68, 35, 145, 160, 30, 70, 60, 158, 109, 69, 199, 232, 31, 24,
                236, 26, 139, 74, 137, 224, 53, 42, 38, 234, 123, 176, 183, 37, 85, 93, 30, 133, 131,
                229, 134, 236, 1, 16, 118, 146, 140, 89, 255, 51, 180, 215, 47, 232, 34, 25, 102, 13,
                56, 131, 164, 9, 210, 233, 60, 181, 116, 83, 208, 175, 171, 186, 232, 123, 208, 51, 7,
                170, 87, 171, 105, 63, 248, 42, 15, 56, 156, 67, 116, 253, 21, 3, 202, 121, 10, 132,
                154, 168, 192, 254, 151, 28, 46, 185, 95, 193, 105, 8, 1, 17, 98, 243, 132, 175, 247,
                192, 47, 159, 210, 186, 136, 68, 80, 86, 206, 0, 198, 196, 18, 219, 121, 185, 136,
                254, 58, 196, 67, 101, 142, 96, 19, 20, 21, 5, 62, 58, 95, 185, 160, 119, 230, 203,
                59, 184, 183, 119, 1, 136, 199, 216, 198, 181, 243, 95, 231, 9, 231, 55, 52, 204, 240,
                27, 127, 112, 0, 18, 22, 100, 166, 124, 30, 16, 27, 252, 198, 92, 248, 14, 61, 168,
                191, 242, 163, 240, 112, 232, 221, 67, 47, 245, 171, 66, 156, 213, 124, 217, 106, 190,
                78, 149, 27, 148, 228, 53, 189, 224, 118, 165, 221, 85, 170, 179, 241, 140, 182, 55,
                31, 143, 65, 91, 191, 68, 153, 65, 169, 205, 104, 63, 110, 79, 1, 101, 177, 34, 75, 0,
                0, 246, 179, 0, 1, 236, 115, 114, 153, 93, 92, 200, 115, 35, 151, 251, 10, 211, 92, 1,
                33, 224, 234, 169, 13, 38, 248, 40, 165, 52, 202, 181, 67, 145, 179, 164, 245, 0, 0,
                0, 0, 0, 9, 81, 127, 32, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0, 0, 0, 52, 176, 197, 0, 198, 250, 122, 243, 190, 219, 173, 58, 61,
                101, 243, 106, 171, 201, 116, 49, 177, 187, 228, 194, 210, 246, 224, 228, 124, 166, 2,
                3, 69, 47, 93, 97, 0, 1, 252, 218, 72, 179, 145, 184, 161, 198, 169, 229, 127, 48, 36,
                123, 192, 213, 169, 117, 149, 244, 166, 23, 132, 7, 138, 209, 126, 17, 194, 168, 213,
                41, 0, 21, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                0, 0, 0, 0, 0, 0, 0,
              ],
            },
            {
              type: "object",
              objectType: "sharedObject",
              objectId: "0x0000000000000000000000000000000000000000000000000000000000000006",
              initialSharedVersion: "1",
              mutable: false,
            },
            {
              type: "object",
              objectType: "sharedObject",
              objectId: "0xc57508ee0d4595e5a8728974a4a93a787d38f339757230d441e895422c07aba9",
              initialSharedVersion: "66",
              mutable: true,
            },
          ],
          transactions: [
            {
              MoveCall: {
                package: "0x5306f64e312b581766351c07af79c72fcb1cd25147157fdc2f8ad76de9a3fb6a",
                module: "vaa",
                function: "parse_and_verify",
                arguments: [
                  {
                    Input: 0,
                  },
                  {
                    Input: 1,
                  },
                  {
                    Input: 2,
                  },
                ],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "vaa",
                function: "verify_only_once",
                arguments: [
                  {
                    Input: 3,
                  },
                  {
                    NestedResult: [0, 0],
                  },
                ],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "complete_transfer",
                function: "authorize_transfer",
                type_arguments: [
                  "0xb231fcda8bbddb31f2ef02e6161444aec64a514e2c89279584ac9806ce9cf037::coin::COIN",
                ],
                arguments: [
                  {
                    Input: 3,
                  },
                  {
                    NestedResult: [1, 0],
                  },
                ],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "complete_transfer",
                function: "redeem_relayer_payout",
                type_arguments: [
                  "0xb231fcda8bbddb31f2ef02e6161444aec64a514e2c89279584ac9806ce9cf037::coin::COIN",
                ],
                arguments: [
                  {
                    NestedResult: [2, 0],
                  },
                ],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "coin_utils",
                function: "return_nonzero",
                type_arguments: [
                  "0xb231fcda8bbddb31f2ef02e6161444aec64a514e2c89279584ac9806ce9cf037::coin::COIN",
                ],
                arguments: [
                  {
                    NestedResult: [3, 0],
                  },
                ],
              },
            },
          ],
        },
        sender: "0xfcda48b391b8a1c6a9e57f30247bc0d5a97595f4a61784078ad17e11c2a8d529",
        gasData: {
          payment: [
            {
              objectId: "0xf0405f2f6e2ef6f86762f973907faeb68e41f8a3fb00326bb58dea73701209f7",
              version: "63608463",
              digest: "CoLRg7oKMJ3T3p3X9yTWGpiNPJL3SvdkQJLbXetfVVYe",
            },
          ],
          owner: "0xfcda48b391b8a1c6a9e57f30247bc0d5a97595f4a61784078ad17e11c2a8d529",
          price: "750",
          budget: "7690416",
        },
      },
      txSignatures: [
        "AFEXiY7OGpEFtlU4YZ/K6IktslWzqRfSqP8e90V+Pqowv8emUBDg875hoDxLU+hRqPPShBDqfTy6IzOZeNSQpQrPyLwQgQtt1dF1BjINPA76mVZwoYzzB1KhblrBFvgl/Q==",
      ],
    },
    events: [
      {
        id: {
          txDigest: "C6T5fBM636j4y34kt4Rowor3qvMaNZ8QmT9n1TRzkHX1",
          eventSeq: "0",
        },
        packageId: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
        transactionModule: "complete_transfer",
        sender: "0xfcda48b391b8a1c6a9e57f30247bc0d5a97595f4a61784078ad17e11c2a8d529",
        type: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d::complete_transfer::TransferRedeemed",
        parsedJson: {
          emitter_address: {
            value: {
              data: [
                236, 115, 114, 153, 93, 92, 200, 115, 35, 151, 251, 10, 211, 92, 1, 33, 224, 234, 169,
                13, 38, 248, 40, 165, 52, 202, 181, 67, 145, 179, 164, 245,
              ],
            },
          },
          emitter_chain: 1,
          sequence: "610687",
        },
        bcs: "5GvwRFsvFeuD7MNApPR4TV4hcYNZHwXoKuAet47kpHZugnKQD1pDG6yk15",
      },
    ],
    timestampMs: "1706107525474",
    checkpoint: "24408383",
  };
};
