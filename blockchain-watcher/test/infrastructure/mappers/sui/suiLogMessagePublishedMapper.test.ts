import { suiLogMessagePublishedMapper } from "../../../../src/infrastructure/mappers/sui/suiLogMessagePublishedMapper";
import { SuiTransactionBlockReceipt } from "../../../../src/domain/entities/sui";
import { describe, expect, it } from "@jest/globals";

let sourceTx: SuiTransactionBlockReceipt;

describe("suiLogMessagePublishedMapper", () => {
  it("should map a transaction block receipt for a source event", () => {
    givenASourceTransaction();

    const result = suiLogMessagePublishedMapper(sourceTx);

    expect(result).not.toBeFalsy();
    expect(result?.name).toEqual("log-message-published");
    expect(result?.chainId).toEqual(21);
    expect(result?.address).toEqual(
      "0x5306f64e312b581766351c07af79c72fcb1cd25147157fdc2f8ad76de9a3fb6a"
    );
    expect(result?.txHash).toEqual("CgSJ3izrN278gW9zEeHon8nTkNUrpTiPR3UW5KCLZFAS");
    expect(result?.blockHeight).toEqual(26841750n);
    expect(result?.attributes.nonce).toEqual(63790);
    expect(result?.attributes.sender).toEqual(
      "0xccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5"
    );
    expect(result?.attributes.sequence).toEqual(104715);
    expect(result?.attributes.consistencyLevel).toEqual(0);
    expect(result?.attributes.payload).toEqual(
      "0100000000000000000000000000000000000000000000000000000000004c4b40000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48000200000000000000000000000057e173f3be02f5d436bcc07efac63a584d01e0a300040000000000000000000000000000000000000000000000000000000000000000"
    );
  });
});

it("should ignores a non source transaction", () => {
  givenANonSourceTransaction();

  const result = suiLogMessagePublishedMapper(sourceTx);

  expect(result).toBeUndefined();
});

