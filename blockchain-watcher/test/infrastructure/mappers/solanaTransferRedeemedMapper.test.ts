import { expect, describe, it, jest } from "@jest/globals";
import { solana } from "../../../src/domain/entities";
import { solanaTransferRedeemedMapper } from "../../../src/infrastructure/mappers";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";

jest.mock("@certusone/wormhole-sdk/lib/cjs/solana/wormhole");

describe("solanaTransferRedeemedMapper", () => {
  it("should map a solana transaction to a log-message-published event", async () => {
    const mockGetPostedMessage = getPostedMessage as jest.MockedFunction<typeof getPostedMessage>;
    mockGetPostedMessage.mockResolvedValueOnce({
      message: {
        emitterChain: 2,
        sequence: 1500n,
        emitterAddress: Buffer.from(
          "0000000000000000000000003ee18b2214aff97000d974cf647e7c347e8fa585",
          "hex"
        ),
        submissionTime: 1700571923,
        nonce: 0,
        consistencyLevel: 1,
        payload: Buffer.from("41QVZTrdrRxb", "base64"),
      } as any,
    });

    const programId = "wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb";
    const tx = {
      blockTime: 1701724272,
      meta: {
        computeUnitsConsumed: 80797,
        err: null,
        fee: 5000,
        innerInstructions: [
          {
            index: 0,
            instructions: [
              {
                accounts: [0, 3],
                data: "3Bxs49175da2o1zw",
                programIdIndex: 4,
                stackHeight: 2,
              },
              {
                accounts: [3],
                data: "9krTCzbLfv4BRBcj",
                programIdIndex: 4,
                stackHeight: 2,
              },
              {
                accounts: [3],
                data: "SYXsBvR59hMYH7jGFg8pjr13roqCKDy5t1HFBVKFNWZ1FPp7",
                programIdIndex: 4,
                stackHeight: 2,
              },
              {
                accounts: [2, 1, 6],
                data: "6jFrQ56LiKZ1",
                programIdIndex: 11,
                stackHeight: 2,
              },
              {
                accounts: [2, 1, 6],
                data: "6AjePwYNteRu",
                programIdIndex: 11,
                stackHeight: 2,
              },
            ],
          },
        ],
        loadedAddresses: {
          readonly: [],
          writable: [],
        },
        logMessages: [
          "Program wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb invoke [1]",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: MintTo",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4492 of 136305 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: MintTo",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4589 of 125187 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb consumed 80797 of 200000 compute units",
          "Program wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb success",
        ],
        postBalances: [
          48364014319, 2039280, 7147804440, 897840, 1, 1134480, 153131100, 2477760, 1113600,
          1127520, 1009200, 934087680, 1141440, 1141440,
        ],
        postTokenBalances: [
          {
            accountIndex: 1,
            mint: "7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs",
            owner: "5yZiE74sGLCT4uRoyeqz4iTYiUwX5uykiPRggCVih9PN",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "98753117",
              decimals: 8,
              uiAmount: 0.98753117,
              uiAmountString: "0.98753117",
            },
          },
        ],
        preBalances: [
          48364917159, 2039280, 7147804440, 0, 1, 1134480, 153131100, 2477760, 1113600, 1127520,
          1009200, 934087680, 1141440, 1141440,
        ],
        preTokenBalances: [
          {
            accountIndex: 1,
            mint: "7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs",
            owner: "5yZiE74sGLCT4uRoyeqz4iTYiUwX5uykiPRggCVih9PN",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "93339",
              decimals: 8,
              uiAmount: 0.00093339,
              uiAmountString: "0.00093339",
            },
          },
        ],
        rewards: [],
        status: {
          Ok: null,
        },
      },
      slot: 234015120,
      transaction: {
        message: {
          header: {
            numReadonlySignedAccounts: 0,
            numReadonlyUnsignedAccounts: 10,
            numRequiredSignatures: 1,
          },
          accountKeys: [
            "7dm9am6Qx7cH64RB99Mzf7ZsLbEfmXM7ihXXCvMiT2X1",
            "4RrFMkY3A5zWdizT61Px222qmSTJqnVszDeBZZNSoAH6",
            "7vfCXTUXx5WJV5JADk17DUJ4ksgau7utNKj4b963voxs",
            "HkpTbh5td45g3SfFsKmjukX2YZUKfEHZG5HfRrw6Tkyi",
            "11111111111111111111111111111111",
            "2gQuwC9GMUCVcXw9VffeCswhKbPeyzHH9ZPEnRBw4K9w",
            "BCD75RNBHrJJpW4dXVagL5mPjzRLnVZq4YirJdjEYMV7",
            "CvYA8s1SnSzQzCv71rjt7Sc9iEVNjz2exRpoucyH2RCE",
            "DapiQYH3BGonhN8cngWcXQ6SrqSm3cwysoznoHr6Sbsx",
            "DujfLgMKW71CT2W8pxknf42FT86VbcK5PjQ6LsutjWKC",
            "SysvarRent111111111111111111111111111111111",
            "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
            "wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb",
          ],
          recentBlockhash: "H9bDwcfav3nq9wvzJi9sEPCm4oHxqS59nZU81mncD6AU",
          instructions: [
            {
              accounts: [0, 8, 7, 3, 9, 1, 1, 2, 5, 6, 10, 4, 11, 12],
              data: "4",
              programIdIndex: 13,
              stackHeight: null,
            },
          ],
          indexToProgramIds: {},
          compiledInstructions: [
            {
              programIdIndex: 13,
              accountKeyIndexes: [0, 8, 7, 3, 9, 1, 1, 2, 5, 6, 10, 4, 11, 12],
              data: new Uint8Array([3]),
            },
          ],
        },
        signatures: [
          "3FySmshUgVCM2N158oNYbeTfZt2typEU32c9ZxdAXiXURFHuTmeJHhc7cSUtqHdwAsbVWWvEsEddWNAKzkjVPSg2",
        ],
      },
      version: "legacy",
    } as any as solana.Transaction;

    const events = await solanaTransferRedeemedMapper(tx, { programId });

    expect(events).toHaveLength(1);
    expect(events[0].name).toBe("transfer-redeemed");
    expect(events[0].address).toBe(programId);
    expect(events[0].chainId).toBe(1);
    expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
    expect(events[0].blockHeight).toBe(BigInt(tx.slot));
    expect(events[0].blockTime).toBe(tx.blockTime);
  });
});
