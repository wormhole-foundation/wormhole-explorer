import { beforeAll, describe, expect, it } from "@jest/globals";
import { HandleSuiTransactions } from "../../../../src/domain/actions/sui/HandleSuiTransactions";
import { TransactionFoundEvent, TransferRedeemed } from "../../../../src/domain/entities";
import { SuiTransactionBlockReceipt } from "../../../../src/domain/entities/sui";

let txs: SuiTransactionBlockReceipt[] = [];
let vaaInfos: Record<string, TransferRedeemed>;

const REDEEM_EVENT =
  "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d::complete_transfer::TransferRedeemed";

const cfg = { id: "handle-sui-transactions-test" };

const statsRepo = {
  count: () => {},
  measure: () => {},
  report: () => Promise.resolve(""),
};

const mapper = (tx: SuiTransactionBlockReceipt): TransactionFoundEvent => {
  return {
    name: "send-event",
    address: "0x12345",
    chainId: 1,
    txHash: tx.digest,
    blockHeight: 0n,
    blockTime: 0,
    attributes: {
      from: "",
      to: "",
      emitterAddress: vaaInfos[tx.digest].emitterAddress,
      emitterChain: vaaInfos[tx.digest].emitterChainId,
      sequence: vaaInfos[tx.digest].sequence,
      status: "ok",
      protocol: "Token Bridge Manual",
    },
  };
};

describe.only("HandleSuiTransactions", () => {
  beforeAll(() => {
    givenTransactions();
  });

  it("returns no transactions when no events configured", async () => {
    const handler = new HandleSuiTransactions(cfg, mapper, () => Promise.resolve(), statsRepo);

    const result = await handler.handle(txs);

    expect(result).toHaveLength(0);
  });

  it("returns mapped transactions filtered by the configured event", async () => {
    const handler = new HandleSuiTransactions(
      {
        ...cfg,
        eventTypes: [REDEEM_EVENT],
      },
      mapper,
      () => Promise.resolve(),
      statsRepo
    );

    const result = await handler.handle(txs);

    expect(result).toHaveLength(2);

    expect(result[0].txHash).toEqual("8Mhx2j5XBhpTa18Hr6WvaVq5nJbaDN3tuf72S491GmNm");
    expect(result[0].attributes.emitterAddress).toEqual(
      "000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7"
    );
    expect(result[0].attributes.emitterChain).toEqual(4);
    expect(result[0].attributes.sequence).toEqual(394747);

    expect(result[1].txHash).toEqual("AGftEd2E2EwF2Sxavjmgx2sgG6b4KNx44tJYZD9U3W7d");
    expect(result[1].attributes.emitterAddress).toEqual(
      "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5"
    );
    expect(result[1].attributes.emitterChain).toEqual(1);
    expect(result[1].attributes.sequence).toEqual(610675);
  });
});