const givenASourceTransaction = () => {
  sourceTx = {
    digest: "CgSJ3izrN278gW9zEeHon8nTkNUrpTiPR3UW5KCLZFAS",
    transaction: {
      data: {
        gasData: {
          payment: [
            {
              objectId: "0x1ad2872de3d42a31d9e00a0d905011d7c4430ed12c783465321e5a05ae2b9e33",
              version: "64933577",
              digest: "A294rJ6JVi5NWaDezZn7jL1K3syVDvaNPQrkTsyYFwbx",
            },
          ],
          owner: "0x8f96433805302fd06592a2a85fd01ee92861dd2f203b2aadfad9d6adcc9e024e",
          price: "749",
          budget: "2605244",
        },
        messageVersion: "v1",
        transaction: {
          kind: "ProgrammableTransaction",
          inputs: [
            {
              type: "object",
              objectType: "immOrOwnedObject",
              objectId: "0x526320c9422b39efa1f5e867757d92b6130d783345ef6a9dc91d630a429f1954",
              version: "70164784",
              digest: "61sZzAWDqkeAc6MTkKPGzHPeRSUe4iVkv9GZSvn4d8Fc",
            },
            {
              type: "object",
              objectType: "immOrOwnedObject",
              objectId: "0xfc1bb5aa69ddebe2bdb538f11a9c13497ec41009cbfce82ad50b0c98cc15a184",
              version: "70164784",
              digest: "9dugruLRgeasK7sSTzj1mdri3cLXS4Rz2AxbvFUWUUyd",
            },
            {
              type: "object",
              objectType: "immOrOwnedObject",
              objectId: "0xfd0d3ab4f0a61b397daf56b6fc06cb1cf13ef9434bb7a93e125a7ccd2dc45213",
              version: "70161224",
              digest: "3983cYhVcDSakbGFYC1TLx6aJoQHjjmNAK2D9q62mDAL",
            },
            { type: "pure", valueType: "u64", value: "5000000" },
            {
              type: "pure",
              valueType: "address",
              value: "0x39340b05c03ffdbdf372d77e188dd68e8f08e1ab4ed222eab5d2fcc020878e13",
            },
            {
              type: "pure",
              valueType: "address",
              value: "0x39340b05c03ffdbdf372d77e188dd68e8f08e1ab4ed222eab5d2fcc020878e13",
            },
            { type: "pure", valueType: "u64", value: "5000000" },
            { type: "pure", valueType: "u64", value: "0" },
            { type: "pure", valueType: "u64", value: "0" },
            {
              type: "object",
              objectType: "sharedObject",
              objectId: "0xc57508ee0d4595e5a8728974a4a93a787d38f339757230d441e895422c07aba9",
              initialSharedVersion: "66",
              mutable: true,
            },
            { type: "pure", valueType: "u16", value: 4 },
            {
              type: "pure",
              valueType: "vector<u8>",
              value: [
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 87, 225, 115, 243, 190, 2, 245, 212, 54, 188,
                192, 126, 250, 198, 58, 88, 77, 1, 224, 163,
              ],
            },
            { type: "pure", valueType: "u64", value: "0" },
            { type: "pure", valueType: "u32", value: 63790 },
            {
              type: "object",
              objectType: "sharedObject",
              objectId: "0xaeab97f96cf9877fee2883315d459552b2b921edc16d7ceac6eab944dd88919c",
              initialSharedVersion: "64",
              mutable: true,
            },
            {
              type: "object",
              objectType: "sharedObject",
              objectId: "0x0000000000000000000000000000000000000000000000000000000000000006",
              initialSharedVersion: "1",
              mutable: false,
            },
          ],
          transactions: [
            { MakeMoveVec: [null, [{ Input: 0 }, { Input: 1 }, { Input: 2 }]] },
            {
              MoveCall: {
                package: "0xf331b35dfa0a560eadbe14280054c70908a487755bab106b5d528562f67ba8d1",
                module: "utils",
                function: "split_coins_and_transfer_rest",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ Result: 0 }, { Input: 3 }, { Input: 4 }],
              },
            },
            { MakeMoveVec: [null, [{ Result: 1 }]] },
            {
              MoveCall: {
                package: "0xf331b35dfa0a560eadbe14280054c70908a487755bab106b5d528562f67ba8d1",
                module: "swap_aggregator",
                function: "initiate_swap",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ Result: 2 }, { Input: 6 }, { Input: 7 }, { Input: 5 }],
              },
            },
            { SplitCoins: ["GasCoin", [{ Input: 8 }]] },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "state",
                function: "verified_asset",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ Input: 9 }],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "transfer_tokens",
                function: "prepare_transfer",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [
                  { NestedResult: [5, 0] },
                  { Result: 3 },
                  { Input: 10 },
                  { Input: 11 },
                  { Input: 12 },
                  { Input: 13 },
                ],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "coin_utils",
                function: "return_nonzero",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ NestedResult: [6, 1] }],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "transfer_tokens",
                function: "transfer_tokens",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ Input: 9 }, { NestedResult: [6, 0] }],
              },
            },
            {
              MoveCall: {
                package: "0x5306f64e312b581766351c07af79c72fcb1cd25147157fdc2f8ad76de9a3fb6a",
                module: "publish_message",
                function: "publish_message",
                arguments: [
                  { Input: 14 },
                  { NestedResult: [4, 0] },
                  { NestedResult: [8, 0] },
                  { Input: 15 },
                ],
              },
            },
          ],
        },
        sender: "0x39340b05c03ffdbdf372d77e188dd68e8f08e1ab4ed222eab5d2fcc020878e13",
      },
      txSignatures: [
        "AC0E+rayE0fL8LbH644pU8br2Lfay8EsOfkuq2x/oXmIYI1gK1zjpRH58fIHpLUPylbLfzCGyVITLdCAHhGQtwB+OJOHLGpRJukyYnp+TUImmLMhwM8Cff+IQHrLjM7+hQ==",
      ],
    },
    effects: {
      status: { status: "success" },
    },
    events: [
      {
        id: { txDigest: "CgSJ3izrN278gW9zEeHon8nTkNUrpTiPR3UW5KCLZFAS", eventSeq: "0" },
        packageId: "0xf331b35dfa0a560eadbe14280054c70908a487755bab106b5d528562f67ba8d1",
        transactionModule: "swap_aggregator",
        sender: "0x39340b05c03ffdbdf372d77e188dd68e8f08e1ab4ed222eab5d2fcc020878e13",
        type: "0xf331b35dfa0a560eadbe14280054c70908a487755bab106b5d528562f67ba8d1::swap_aggregator::InitiateSwapEvent",
        parsedJson: {
          amount: "5000000",
          coin_type: {
            name: "5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
          },
          steps: "0",
        },
        bcs: "Ji2VvXjm7tZ2a7qdSke7xSq9jfpdLYQNcKRDxpgGjECjyLkoSHbCoK2zhDnRTSiwR2K3PoPbCjWKdHw1YVY9PgYQiRvUmrwDGay6yctQidjxNTuEdConUwdaHZF8ZY3",
      },
      {
        id: { txDigest: "CgSJ3izrN278gW9zEeHon8nTkNUrpTiPR3UW5KCLZFAS", eventSeq: "1" },
        packageId: "0x5306f64e312b581766351c07af79c72fcb1cd25147157fdc2f8ad76de9a3fb6a",
        transactionModule: "publish_message",
        sender: "0x39340b05c03ffdbdf372d77e188dd68e8f08e1ab4ed222eab5d2fcc020878e13",
        type: "0x5306f64e312b581766351c07af79c72fcb1cd25147157fdc2f8ad76de9a3fb6a::publish_message::WormholeMessage",
        parsedJson: {
          consistency_level: 0,
          nonce: 63790,
          payload: [
            1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
            0, 76, 75, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 160, 184, 105, 145, 198, 33, 139, 54,
            193, 209, 157, 74, 46, 158, 176, 206, 54, 6, 235, 72, 0, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0,
            0, 0, 0, 87, 225, 115, 243, 190, 2, 245, 212, 54, 188, 192, 126, 250, 198, 58, 88, 77,
            1, 224, 163, 0, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
            0, 0, 0, 0, 0, 0, 0, 0, 0,
          ],
          sender: "0xccceeb29348f71bdd22ffef43a2a19c1f5b5e17c5cca5411529120182672ade5",
          sequence: "104715",
          timestamp: "1708526724",
        },
        bcs: "HPnaCsDcXUeB8SdsicJc7RuEYknvcKXrcxiRq7DLh9Uzug3iHiQ7ZXJu8BmdvgXGReVY5QMx8dH9CoUQTGpSdjp6Nddu28n5HfPvMoQqTwD1tgWwF8GAJ9ott29LKrRSfHJsJbA7ekw6uisZH12pMEvDKZwyHf1zaPwbDzHNTKNASMW6FTSHqWVL2T3p2Zn127qCpQhWiGMomUAf9sUFC2Qchs4zZFmCZXZKXNnBVeSE6dEvuGPQ19TnU11NGXpWK",
      },
    ],
    timestampMs: "1708526724599",
    checkpoint: "26841750",
  };
};

