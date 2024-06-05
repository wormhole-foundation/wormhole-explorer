import { expect, describe, it, jest } from "@jest/globals";
import { solana } from "../../../../src/domain/entities";
import { solanaTransferRedeemedMapper } from "../../../../src/infrastructure/mappers";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";

jest.mock("@certusone/wormhole-sdk/lib/cjs/solana/wormhole");

describe("solanaTransferRedeemedMapper", () => {
  it("should map a token bridge tx to a transfer-redeemed event", async () => {
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

    const programs = {
      DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe: {
        instructions: ["02", "03", "09", "0a"],
        vaaAccountIndex: 2,
      },
    };
    const tx = {
      blockTime: 1701724272,
      meta: {
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
        logMessages: [
          "Program DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe invoke [1]",
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
          "Program DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe consumed 80797 of 200000 compute units",
          "Program DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe success",
        ],
        status: {
          Ok: null,
        },
      },
      slot: 234015120,
      chainId: 1,
      chain: "solana",
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
            "DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe",
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

    const events = await solanaTransferRedeemedMapper(tx, { programs });

    if (events) {
      expect(events).toHaveLength(1);
      expect(events[0].name).toBe("transfer-redeemed");
      expect(events[0].address).toBe("DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe");
      expect(events[0].chainId).toBe(1);
      expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
      expect(events[0].blockHeight).toBe(BigInt(tx.slot));
      expect(events[0].blockTime).toBe(tx.blockTime);
      expect(events[0].attributes.methodsByAddress).toBe("completeWrappedInstruction");
      expect(events[0].attributes.status).toBe("completed");
    }
  });

  it("should map a tx involving token bridge relayer (aka connect) to a transfer-redeemed event", async () => {
    const mockGetPostedMessage = getPostedMessage as jest.MockedFunction<typeof getPostedMessage>;
    mockGetPostedMessage.mockResolvedValueOnce({
      message: {
        emitterChain: 4,
        sequence: 5185,
        emitterAddress: Buffer.from(
          "0000000000000000000000009dcf9d205c9de35334d646bee44b2d2859712a09",
          "hex"
        ),
        submissionTime: 1700571923,
        nonce: 0,
        consistencyLevel: 1,
        payload: Buffer.from("41QVZTrdrRxb", "base64"),
      } as any,
    });

    const programs = {
      DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe: {
        instructions: ["02", "03", "09", "0a"],
        vaaAccountIndex: 2,
      },
    };
    const tx = {
      blockTime: 1701701948,
      chainId: 1,
      chain: "solana",
      meta: {
        innerInstructions: [
          {
            index: 0,
            instructions: [
              {
                accounts: [0, 1],
                data: "11119os1e9qSs2u7TsThXqkBSRVFxhmYaFKFZ1waB2X7armDmvK3p5GmLdUxYdg3h7QSrL",
                programIdIndex: 7,
                stackHeight: 2,
              },
              {
                accounts: [1, 4],
                data: "6NejZzEkDLeuHiYpvQR3Ck46Sw6FQeXFmX5TGWpBSLgJ1",
                programIdIndex: 22,
                stackHeight: 2,
              },
              {
                accounts: [0, 15, 8, 5, 16, 1, 9, 1, 4, 12, 20, 21, 7, 11, 22],
                data: "B",
                programIdIndex: 18,
                stackHeight: 2,
              },
              {
                accounts: [0, 5],
                data: "11112ncWAFpbecrgZiGiLpaHnEYkj7ECUfBBRHr4H5tFCq9bHFXWRyWUjj586frtFc19oa",
                programIdIndex: 7,
                stackHeight: 3,
              },
              {
                accounts: [4, 1, 20],
                data: "6j1A9VR8zuFm",
                programIdIndex: 22,
                stackHeight: 3,
              },
              {
                accounts: [1, 2, 9],
                data: "3tMLEJ9BQpG7",
                programIdIndex: 22,
                stackHeight: 2,
              },
              {
                accounts: [1, 6, 9],
                data: "3qeniiQUmAqm",
                programIdIndex: 22,
                stackHeight: 2,
              },
              {
                accounts: [1, 0, 9],
                data: "A",
                programIdIndex: 22,
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
          "Program 3bPRWXqtSfUaCw3S4wdgvypQtsSzcmvDeaqSqPDkncrg invoke [1]",
          "Program log: Instruction: CompleteWrappedTransferWithRelay",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: InitializeAccount3",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4214 of 223786 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe invoke [2]",
          "Program log: Instruction: LegacyCompleteTransferWithPayloadWrapped",
          "Program 11111111111111111111111111111111 invoke [3]",
          "Program 11111111111111111111111111111111 success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [3]",
          "Program log: Instruction: MintTo",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4492 of 128286 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe consumed 50570 of 173121 compute units",
          "Program DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: Transfer",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 119156 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: Transfer",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 111686 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: CloseAccount",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3015 of 104235 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program 3bPRWXqtSfUaCw3S4wdgvypQtsSzcmvDeaqSqPDkncrg consumed 150503 of 250000 compute units",
          "Program 3bPRWXqtSfUaCw3S4wdgvypQtsSzcmvDeaqSqPDkncrg success",
          "Program ComputeBudget111111111111111111111111111111 invoke [1]",
          "Program ComputeBudget111111111111111111111111111111 success",
        ],
        status: {
          Ok: null,
        },
      },
      slot: 262968784,
      transaction: {
        message: {
          header: {
            numReadonlySignedAccounts: 0,
            numReadonlyUnsignedAccounts: 16,
            numRequiredSignatures: 1,
          },
          accountKeys: [
            "hiUN9rS9VTPVGYc71Vf2d6iyFLvsQaSsqWhxydqdaZf",
            "14UpGeFGK9iEhVTgMbdd7RHZmBKb8BYxBpAYGtBWigeT",
            "3pjtJPtu7Z9NzQinCaaRyUsZPwSrNjGxYi7RMwkigk47",
            "8DW7zrpEe9EVxbD8PBfmEKNkCwXPFTVVukHXirrcE9iV",
            "BaGfF51MQ3a61papTRDYaNefBgTQ9ywnVne5fCff4bxT",
            "bMDMKEYXfWM2h5AyJL4kzBXLr8Wms29NSWeGQGLgEad",
            "EGdE1V4GLFyZH5FFDtsxZaRkw84WhwPWJkCV4yT2L6F4",
            "11111111111111111111111111111111",
            "2EqgRpRxi1MR8QLycFDcMKws1Kv56dQcwSmXKkFJZgnW",
            "2X2u43DR3odTT4jQKFqsG5f4SCfbEmz4pAHR2EVd3Xs3",
            "3bPRWXqtSfUaCw3S4wdgvypQtsSzcmvDeaqSqPDkncrg",
            "3u8hJUVTA4jH1wYAyUur7FFZVQ8H635K3tSHHF4ssjQ5",
            "5rmBUDWruRcWU4S4JyjPLZhcYJARYw4FeU9brEo4nUzo",
            "7pFgBNscwYBfKs1Bsi7wgyibmtNCmK4442pgL4xiJcTr",
            "8huuoQHYxWGs3oYoKXJBsBgPousVa9XfkPyHWrpPH1B8",
            "8PFZNjn19BBYVHNp4H31bEW7eAmu78Yf2RKV8EeA461K",
            "9A7Z7kJw7hPsBJQDSeFU63DsgZGhSTjXmGLYS81yCHBN",
            "ComputeBudget111111111111111111111111111111",
            "DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe",
            "HeMY4WFgEA8zkbp7HMCma5mswLArPzPe53EkYcXDJTUV",
            "rRsXLHe7sBHdyKU3KY3wbcgWvoT1Ntqudf6e9PKusgb",
            "SysvarRent111111111111111111111111111111111",
            "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
          ],
          recentBlockhash: "KVvqcsFkfd4mgWjqjsEATKvciAwCwUgdeVHeRh4LqyT",
          instructions: [
            {
              accounts: [0, 9, 2, 19, 4, 6, 3, 13, 14, 1, 12, 15, 8, 5, 16, 20, 11, 18, 7, 22, 21],
              data: "9ewcWKkpwjKSXZao88ZRqjejhWC1Bf6cFzdxHbDN4taHHCL3zM9P3xi",
              programIdIndex: 10,
              stackHeight: null,
            },
            {
              accounts: [],
              data: "HnkkG7",
              programIdIndex: 17,
              stackHeight: null,
            },
          ],
          indexToProgramIds: {},
          compiledInstructions: [
            {
              programIdIndex: 10,
              accountKeyIndexes: [
                0, 9, 2, 19, 4, 6, 3, 13, 14, 1, 12, 15, 8, 5, 16, 20, 11, 18, 7, 22, 21,
              ],
              data: new Uint8Array([
                174, 44, 4, 91, 81, 201, 235, 255, 59, 128, 71, 194, 194, 46, 49, 88, 200, 5, 254,
                175, 217, 196, 30, 63, 1, 233, 245, 96, 162, 12, 73, 62, 205, 171, 142, 159, 18, 6,
                57, 151,
              ]),
            },
            {
              programIdIndex: 17,
              accountKeyIndexes: [],
              data: new Uint8Array([2, 144, 208, 3, 0]),
            },
          ],
        },
        signatures: [
          "5Cu3tD15AtcQ5NGK6PFT9UMVmqh94ARz8FXpFts5G1nNzQnXXaugP3ELa79P9xCwESC5Kw7FtGHUgh7vz8DuP8tM",
        ],
      },

      version: "legacy",
    } as any as solana.Transaction;

    const events = await solanaTransferRedeemedMapper(tx, { programs });

    if (events) {
      expect(events).toHaveLength(1);
      expect(events[0].name).toBe("transfer-redeemed");
      expect(events[0].address).toBe("DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe");
      expect(events[0].chainId).toBe(1);
      expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
      expect(events[0].blockHeight).toBe(BigInt(tx.slot));
      expect(events[0].blockTime).toBe(tx.blockTime);
      expect(events[0].attributes.methodsByAddress).toBe("CompleteWrappedWithPayloadInstruction");
      expect(events[0].attributes.status).toBe("completed");
    }
  });
});