function givenTransactions() {
  txs = [
    {
      digest: "8Mhx2j5XBhpTa18Hr6WvaVq5nJbaDN3tuf72S491GmNm",
      transaction: {} as any,
      events: [
        {
          id: { txDigest: "8Mhx2j5XBhpTa18Hr6WvaVq5nJbaDN3tuf72S491GmNm", eventSeq: "0" },
          packageId: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
          transactionModule: "complete_transfer",
          sender: "0xebcfeff722c1c5a25cb5654d95f8da059702180207990a6c286b952dd430c26a",
          type: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d::complete_transfer::TransferRedeemed",
          parsedJson: [Object],
          bcs: "J6LtZQPC5w4FPPJDej6ur8QJfjbwfW3TvmPFg17nFaZLziix6wmyfNv9xK",
        },
      ],
      timestampMs: "1706107632222",
      checkpoint: "24408491",
    },
    {
      digest: "JDkjZeDooHNW4hN9vD7rCedJdoRdBd138AhDUmgxxeYR",
      transaction: {} as any,
      events: [
        {
          id: { txDigest: "JDkjZeDooHNW4hN9vD7rCedJdoRdBd138AhDUmgxxeYR", eventSeq: "0" },
          packageId: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e",
          transactionModule: "pyth",
          sender: "0x02a212de6a9dfa3a69e22387acfbafbb1a9e591bd9d636e7895dcfc8de05f331",
          type: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e::event::PriceFeedUpdateEvent",
          parsedJson: {} as any,
          bcs: "55FXDcsje9zDJJ9agwYbQs8U6VcJF9pb1zsgXrZwywJZJSV4JGqXSQPYReiWiS3pcCSqd8YXfH3K94P7JmCgaU9raigpSFqK8Xcams4aUK24ov5b7rz2ybjBugwtKSQK8DabkPbzcm3tucWy6VovP",
        },
        {
          id: { txDigest: "JDkjZeDooHNW4hN9vD7rCedJdoRdBd138AhDUmgxxeYR", eventSeq: "1" },
          packageId: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e",
          transactionModule: "pyth",
          sender: "0x02a212de6a9dfa3a69e22387acfbafbb1a9e591bd9d636e7895dcfc8de05f331",
          type: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e::event::PriceFeedUpdateEvent",
          parsedJson: {} as any,
          bcs: "55sXKZQLRJEYi7a5tKbMrpeptEnnQiZVy31xsoV8VrmoXdMAwG11EbJBB7BuB2de6r8Nz13bbfKNEteDK65Fb1NHPxYkCwEan3zkX6Sb32KoQk5NLacqsTM3zKaVsvMuAGMJc5qiEd2E5RGmAC73m",
        },
        {
          id: { txDigest: "JDkjZeDooHNW4hN9vD7rCedJdoRdBd138AhDUmgxxeYR", eventSeq: "2" },
          packageId: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e",
          transactionModule: "pyth",
          sender: "0x02a212de6a9dfa3a69e22387acfbafbb1a9e591bd9d636e7895dcfc8de05f331",
          type: "0x8d97f1cd6ac663735be08d1d2b6d02a159e711586461306ce60a2b7a6a565a9e::event::PriceFeedUpdateEvent",
          parsedJson: {} as any,
          bcs: "58HfNyLepUTv5di5XsbeqQL8zXttiTW5qDL8fbC3yYqqfEHECSk3yg8ocqzM2jxeRUWrVSddRaKafMLYT5GWaF72c8U32A9P7AGp1K24vz9rpwKn7cR7Zrr3kX4uo2dy36Ly1sWoEQe1vTUsGwcb1",
        },
      ],
      timestampMs: "1706107711822",
      checkpoint: "24408569",
    },
    {
      digest: "AGftEd2E2EwF2Sxavjmgx2sgG6b4KNx44tJYZD9U3W7d",
      transaction: {} as any,
      events: [
        {
          id: { txDigest: "AGftEd2E2EwF2Sxavjmgx2sgG6b4KNx44tJYZD9U3W7d", eventSeq: "0" },
          packageId: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d",
          transactionModule: "complete_transfer",
          sender: "0xc748251620ff8bcdc7de89f1b8f0d4b1023790525265534bce8545a710f81662",
          type: "0x26efee2b51c911237888e5dc6702868abca3c7ac12c53f76ef8eba0697695e3d::complete_transfer::TransferRedeemed",
          parsedJson: {} as any,
          bcs: "5GvwRFsvFeuD7MNApPR4TV4hcYNZHwXoKuAet47kpHZugnKNCbk3aH4Sns",
        },
      ],
      timestampMs: "1706107714828",
      checkpoint: "24408572",
    },
  ];

  vaaInfos = {
    "8Mhx2j5XBhpTa18Hr6WvaVq5nJbaDN3tuf72S491GmNm": {
      emitterAddress: "000000000000000000000000b6f6d86a8f9879a9c87f643768d9efc38c1da6e7",
      emitterChainId: 4,
      sequence: 394747,
    },
    AGftEd2E2EwF2Sxavjmgx2sgG6b4KNx44tJYZD9U3W7d: {
      emitterAddress: "ec7372995d5cc8732397fb0ad35c0121e0eaa90d26f828a534cab54391b3a4f5",
      emitterChainId: 1,
      sequence: 610675,
    },
  };
}