const givenANonSourceTransaction = () => {
  sourceTx = {
    digest: "CgSJ3izrN278gW9zEeHon8nTkNUrpTiPR3UW5KCLZFAS",
    transaction: {
      data: {
        gasData: {
          payment: [
            {
              objectId: "0x1ad2872de3d42a31d9e00a0d905011d7c4430ed12c783465321e5a05ae2b9e33",
              version: "64933577",
              digest: "A294rJ6JVi5NWaDezZn7jL1K3syVDvaNPQrkTsyYFwbx",
            },
          ],
          owner: "0x8f96433805302fd06592a2a85fd01ee92861dd2f203b2aadfad9d6adcc9e024e",
          price: "749",
          budget: "2605244",
        },
        messageVersion: "v1",
        transaction: {
          kind: "ProgrammableTransaction",
          inputs: [
            {
              type: "object",
              objectType: "immOrOwnedObject",
              objectId: "0x526320c9422b39efa1f5e867757d92b6130d783345ef6a9dc91d630a429f1954",
              version: "70164784",
              digest: "61sZzAWDqkeAc6MTkKPGzHPeRSUe4iVkv9GZSvn4d8Fc",
            },
            {
              type: "object",
              objectType: "immOrOwnedObject",
              objectId: "0xfc1bb5aa69ddebe2bdb538f11a9c13497ec41009cbfce82ad50b0c98cc15a184",
              version: "70164784",
              digest: "9dugruLRgeasK7sSTzj1mdri3cLXS4Rz2AxbvFUWUUyd",
            },
            {
              type: "object",
              objectType: "immOrOwnedObject",
              objectId: "0xfd0d3ab4f0a61b397daf56b6fc06cb1cf13ef9434bb7a93e125a7ccd2dc45213",
              version: "70161224",
              digest: "3983cYhVcDSakbGFYC1TLx6aJoQHjjmNAK2D9q62mDAL",
            },
            { type: "pure", valueType: "u64", value: "5000000" },
            {
              type: "pure",
              valueType: "address",
              value: "0x39340b05c03ffdbdf372d77e188dd68e8f08e1ab4ed222eab5d2fcc020878e13",
            },
            {
              type: "pure",
              valueType: "address",
              value: "0x39340b05c03ffdbdf372d77e188dd68e8f08e1ab4ed222eab5d2fcc020878e13",
            },
            { type: "pure", valueType: "u64", value: "5000000" },
            { type: "pure", valueType: "u64", value: "0" },
            { type: "pure", valueType: "u64", value: "0" },
            {
              type: "object",
              objectType: "sharedObject",
              objectId: "0xc57508ee0d4595e5a8728974a4a93a787d38f339757230d441e895422c07aba9",
              initialSharedVersion: "66",
              mutable: true,
            },
            { type: "pure", valueType: "u16", value: 4 },
            {
              type: "pure",
              valueType: "vector<u8>",
              value: [
                0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 87, 225, 115, 243, 190, 2, 245, 212, 54, 188,
                192, 126, 250, 198, 58, 88, 77, 1, 224, 163,
              ],
            },
            { type: "pure", valueType: "u64", value: "0" },
            { type: "pure", valueType: "u32", value: 63790 },
            {
              type: "object",
              objectType: "sharedObject",
              objectId: "0xaeab97f96cf9877fee2883315d459552b2b921edc16d7ceac6eab944dd88919c",
              initialSharedVersion: "64",
              mutable: true,
            },
            {
              type: "object",
              objectType: "sharedObject",
              objectId: "0x0000000000000000000000000000000000000000000000000000000000000006",
              initialSharedVersion: "1",
              mutable: false,
            },
          ],
          transactions: [
            { MakeMoveVec: [null, [{ Input: 0 }, { Input: 1 }, { Input: 2 }]] },
            {
              MoveCall: {
                package: "0xf331b35dfa0a560eadbe14280054c70908a487755bab106b5d528562f67ba8d1",
                module: "utils",
                function: "split_coins_and_transfer_rest",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ Result: 0 }, { Input: 3 }, { Input: 4 }],
              },
            },
            { MakeMoveVec: [null, [{ Result: 1 }]] },
            {
              MoveCall: {
                package: "0xf331b35dfa0a560eadbe14280054c70908a487755bab106b5d528562f67ba8d1",
                module: "swap_aggregator",
                function: "initiate_swap",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ Result: 2 }, { Input: 6 }, { Input: 7 }, { Input: 5 }],
              },
            },
            { SplitCoins: ["GasCoin", [{ Input: 8 }]] },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "state",
                function: "verified_asset",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ Input: 9 }],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "transfer_tokens",
                function: "prepare_transfer",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [
                  { NestedResult: [5, 0] },
                  { Result: 3 },
                  { Input: 10 },
                  { Input: 11 },
                  { Input: 12 },
                  { Input: 13 },
                ],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "coin_utils",
                function: "return_nonzero",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ NestedResult: [6, 1] }],
              },
            },
            {
              MoveCall: {
                package: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
                module: "transfer_tokens",
                function: "transfer_tokens",
                type_arguments: [
                  "0x5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
                ],
                arguments: [{ Input: 9 }, { NestedResult: [6, 0] }],
              },
            },
            {
              MoveCall: {
                package: "0x5306f64e312b581766351c07af79c72fcb1cd25147157fdc2f8ad76de9a3fb6a",
                module: "publish_message",
                function: "publish_message",
                arguments: [
                  { Input: 14 },
                  { NestedResult: [4, 0] },
                  { NestedResult: [8, 0] },
                  { Input: 15 },
                ],
              },
            },
          ],
        },
        sender: "0x39340b05c03ffdbdf372d77e188dd68e8f08e1ab4ed222eab5d2fcc020878e13",
      },
      txSignatures: [
        "AC0E+rayE0fL8LbH644pU8br2Lfay8EsOfkuq2x/oXmIYI1gK1zjpRH58fIHpLUPylbLfzCGyVITLdCAHhGQtwB+OJOHLGpRJukyYnp+TUImmLMhwM8Cff+IQHrLjM7+hQ==",
      ],
    },
    effects: {
      status: { status: "success" },
    },
    events: [
      {
        id: { txDigest: "CgSJ3izrN278gW9zEeHon8nTkNUrpTiPR3UW5KCLZFAS", eventSeq: "0" },
        packageId: "0xf331b35dfa0a560eadbe14280054c70908a487755bab106b5d528562f67ba8d1",
        transactionModule: "swap_aggregator",
        sender: "0x39340b05c03ffdbdf372d77e188dd68e8f08e1ab4ed222eab5d2fcc020878e13",
        type: "0xf331b35dfa0a560eadbe14280054c70908a487755bab106b5d528562f67ba8d1::swap_aggregator::InitiateSwapEvent",
        parsedJson: {
          amount: "5000000",
          coin_type: {
            name: "5d4b302506645c37ff133b98c4b50a5ae14841659738d6d733d59d0d217a93bf::coin::COIN",
          },
          steps: "0",
        },
        bcs: "Ji2VvXjm7tZ2a7qdSke7xSq9jfpdLYQNcKRDxpgGjECjyLkoSHbCoK2zhDnRTSiwR2K3PoPbCjWKdHw1YVY9PgYQiRvUmrwDGay6yctQidjxNTuEdConUwdaHZF8ZY3",
      },
    ],
    timestampMs: "1708526724599",
    checkpoint: "26841750",
  };
};
