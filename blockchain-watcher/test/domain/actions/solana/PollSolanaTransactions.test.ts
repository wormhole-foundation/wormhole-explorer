import { afterEach, describe, it, expect, jest } from "@jest/globals";
import {
  PollSolanaTransactions,
  PollSolanaTransactionsConfig,
  PollSolanaTransactionsMetadata,
} from "../../../../src/domain/actions";
import {
  MetadataRepository,
  SolanaSlotRepository,
  StatRepository,
} from "../../../../src/domain/repositories";
import { thenWaitForAssertion } from "../../../wait-assertion";
import { Fallible, solana } from "../../../../src/domain/entities";

let pollSolanaTransactions: PollSolanaTransactions;
let cfg: PollSolanaTransactionsConfig;

let metadataRepo: MetadataRepository<PollSolanaTransactionsMetadata>;
let metadataGetSpy: jest.SpiedFunction<MetadataRepository<any>["get"]>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<any>["save"]>;
let solanaSlotRepo: SolanaSlotRepository;
let getLatestSlotSpy: jest.SpiedFunction<SolanaSlotRepository["getLatestSlot"]>;
let getSigsSpy: jest.SpiedFunction<SolanaSlotRepository["getSignaturesForAddress"]>;
let getBlockSpy: jest.SpiedFunction<SolanaSlotRepository["getBlock"]>;
let handlers = {
  working: (logs: any[]) => Promise.resolve(),
};
let handlerSpy: jest.SpiedFunction<(transactions: any[]) => Promise<void>>;
let statsRepo: StatRepository;

describe("PollSolanaTransactions", () => {
  afterEach(async () => {
    await pollSolanaTransactions.stop();
  });

  it("should be able to read transactions from latest slot when no fromSlot is configured", async () => {
    const currentSlot = 10;
    const expectedSigs = givenSigs();
    const expectedTxs = givenTxs();

    givenCfg();
    givenStatsRepository();
    givenMetadataRepository();
    givenSolanaSlotRepository(currentSlot, givenBlock(1), expectedSigs, expectedTxs);
    givenPollSolanaTransactions();

    pollSolanaTransactions.run([handlers.working]);

    await thenWaitForAssertion(
      () => expect(metadataGetSpy).toHaveBeenCalledWith(cfg.id),
      () => expect(getLatestSlotSpy).toHaveBeenCalledWith(cfg.commitment),
      () => expect(getBlockSpy).toHaveBeenCalledWith(currentSlot),
      () => expect(handlerSpy).toHaveBeenCalledWith(expectedTxs),
      () =>
        expect(getSigsSpy).toHaveBeenCalledWith(
          cfg.programId,
          expectedSigs[0].signature,
          expectedSigs[expectedSigs.length - 1].signature,
          cfg.signaturesLimit
        ),
      () =>
        expect(metadataSaveSpy).toHaveBeenCalledWith(cfg.id, {
          lastSlot: currentSlot,
        })
    );
  });

  it("should be able to read transactions from last known slot", async () => {
    const latestSlot = 100;
    const lastSlot = 10;
    const expectedSigs = givenSigs();
    const expectedTxs = givenTxs();

    givenCfg();
    givenStatsRepository();
    givenMetadataRepository({ lastSlot });
    givenSolanaSlotRepository(latestSlot, givenBlock(1), expectedSigs, expectedTxs);
    givenPollSolanaTransactions();

    pollSolanaTransactions.run([handlers.working]);

    await thenWaitForAssertion(
      () => expect(getBlockSpy).toHaveBeenCalledWith(lastSlot + 1),
      () => expect(handlerSpy).toHaveBeenCalledWith(expectedTxs),
      () =>
        expect(getSigsSpy).toHaveBeenCalledWith(
          cfg.programId,
          expectedSigs[0].signature,
          expectedSigs[expectedSigs.length - 1].signature,
          cfg.signaturesLimit
        ),
      () =>
        expect(metadataSaveSpy).toHaveBeenCalledWith(cfg.id, {
          lastSlot: latestSlot,
        })
    );
  });
});

const givenCfg = () => {
  cfg = new PollSolanaTransactionsConfig("anId", "programID", "confirmed");
};

const givenMetadataRepository = (data?: PollSolanaTransactionsMetadata) => {
  metadataRepo = {
    get: () => Promise.resolve(data),
    save: () => Promise.resolve(),
  };
  metadataGetSpy = jest.spyOn(metadataRepo, "get");
  metadataSaveSpy = jest.spyOn(metadataRepo, "save");
};

const givenBlock = (blockTime: number) => {
  return {
    blockhash: "aBlockHash",
    transactions: [
      {
        transaction: {
          message: {
            accountKeys: [],
            compiledInstructions: [
              {
                programIdIndex: 0,
                accounts: [],
                data: "",
              },
            ],
            instructions: [
              {
                programIdIndex: 0,
                accounts: [],
                data: "",
              },
            ],
          },
          signatures: ["aSignature"],
        },
      },
    ],
    blockTime,
  } as any as solana.Block;
};

const givenSigs = () => {
  return [
    {
      signature: "aSignature",
      err: null,
      blockTime: 1,
    },
  ];
};

const givenTxs = () => {
  return [
    {
      signatures: ["aSignature"],
      message: {
        instructions: [
          {
            programIdIndex: 0,
            accounts: [],
            data: "",
          },
        ],
      },
    },
  ] as any as solana.Transaction[];
};

const givenSolanaSlotRepository = (
  currentSlot: number,
  block: solana.Block = {} as any as solana.Block,
  sigs: solana.ConfirmedSignatureInfo[] = [],
  txs: solana.Transaction[] = []
) => {
  solanaSlotRepo = {
    getLatestSlot: () => Promise.resolve(currentSlot),
    getBlock: () => Promise.resolve(Fallible.ok(block)),
    getSignaturesForAddress: () => Promise.resolve(sigs),
    getTransactions: () => Promise.resolve(txs),
  };
  getBlockSpy = jest.spyOn(solanaSlotRepo, "getBlock");
  getLatestSlotSpy = jest.spyOn(solanaSlotRepo, "getLatestSlot");
  getSigsSpy = jest.spyOn(solanaSlotRepo, "getSignaturesForAddress");
};

const givenPollSolanaTransactions = () => {
  handlerSpy = jest.spyOn(handlers, "working");
  pollSolanaTransactions = new PollSolanaTransactions(metadataRepo, solanaSlotRepo, statsRepo, cfg);
};

const givenStatsRepository = () => {
  statsRepo = {
    count: () => {},
    measure: () => {},
    report: () => Promise.resolve(""),
  };
};
