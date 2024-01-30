import { describe, expect, it, jest } from "@jest/globals";
import { Checkpoint } from "@mysten/sui.js/client";

import { GetSuiTransactions } from "../../../../src/domain/actions/sui/GetSuiTransactions";
import { Range } from "../../../../src/domain/entities";
import { SuiRepository } from "../../../../src/domain/repositories";
import { SuiTransactionBlockReceipt } from "../../../../src/domain/entities/sui";

let getCheckpointsSpy: jest.SpiedFunction<SuiRepository["getCheckpoints"]>;
let getTransactionBlockReceiptsSpy: jest.SpiedFunction<
  SuiRepository["getTransactionBlockReceipts"]
>;

let suiRepo: SuiRepository;
let getSuiTransactions: GetSuiTransactions;

describe("GetSuiTransactions", () => {
  it("should return an empty array when the range is invalid", async () => {
    const range: Range = {
      from: 2n,
      to: 1n,
    };

    givenGetSuiTransactions();

    const result = await getSuiTransactions.execute(range);

    expect(result).toEqual([]);
  });

  it("should return an array of transaction block receipts given a valid range", async () => {
    const range: Range = { from: 30n, to: 38n };

    givenSuiRepo(range);
    givenGetSuiTransactions();

    const result = await getSuiTransactions.execute(range);

    expect(result[0].digest).toEqual("Hzp86Ua9dAA3ML4Ch6ygNFFVWCv35B5Hda7QmzpbXeLz");
    expect(result[1].digest).toEqual("LSLrvEfSRBoA4QbknzAxBcvFZPPYApeHMLp8rExEx8R");
    expect(result[2].digest).toEqual("9vFFvE4MNXYRWczMEf6mmKbbuscjipobYTWZB1Rex6Uc");
    expect(getCheckpointsSpy).toHaveBeenCalledTimes(1);
    expect(getTransactionBlockReceiptsSpy).toHaveBeenCalledTimes(1);
  });
});

function givenSuiRepo(range: Range) {
  const checkpoints: Checkpoint[] = [
    {
      epoch: "285",
      sequenceNumber: "24246904",
      digest: "FYoPczasCFhfeicq3XahtCicyy8uC4wRsNFEiGm2nNW",
      networkTotalTransactions: "1063849022",
      previousDigest: "HanjkjkfkpMW6pgNzW4dpWAmK2HeGvNsY9XFNBjaXKP1",
      epochRollingGasCostSummary: {
        computationCost: "169602918500",
        storageCost: "1234564033600",
        storageRebate: "1198647002916",
        nonRefundableStorageFee: "12107545484",
      },
      timestampMs: "1705946928610",
      transactions: [
        "Hzp86Ua9dAA3ML4Ch6ygNFFVWCv35B5Hda7QmzpbXeLz",
        "LSLrvEfSRBoA4QbknzAxBcvFZPPYApeHMLp8rExEx8R",
      ],
      checkpointCommitments: [],
      validatorSignature: "i9kUhaIs2SOZFMqYgAOf9rjZJs8lVWGUlxJwsiGcwW8Eg7u4EfWGHjJulg6ZwMWb",
    },
    {
      epoch: "285",
      sequenceNumber: "24246905",
      digest: "A2CFNCm4z3wsMdVXpB2EH6esjxncQrBk6mgadJEJdkay",
      networkTotalTransactions: "1063849029",
      previousDigest: "FYoPczasCFhfeicq3XahtCicyy8uC4wRsNFEiGm2nNW",
      epochRollingGasCostSummary: {
        computationCost: "169631418500",
        storageCost: "1234652687600",
        storageRebate: "1198722664260",
        nonRefundableStorageFee: "12108309740",
      },
      timestampMs: "1705946929608",
      transactions: ["9vFFvE4MNXYRWczMEf6mmKbbuscjipobYTWZB1Rex6Uc"],
      checkpointCommitments: [],
      validatorSignature: "ue+oUykmKpVUg77xMqowOFb16Ouw35KLL4AIfROouT1LAPsJ6nQirAT8vikHuEa0",
    },
  ];

  const txs: SuiTransactionBlockReceipt[] = [
    {
      digest: "Hzp86Ua9dAA3ML4Ch6ygNFFVWCv35B5Hda7QmzpbXeLz",
      transaction: {} as any,
      events: [],
      timestampMs: "1705946928610",
      checkpoint: "24246904",
    },
    {
      digest: "LSLrvEfSRBoA4QbknzAxBcvFZPPYApeHMLp8rExEx8R",
      transaction: {} as any,
      events: [
        {
          id: { txDigest: "LSLrvEfSRBoA4QbknzAxBcvFZPPYApeHMLp8rExEx8R", eventSeq: "0" },
          packageId: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e",
          transactionModule: "pyth",
          sender: "0x02a212de6a9dfa3a69e22387acfbafbb1a9e591bd9d636e7895dcfc8de05f331",
          type: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e::event::PriceFeedUpdateEvent",
          parsedJson: {},
          bcs: "5ApxYAdcmDtesn1Pif8BoF1xGqCwLvVuJ8DXQEryBXBeShUk72AJR4ENZ7nBVjr4PwtZnUXGzV3M2T3nBkHxGZFHeKeN6cz1hHUYLisnPuQhv7pKEPR7FSHzcE4B1F4AtKiYGWrHG5wrA4wWfyBLB",
        },
        {
          id: { txDigest: "LSLrvEfSRBoA4QbknzAxBcvFZPPYApeHMLp8rExEx8R", eventSeq: "1" },
          packageId: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e",
          transactionModule: "pyth",
          sender: "0x02a212de6a9dfa3a69e22387acfbafbb1a9e591bd9d636e7895dcfc8de05f331",
          type: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e::event::PriceFeedUpdateEvent",
          parsedJson: {},
          bcs: "55FXDcsje9zDJJ9agwYbQs8U6VcJF9pb1zsgXrZwywJZJSVnhTcrTQbB26iKkLZnBJcTFFxiGu37eoGqpsH99Eqp4Tz6YJ6P2BmyGsvEr3HWqGEYMFNemhGjacYHMB6i9cTNRWnUavrtw7q2TzaYo",
        },
        {
          id: { txDigest: "LSLrvEfSRBoA4QbknzAxBcvFZPPYApeHMLp8rExEx8R", eventSeq: "2" },
          packageId: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e",
          transactionModule: "pyth",
          sender: "0x02a212de6a9dfa3a69e22387acfbafbb1a9e591bd9d636e7895dcfc8de05f331",
          type: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e::event::PriceFeedUpdateEvent",
          parsedJson: {},
          bcs: "597UCGRvQTDbbKmaK6NneKvAsmDrE5sb4hsdhrskwWiXrUwsL7xGNKGypJyWKRw6y3BcWiay8wRnJMmXKg2RudYDq6HvuFVuwv7S54tgBcTXYBHsyDkpT3xBBzemXNpa7LqGtxMDhM1Up4QACu8QX",
        },
      ],
      timestampMs: "1705946928610",
      checkpoint: "24246904",
    },
    {
      digest: "9vFFvE4MNXYRWczMEf6mmKbbuscjipobYTWZB1Rex6Uc",
      transaction: {
        data: {
          messageVersion: "v1",
          transaction: {} as any,
          sender: "0x0000000000000000000000000000000000000000000000000000000000000000",
          gasData: {} as any,
        },
        txSignatures: [
          "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
        ],
      },
      events: [],
      timestampMs: "1705946929608",
      checkpoint: "24246905",
    },
  ];

  suiRepo = {
    getCheckpoints: () => Promise.resolve(checkpoints),
    getTransactionBlockReceipts: () => Promise.resolve(txs),
    getLastCheckpointNumber: () => Promise.resolve(0n),
    getLastCheckpoint: () => Promise.resolve({} as any),
    getCheckpoint: (id: string | bigint | number) => Promise.resolve({} as any),
    queryTransactions: () => Promise.resolve([]),
    queryTransactionsByEvent: () => Promise.resolve([]),
  };

  getCheckpointsSpy = jest.spyOn(suiRepo, "getCheckpoints");
  getTransactionBlockReceiptsSpy = jest.spyOn(suiRepo, "getTransactionBlockReceipts");
}

function givenGetSuiTransactions() {
  getSuiTransactions = new GetSuiTransactions(suiRepo);
}
