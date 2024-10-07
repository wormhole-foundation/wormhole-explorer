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
      DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe: [
        {
          instructions: ["02", "03", "09", "0a"],
          vaaAccountIndex: 2,
        },
      ],
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

    expect(events).toHaveLength(1);
    expect(events[0].name).toBe("transfer-redeemed");
    expect(events[0].address).toBe("DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe");
    expect(events[0].chainId).toBe(1);
    expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
    expect(events[0].blockHeight).toBe(BigInt(tx.slot));
    expect(events[0].blockTime).toBe(tx.blockTime);
    expect(events[0].attributes.methodsByAddress).toBe("completeWrappedInstruction");
    expect(events[0].attributes.status).toBe("completed");
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
      DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe: [
        {
          instructions: ["02", "03", "09", "0a"],
          vaaAccountIndex: 2,
        },
      ],
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

    expect(events).toHaveLength(1);
    expect(events[0].name).toBe("transfer-redeemed");
    expect(events[0].address).toBe("DZnkkTmCiFWfYTfT41X3Rd1kDgozqzxWaHqsw6W4x2oe");
    expect(events[0].chainId).toBe(1);
    expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
    expect(events[0].blockHeight).toBe(BigInt(tx.slot));
    expect(events[0].blockTime).toBe(tx.blockTime);
    expect(events[0].attributes.methodsByAddress).toBe("CompleteWrappedWithPayloadInstruction");
    expect(events[0].attributes.status).toBe("completed");
  });

  it("should map a fast transfer order protocol (e.g Fast Transfer - Method ExecuteFastOrderCctp)", async () => {
    const mockGetPostedMessage = getPostedMessage as jest.MockedFunction<typeof getPostedMessage>;
    mockGetPostedMessage.mockResolvedValueOnce({
      message: {
        sequence: 1n,
        emitterChain: 30,
        emitterAddress: Buffer.from([
          0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 112, 40, 124, 121, 238, 65, 197, 209, 223, 130, 89,
          205, 104, 186, 8, 144, 205, 56, 156, 71,
        ]),
        payload: Buffer.from("41QVZTrdrRxb", "base64"),
      } as any,
    });

    const programs = {
      HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb: [
        {
          instructions: ["ddb2b82bf7f85aa0"],
          vaaAccountIndex: 2,
        },
        {
          instructions: ["b0261e11e64ece9d"],
          vaaAccountIndex: 4,
        },
      ],
    };
    const tx = {
      blockTime: 1724706142,
      meta: {
        computeUnitsConsumed: 163056,
        err: null,
        fee: 5000,
        innerInstructions: [
          {
            index: 0,
            instructions: [
              { accounts: [4, 5, 3], data: "3sgej3QLG55Z", programIdIndex: 27, stackHeight: 2 },
              {
                accounts: [4, 3],
                data: "bmb5c5CAGhthE9v8eJxnk1wUdo9Uq18S7wCSJcUbzKeWfEH",
                programIdIndex: 27,
                stackHeight: 2,
              },
              { accounts: [0, 9], data: "3Bxs4HanWsHUZCbH", programIdIndex: 26, stackHeight: 2 },
              {
                accounts: [7, 1, 14, 8, 0, 9, 28, 26, 29],
                data: "B8WHcY2deAdqdmcCEmopDLnsGfKbi6ARtCKC5hMxCJZ4qXgRA1vCgCcxErD52TXWk3Pc7hb71UdFhNjeaFiMpeQKndksP2NFHBzcEby9tHLP6yRqsuE6FqpZjqE5rA8asMrUPJumAx88B4Tb95ycjdYS3NbP58FVVVwBpmkwWbQZa69XHZb4teQ9cpycHEhRpwStmWdw2BmwYWTyCixf3yegqz5jBzwYL3T59oYtNPygpn2b1JUy3cnoZDeysm8ATa49QxZNj87EYZCwM4KwjF3VsVyZqdyWF8wt8ZhmgwSWt8KzyNHKk5Q5mc39pk4zUmmdnDQK7G3u5eybFVuSQy8bA6mU4aYvyg2zZwvWGtQqJTAhBUwkKriCMrtnFYbiTs6mUwk8rdGk7HMqwHauHtg9iPSLtk9Uca6AZ5eYr3Nh9KeoSUm5qa9P9oMhJZW",
                programIdIndex: 18,
                stackHeight: 2,
              },
              { accounts: [0, 1], data: "3Bxs4Kfc3ADriqgs", programIdIndex: 26, stackHeight: 3 },
              { accounts: [1], data: "9krTDSgmoZrQBxWK", programIdIndex: 26, stackHeight: 3 },
              {
                accounts: [1],
                data: "SYXsBvR59WTsF4KEVN8LCQ1X9MekXCGPPNo3Af36taxCQBED",
                programIdIndex: 26,
                stackHeight: 3,
              },
              {
                accounts: [14, 0, 19, 4, 11, 20, 21, 22, 12, 10, 2, 25, 24, 27, 26, 23, 24],
                data: "CyFvRa11cBLDpjJjoaxy8imSGz3CNyx9QhpidQ4NAYtrpnAKwD9RY2pw84kVtRWRoyxdgyL4AH9WUBMRkGWGV7BHaaVKzkqvZFrP2m6VohvfGipPyWN",
                programIdIndex: 24,
                stackHeight: 2,
              },
              { accounts: [4, 10, 14], data: "714EubjMsC3q", programIdIndex: 27, stackHeight: 3 },
              {
                accounts: [0, 19, 11, 2, 24, 26],
                data: "7qMYLq4qUQBC94ph5cgKMqaBwGddVPprVRr9KKBQAFxVnE7xEpunQHhaUyhMpYx5dRAHjpA6pfSVbM5ADVthvdGmZgYPq6ECqKe6gv79yyVHNekCsBxRAyu23yFqJffmekwtKJ5U5Rm46HoqPcXPPSScgUwQJsakveTFzfJkLSd5WUzsiAzz1m9ytqFqgFuhUjvsX3AaCaqgCTYtoPgkSg64Rzi9giETpSCBJQ6NQtcSjzCCtJSdLETatvs13ivUbwkXeiePH5UdgkTJ6qW7QrmJ12L6LBFKdg",
                programIdIndex: 25,
                stackHeight: 3,
              },
              {
                accounts: [0, 2],
                data: "111184n6VJMYL8cUvJtKu66h1PA5AuP5c1YAiePkgYRo4DRGGzmGkRpowtZ9qczGP59b7j",
                programIdIndex: 26,
                stackHeight: 4,
              },
              {
                accounts: [23],
                data: "EVM9wLnauu9DWUq4iuSUfm45YY2L5ZHcZgpzQCXN3P2DxeDsSzmpM11FHkXrwaz9qEYLSVP2ZUDwG5nJWPp6GgRwR6CpVyMAn4iWnLmeeuNriPyriBAhWt8E6KPspACCeobGSmoK2TRDCoYjQ2wm9gScJJ3u7BtByKeWZTo57AW7LXGHywVQNgXqVKahkXChVx3U5EriDTWksajC61f1CVDHM2FBLdMuYcZhrUMssPb23bC9NM1wKd5oYzTZF62ecwJCo4LCRoWn",
                programIdIndex: 24,
                stackHeight: 3,
              },
              { accounts: [4, 6, 14], data: "A", programIdIndex: 27, stackHeight: 2 },
            ],
          },
        ],
        loadedAddresses: { readonly: [], writable: [] },
        logMessages: [
          "Program HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb invoke [1]",
          "Program log: Instruction: ExecuteFastOrderCctp",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: Transfer",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 163008 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: SetAuthority",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 2792 of 156155 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program data: SofnBahqwnVIBG1/eMiOJ/8X0gupUoNlUsMz3f7MgLUyz8L/H66UtxALDerq1eCJ2JOgEVMoCmGfdpxuqIG/20FSPJAz9hFWHgACAwAAAAA=",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth invoke [2]",
          "Program log: Sequence: 3",
          "Program 11111111111111111111111111111111 invoke [3]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [3]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [3]",
          "Program 11111111111111111111111111111111 success",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth consumed 27068 of 139922 compute units",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [2]",
          "Program log: Instruction: DepositForBurnWithCaller",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [3]",
          "Program log: Instruction: Burn",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4753 of 81596 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd invoke [3]",
          "Program log: Instruction: SendMessageWithCaller",
          "Program 11111111111111111111111111111111 invoke [4]",
          "Program 11111111111111111111111111111111 success",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd consumed 16752 of 71084 compute units",
          "Program return: CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd +ggBAAAAAAA=",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [3]",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 3632 of 50364 compute units",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 62014 of 106678 compute units",
          "Program return: CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 +ggBAAAAAAA=",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: CloseAccount",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3015 of 41974 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb consumed 163056 of 200000 compute units",
          "Program HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb success",
        ],
        postBalances: [
          891454820, 3765360, 2923200, 2422080, 0, 2039280, 11240800, 1057920, 946560, 226568224,
          318749174587, 2512560, 1795680, 1141440, 2157600, 3215520, 1183200, 1642560, 1141440, 0,
          1649520, 1197120, 1405920, 0, 1141440, 1141440, 1, 934087680, 1169280, 1009200,
        ],
        postTokenBalances: [
          {
            accountIndex: 5,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "5umD36Tb37L3s7UGaRmyuLSxJ8JEukRWkFWiumgsVdcU",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "806200000",
              decimals: 6,
              uiAmount: 806.2,
              uiAmountString: "806.2",
            },
          },
        ],
        preBalances: [
          898148480, 0, 0, 2422080, 2039280, 2039280, 9201520, 1057920, 946560, 226568124,
          318749174587, 2512560, 1795680, 1141440, 2157600, 3215520, 1183200, 1642560, 1141440, 0,
          1649520, 1197120, 1405920, 0, 1141440, 1141440, 1, 934087680, 1169280, 1009200,
        ],
        preTokenBalances: [
          {
            accountIndex: 4,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "5r8Gfm2VXxV6BnXvkNqshyD12kMMkXGF9jSLVJowWtWJ",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "103050500",
              decimals: 6,
              uiAmount: 103.0505,
              uiAmountString: "103.0505",
            },
          },
          {
            accountIndex: 5,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "5umD36Tb37L3s7UGaRmyuLSxJ8JEukRWkFWiumgsVdcU",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "801249500",
              decimals: 6,
              uiAmount: 801.2495,
              uiAmountString: "801.2495",
            },
          },
        ],
        rewards: [],
        status: { Ok: null },
      },
      slot: 286014729,
      transaction: {
        message: {
          header: {
            numReadonlySignedAccounts: 0,
            numReadonlyUnsignedAccounts: 17,
            numRequiredSignatures: 1,
          },
          staticAccountKeys: [
            "RoXd4qyJU6D1bA2WyG5i5vM5nL8gnEDAJ95Nc29WNmL",
            "CNaBUfHG3jgHV3aWaG6eLYFe9qgxGGyntTohYXkhJ7kD",
            "4BGub8VugkchHK1MDz1AxhbQ9eUvuYDbgT4mZoFS7Hfg",
            "5r8Gfm2VXxV6BnXvkNqshyD12kMMkXGF9jSLVJowWtWJ",
            "GY4yxpXNoxTZHSrtDw1Nvkqo1zG1Yw4fUfSoUttgyYgK",
            "7oA8vSPxsuHhub1SkNrikm9RE8DUhB2uqeTZ3GXbjQwf",
            "5umD36Tb37L3s7UGaRmyuLSxJ8JEukRWkFWiumgsVdcU",
            "2yVjuQwpsvdsrywzsJJVs9Ueh4zayyo5DYJbBNc3DDpn",
            "8xTgiBKwaSWx5iEPrhACqn4j5JXuQnDUt4sNUq8ZaN7t",
            "9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy",
            "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            "BWrwSWjbikT3H7qHAkUEbLmwDQoB4ZDJ4wcSEhSPTZCu",
            "72bvEFk2Usi2uYc1SnaTNhBcQPc6tiJWXr9oKk7rkd4C",
            "HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb",
            "8sLeDrpnUfSv69KXzKMKMVTpxP7D8iPue5QrHJgyu5XP",
            "25dJPCKy6Q4FFfDyZt6GdMVjZrDBu7bht19ZobSWrABw",
            "3NaMGAyMcXZJ7QvdmMCh9QjP7ShBksS7gh4Jje3Z4jpY",
            "GA6fw8zJxAMNBpBDidtGR2qRWsk4xAzm7tWUSKhnRGCu",
            "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
            "X5rMYSBWMqeWULSdDKXXATBjqk9AJF8odHpYJYeYA9H",
            "Afgq3BHEfCE7d78D2XE9Bfyu2ieDqvE24xX8KDwreBms",
            "REzxi9nX3Eqseha5fBiaJhTC6SFJx4qJhP83U4UCrtc",
            "DBD8hAwLDRQkTsu6EqviaYNGKPnsAMmQonxf7AH8ZcFY",
            "CNfZLeeL4RUxwfPnjA3tLiQt4y43jp4V7bMpga673jf9",
            "CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3",
            "CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd",
            "11111111111111111111111111111111",
            "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            "SysvarC1ock11111111111111111111111111111111",
            "SysvarRent111111111111111111111111111111111",
          ],
          recentBlockhash: "EwTXpGzmom4wk4YyAn7wMb3c2cA82g2E7yxVqMnViTgu",
          compiledInstructions: [
            {
              programIdIndex: 13,
              accountKeyIndexes: [
                0, 1, 2, 14, 15, 3, 4, 16, 5, 5, 5, 6, 17, 7, 8, 9, 18, 10, 19, 11, 20, 21, 22, 12,
                23, 24, 25, 26, 27, 28, 29,
              ],
              data: { type: "Buffer", data: [176, 38, 30, 17, 230, 78, 206, 157] },
            },
          ],
          addressTableLookups: [],
          accountKeys: [
            "RoXd4qyJU6D1bA2WyG5i5vM5nL8gnEDAJ95Nc29WNmL",
            "CNaBUfHG3jgHV3aWaG6eLYFe9qgxGGyntTohYXkhJ7kD",
            "4BGub8VugkchHK1MDz1AxhbQ9eUvuYDbgT4mZoFS7Hfg",
            "5r8Gfm2VXxV6BnXvkNqshyD12kMMkXGF9jSLVJowWtWJ",
            "GY4yxpXNoxTZHSrtDw1Nvkqo1zG1Yw4fUfSoUttgyYgK",
            "7oA8vSPxsuHhub1SkNrikm9RE8DUhB2uqeTZ3GXbjQwf",
            "5umD36Tb37L3s7UGaRmyuLSxJ8JEukRWkFWiumgsVdcU",
            "2yVjuQwpsvdsrywzsJJVs9Ueh4zayyo5DYJbBNc3DDpn",
            "8xTgiBKwaSWx5iEPrhACqn4j5JXuQnDUt4sNUq8ZaN7t",
            "9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy",
            "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            "BWrwSWjbikT3H7qHAkUEbLmwDQoB4ZDJ4wcSEhSPTZCu",
            "72bvEFk2Usi2uYc1SnaTNhBcQPc6tiJWXr9oKk7rkd4C",
            "HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb",
            "8sLeDrpnUfSv69KXzKMKMVTpxP7D8iPue5QrHJgyu5XP",
            "25dJPCKy6Q4FFfDyZt6GdMVjZrDBu7bht19ZobSWrABw",
            "3NaMGAyMcXZJ7QvdmMCh9QjP7ShBksS7gh4Jje3Z4jpY",
            "GA6fw8zJxAMNBpBDidtGR2qRWsk4xAzm7tWUSKhnRGCu",
            "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
            "X5rMYSBWMqeWULSdDKXXATBjqk9AJF8odHpYJYeYA9H",
            "Afgq3BHEfCE7d78D2XE9Bfyu2ieDqvE24xX8KDwreBms",
            "REzxi9nX3Eqseha5fBiaJhTC6SFJx4qJhP83U4UCrtc",
            "DBD8hAwLDRQkTsu6EqviaYNGKPnsAMmQonxf7AH8ZcFY",
            "CNfZLeeL4RUxwfPnjA3tLiQt4y43jp4V7bMpga673jf9",
            "CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3",
            "CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd",
            "11111111111111111111111111111111",
            "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            "SysvarC1ock11111111111111111111111111111111",
            "SysvarRent111111111111111111111111111111111",
          ],
        },
        signatures: [
          "2mTostoa2unw1wJ6Vwnm16cgYR7gL71XznYDPuhUrvh8YTBDH9T3YfEQqe9e95nRCB3cRspuxho3ZqDyeNtHDfUF",
        ],
      },
      version: 0,
      chain: "solana",
      chainId: 1,
    } as any as solana.Transaction;

    const events = await solanaTransferRedeemedMapper(tx, { programs });

    expect(events).toHaveLength(1);
    expect(events[0].name).toBe("transfer-redeemed");
    expect(events[0].address).toBe("HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb");
    expect(events[0].chainId).toBe(1);
    expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
    expect(events[0].blockHeight).toBe(BigInt(tx.slot));
    expect(events[0].blockTime).toBe(tx.blockTime);
    expect(events[0].attributes.methodsByAddress).toBe("executeFastOrderCctp");
    expect(events[0].attributes.status).toBe("completed");
    expect(events[0].attributes.emitterChain).toBe(30);
    expect(events[0].attributes.emitterAddress).toBe(
      "00000000000000000000000070287c79ee41c5d1df8259cd68ba0890cd389c47"
    );
    expect(events[0].attributes.sequence).toBe(1);
  });

  it("should map a fast transfer order protocol (e.g Slow Transfer - Method PrepareOrderResponseCctp)", async () => {
    const mockGetPostedMessage = getPostedMessage as jest.MockedFunction<typeof getPostedMessage>;
    mockGetPostedMessage.mockResolvedValueOnce({
      message: {
        sequence: 11n,
        emitterChain: 23,
        emitterAddress: Buffer.from([
          0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 112, 40, 124, 121, 238, 65, 197, 209, 223, 130, 89,
          205, 104, 186, 8, 144, 205, 56, 156, 71,
        ]),
        payload: Buffer.from("41QVZTrdrRxb", "base64"),
      } as any,
    });

    const programs = {
      HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb: [
        {
          instructions: ["ddb2b82bf7f85aa0"],
          vaaAccountIndex: 2,
        },
        {
          instructions: ["b0261e11e64ece9d"],
          vaaAccountIndex: 4,
        },
      ],
    };
    const tx = {
      blockTime: 1724523253,
      meta: {
        computeUnitsConsumed: 434090,
        err: null,
        fee: 255000,
        innerInstructions: [
          {
            index: 0,
            instructions: [
              {
                accounts: [0, 1],
                data: "1111b9Gj3hkyQeSwxaBGzYNzQnTfUvoaAqjzZoWNUvVLcBRf4Xv43etxnNp7fHJsfHmRf",
                programIdIndex: 33,
                stackHeight: 2,
              },
              {
                accounts: [0, 2],
                data: "11119os1e9qSs2u7TsThXqkBSRVFxhmYaFKFZ1waB2X7armDmvK3p5GmLdUxYdg3h7QSrL",
                programIdIndex: 33,
                stackHeight: 2,
              },
              {
                accounts: [2, 16],
                data: "6auA4gTs5KC3Wy7cyiUc3ZiKsyFkZgzqxfaT5jcc1vYT3",
                programIdIndex: 32,
                stackHeight: 2,
              },
              {
                accounts: [
                  0, 24, 25, 18, 3, 30, 33, 13, 31, 26, 27, 28, 19, 14, 17, 20, 32, 29, 30,
                ],
                data: "CaZrSwak5ygo6g7VAFXuyEFdRcnXVqDzsq9Er7mmCJeLaPfRXG7h61RS5iHQgBtziwqpmuthgEyn3GNaVH9e6Vx5EDwbauidgtjGNUZ5wkNd8D4BhDExgVUnkLALA5xzXqJCXnNK1XREFHTUf2HnymnDZ9hWR2sAvexRUs61YQQowADe95bBaqZLw9k6LE4TVAaF1tRP9F7o4989ePVSqczdtWJbuAxyTQfFFgkvFsYfvBxDPJ8SEWjVuPBiDYaP6ySC5RdPi5f5yHDxHoVYdHaQ6WY69UoN63wRXUp5cBKynbrXwUNEmFVQRwVU7sbKZM64CmnwLR3mWnDgSW2PCcMgFFPRnhiGsKJx2PVCzPPcNJ1ur7WvrQ9JWs4ACZ6rS8B7Bki6XKJC9ZeUr25BRdzcoiHdTRmTDveRu1bJ3WkTrxiEDZE9XKd2EvhkPU5V3VsNQtY6VMZoGq18kR8bksY1mywq9Mg59y2damevbHMiPXmbNBAmCEFC3AXtbEppFBvfX8ucvr7xbxc9Rdm7LFHab4",
                programIdIndex: 31,
                stackHeight: 2,
              },
              {
                accounts: [25, 26, 27, 28, 19, 14, 17, 20, 32, 29, 30],
                data: "26qRba3s3oVXnyz8M8r1dZqrhQ2YF5XPpfnjsjGsHoPEdTFUwEG1CRGDuc9A9naT3jmfDiyhysAbxyQtZ5QEVdLwGF34P9YPiHNrA4xbBripTLgsvXA5tWUccmxogTVnmKyZQZJHWw8B6JkTttcZ8Lrbdi8TYsry8wjqFSrTcdP3y6huWZjFFpNdumPkQaGH3inRPn4CYAEXDYZgZBtHXJ3zmStjqpSaJZscJ4XeoZMkWEmvWpoWdjRu",
                programIdIndex: 30,
                stackHeight: 3,
              },
              { accounts: [20, 17, 28], data: "3Jty2bi8FBFm", programIdIndex: 32, stackHeight: 4 },
              {
                accounts: [29],
                data: "2qWhKzSZDTHhUyxxdiKMXN7bTXAsqASouTbqvp5YSLfiFLhpsCV5jxwib2H61SehT47VkUo4EtGezc6FJYzwgmL3LMT4vcg9QR2nAaL4puW1iE6XEqoWwJwmi",
                programIdIndex: 30,
                stackHeight: 4,
              },
              {
                accounts: [13],
                data: "51QXw9Nm2rScqJP5v7j9KGMFdHNYDgdHUMCtabz7CGWmQCkzMoScx6S3e1XU6Zx2XF5FmL3ag2WRmr5PMaB3t8dQfaXDGiZRNmxppJqD3yU8LXBzBrs9xmYu3c758851Bm4aNeKptBNDjeL1v5Z7ebHre8CPLN3kAXqBdgRr8cv8j1hq3nVAJRjvR9jV5RMHPKChFBMYMaHQAyXuPxCmzZg96y1MktZUMRHZNVm9586fVcVZgp4WxSpkzVvWBy1ySQkc9xxoRUV6WungR6VxGcStAubEfrDHWgBD8ktoK9NFUX6W2s6gYEqk",
                programIdIndex: 31,
                stackHeight: 3,
              },
              { accounts: [17, 2, 24], data: "3Jty2bi8FBFm", programIdIndex: 32, stackHeight: 2 },
            ],
          },
          {
            index: 1,
            instructions: [
              {
                accounts: [0, 7],
                data: "11112o2d9BYmobGo5ZCscM2krbx6UhFbM6gD7aJouFLt6LuUVJL7KVx2teSAV2Jevo8U3F",
                programIdIndex: 33,
                stackHeight: 2,
              },
              { accounts: [2, 6, 1], data: "3mfhjUFPrrW3", programIdIndex: 32, stackHeight: 2 },
              {
                accounts: [2, 1],
                data: "bmb5c5CAGhthE9v8eJxnk1wUdo9Uq18S7wCSJcUbzKeWfEH",
                programIdIndex: 32,
                stackHeight: 2,
              },
              { accounts: [0, 23], data: "3Bxs4HanWsHUZCbH", programIdIndex: 33, stackHeight: 2 },
              {
                accounts: [21, 4, 24, 22, 0, 23, 37, 33, 38],
                data: "LCZywoxFdEvKzjGVsNBByWK7C61Sfmy2iC9UgMkg3kqh5Vme34FPERi9J5cyh5iK2MLMoe33TvVtVFCFavMjxLbZbFqremWdooW3yyzPZRwfERCeDicwTLGvJ5kbXbaWAMdQoaq3XXU21PrE5N1MWmmAZxkV7VZKB5P3ZrkfnhniRZTUxT2L6BEkzVBuPthsL6L68EbKAWVciwKEY7tvm2RSMBJWru9HX3Jq4Cq64ndnzXwsirAYfErmF6oK6tdJuaKk3ekNu9no1jTr1vrXCmqia6EcS8hTkrNMnw1EU4LJg891KzRuCN5J1o6CNBVy7ftEoTtbYsq6o2rufyQfN5tGhCa4rEJuDmcf928gdBKTESNoi",
                programIdIndex: 34,
                stackHeight: 2,
              },
              { accounts: [0, 4], data: "3Bxs4Z3B2ZKc9hmy", programIdIndex: 33, stackHeight: 3 },
              { accounts: [4], data: "9krTDH9oTmqj5JFq", programIdIndex: 33, stackHeight: 3 },
              {
                accounts: [4],
                data: "SYXsBvR59WTsF4KEVN8LCQ1X9MekXCGPPNo3Af36taxCQBED",
                programIdIndex: 33,
                stackHeight: 3,
              },
              {
                accounts: [24, 0, 35, 2, 18, 26, 36, 28, 19, 16, 5, 31, 30, 32, 33, 29, 30],
                data: "CyFvRa11cBLL1YPrajA1BMWEYXy5zHmWRYHwTdetGs4xnPgQqqAd8bSvLUeseKRsrZsB7F21XYwosSPXZaYsrnM37EPpCDFLg5CkE9D3VqZrRVmPrKY",
                programIdIndex: 30,
                stackHeight: 2,
              },
              { accounts: [2, 16, 24], data: "7BmcZEMXQVkT", programIdIndex: 32, stackHeight: 3 },
              {
                accounts: [0, 35, 18, 5, 30, 33],
                data: "7qMYLq4qUQBCGf5hVZ5L7oVPEu61pvhDesFoFqCaUQzV18rvzemhTy3T79rV6pNUVTvD3L9wA2uUM3S4yajjFewpJBZD7eirAUr9LvAFqRGutw22KuZtwhu1QrEk1ATZrvhy9noHX5SreXfghDeLJc28RfiwdX6dbR19YqUZsdQ6Ez1eDN3gLjkTmc7FNkh58zeJTBmGQkoDamfki2xeVeYwZxLw6A8GKYHioHBKAvVftjVM8cGJM6wYMQLiHn6sDVNWXKVMKCY6hrivDpddMsagfTEVdQPzZc",
                programIdIndex: 31,
                stackHeight: 3,
              },
              {
                accounts: [0, 5],
                data: "111184n6VJMYL8cUvJtKu66h1PA5AuP5c1YAiePkgYRo4DRGGzmGkRpowtZ9qczGP59b7j",
                programIdIndex: 33,
                stackHeight: 4,
              },
              {
                accounts: [29],
                data: "EVM9wLnauu9DWUq4iuSUfkqPkq5eEm86JFEyFsUYPDnrrxcTtH9jiNLuDEcriH2wGrA5bNSTiX6n1iEv9WSCsokJjjjU7Nrk5BV9mKTZqMckS9p2AQ28XGk2UbFHspp4Z1SdmPt4CcE1xBbKSSHbWi6mLXQY3bMk7KLSKxSAmmLBdeH96CRxi3rgvscLmHoRwTWhxvEefsjYkdqWfv3bQ4tGpq7ZM5WAPtwmiP8L5na2umSYCT59JEDu82hAk1gdHGeB4cTRhVci",
                programIdIndex: 30,
                stackHeight: 3,
              },
              { accounts: [2, 0, 24], data: "A", programIdIndex: 32, stackHeight: 2 },
            ],
          },
        ],
        loadedAddresses: {
          readonly: [
            "8sLeDrpnUfSv69KXzKMKMVTpxP7D8iPue5QrHJgyu5XP",
            "CFtn7PC5NsaFAuG65LwvhcGVD2MiqSpMJ7yvpyhsgJwW",
            "Afgq3BHEfCE7d78D2XE9Bfyu2ieDqvE24xX8KDwreBms",
            "REzxi9nX3Eqseha5fBiaJhTC6SFJx4qJhP83U4UCrtc",
            "DBD8hAwLDRQkTsu6EqviaYNGKPnsAMmQonxf7AH8ZcFY",
            "CNfZLeeL4RUxwfPnjA3tLiQt4y43jp4V7bMpga673jf9",
            "CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3",
            "CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd",
            "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            "11111111111111111111111111111111",
            "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
            "X5rMYSBWMqeWULSdDKXXATBjqk9AJF8odHpYJYeYA9H",
            "BWyFzH6LsnmDAaDWbGsriQ9SiiKq1CF6pbH4Ye3kzSBV",
            "SysvarC1ock11111111111111111111111111111111",
            "SysvarRent111111111111111111111111111111111",
          ],
          writable: [
            "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            "HUXc7MBf55vWrrkevVbmJN8HAyfFtjLcPLBt9yWngKzm",
            "BWrwSWjbikT3H7qHAkUEbLmwDQoB4ZDJ4wcSEhSPTZCu",
            "72bvEFk2Usi2uYc1SnaTNhBcQPc6tiJWXr9oKk7rkd4C",
            "FSxJ85FXVsXSr51SeWf9ciJWTcRnqKFSmBgRDeL3KyWw",
            "2yVjuQwpsvdsrywzsJJVs9Ueh4zayyo5DYJbBNc3DDpn",
            "8xTgiBKwaSWx5iEPrhACqn4j5JXuQnDUt4sNUq8ZaN7t",
            "9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy",
          ],
        },
        logMessages: [
          "Program HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb invoke [1]",
          "Program log: Instruction: PrepareOrderResponseCctp",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: InitializeAccount3",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4241 of 451684 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd invoke [2]",
          "Program log: Instruction: ReceiveMessage",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [3]",
          "Program log: Instruction: HandleReceiveMessage",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [4]",
          "Program log: Instruction: Transfer",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 273035 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [4]",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 3632 of 265175 compute units",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 38337 of 298296 compute units",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd invoke [3]",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd consumed 2133 of 255763 compute units",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd success",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd consumed 188526 of 434974 compute units",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: Transfer",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 242511 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb consumed 263851 of 500000 compute units",
          "Program HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb success",
          "Program HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb invoke [1]",
          "Program log: Instruction: SettleAuctionNoneCctp",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: Transfer",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 192155 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: SetAuthority",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 2792 of 185304 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program data: PZeDql/L25ME+0p1ebbceTR7u5OENZwS1+jmT4jXeNTDhf5CgD3XngABoj1kYcguq9I18VpA+ptgWzp4E5/ru0B/gzILZFCMTnFN/rtXAAAAAAECBgAAAA==",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth invoke [2]",
          "Program log: Sequence: 2",
          "Program 11111111111111111111111111111111 invoke [3]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [3]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [3]",
          "Program 11111111111111111111111111111111 success",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth consumed 27068 of 168853 compute units",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [2]",
          "Program log: Instruction: DepositForBurnWithCaller",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [3]",
          "Program log: Instruction: Burn",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4753 of 110527 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd invoke [3]",
          "Program log: Instruction: SendMessageWithCaller",
          "Program 11111111111111111111111111111111 invoke [4]",
          "Program 11111111111111111111111111111111 success",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd consumed 16752 of 100015 compute units",
          "Program return: CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd hQUBAAAAAAA=",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [3]",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 3632 of 79295 compute units",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 62014 of 135609 compute units",
          "Program return: CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 hQUBAAAAAAA=",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: CloseAccount",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3015 of 70914 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb consumed 169939 of 236149 compute units",
          "Program HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb success",
          "Program ComputeBudget111111111111111111111111111111 invoke [1]",
          "Program ComputeBudget111111111111111111111111111111 success",
          "Program ComputeBudget111111111111111111111111111111 invoke [1]",
          "Program ComputeBudget111111111111111111111111111111 success",
        ],
        postBalances: [
          974542620, 0, 0, 6598080, 3368640, 2923200, 2039280, 1566000, 1141440, 2818800, 1642560,
          1642560, 2637840, 0, 1426800, 1, 318749174587, 2039280, 2512560, 1795680, 2039280,
          1057920, 946560, 226388124, 2157600, 0, 1649520, 1197120, 1405920, 0, 1141440, 1141440,
          934087680, 1, 1141440, 0, 1197120, 1169280, 1009200,
        ],
        postTokenBalances: [
          {
            accountIndex: 6,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "AdAVF5KmmGmpNQhjY7FL96wZLEynD6Mx3VXJTZf2yFps",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "1471938125",
              decimals: 6,
              uiAmount: 1471.938125,
              uiAmountString: "1471.938125",
            },
          },
          {
            accountIndex: 17,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "8sLeDrpnUfSv69KXzKMKMVTpxP7D8iPue5QrHJgyu5XP",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: { amount: "0", decimals: 6, uiAmount: null, uiAmountString: "0" },
          },
          {
            accountIndex: 20,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "DBD8hAwLDRQkTsu6EqviaYNGKPnsAMmQonxf7AH8ZcFY",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "47572791323622",
              decimals: 6,
              uiAmount: 47572791.323622,
              uiAmountString: "47572791.323622",
            },
          },
        ],
        preBalances: [
          982655560, 0, 0, 6598080, 0, 0, 2039280, 0, 1141440, 2818800, 1642560, 1642560, 2637840,
          0, 1426800, 1, 318749174587, 2039280, 2512560, 1795680, 2039280, 1057920, 946560,
          226388024, 2157600, 0, 1649520, 1197120, 1405920, 0, 1141440, 1141440, 934087680, 1,
          1141440, 0, 1197120, 1169280, 1009200,
        ],
        preTokenBalances: [
          {
            accountIndex: 6,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "AdAVF5KmmGmpNQhjY7FL96wZLEynD6Mx3VXJTZf2yFps",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "1469738125",
              decimals: 6,
              uiAmount: 1469.738125,
              uiAmountString: "1469.738125",
            },
          },
          {
            accountIndex: 17,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "8sLeDrpnUfSv69KXzKMKMVTpxP7D8iPue5QrHJgyu5XP",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: { amount: "0", decimals: 6, uiAmount: null, uiAmountString: "0" },
          },
          {
            accountIndex: 20,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "DBD8hAwLDRQkTsu6EqviaYNGKPnsAMmQonxf7AH8ZcFY",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "47572891823622",
              decimals: 6,
              uiAmount: 47572891.823622,
              uiAmountString: "47572891.823622",
            },
          },
        ],
        rewards: [],
        status: { Ok: null },
      },
      slot: 285590214,
      transaction: {
        message: {
          header: {
            numReadonlySignedAccounts: 0,
            numReadonlyUnsignedAccounts: 8,
            numRequiredSignatures: 1,
          },
          staticAccountKeys: [
            "HuHz5VBEdgx4Zp3ib54tUekUorkR4AvLQkMCASh5agvx",
            "EmTPkGLH2LmghdXTdKQi5jtVUNAkHFaFxfHEF2vCdQ25",
            "8DMeu4XAG4WAwk4orZRqL2buMAs67eBNCYge2hFQTnh6",
            "6mnJWykJokxo5azmMRiXZ4AXsVS8zEG9FcEquQepZKB3",
            "BQHoc2wSMMu8ZmMDbrdjK6G2U1epENgAPj1jkdpkWNaN",
            "BBybKRw2MSkqeXSNmp3RboGHkgE84Y5iLR4JK4McvYsk",
            "BvKLsJ3s3T6jki9XmvvZN33VRtzrKMykyqBwm9q8xTWp",
            "LSsg7o2CTUSLY8uf2NrwixFstRytUS2bx5jSt8ujNYu",
            "HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb",
            "DPmn6hk8Y4txMmpAEtfG5UiYzFzRxVsniymDqQiJNsYK",
            "GA6fw8zJxAMNBpBDidtGR2qRWsk4xAzm7tWUSKhnRGCu",
            "9fd1z3q2Ef9s2NR2tM3856eJ2caDLNjJNJvb9wNpCbKa",
            "B34tcEHPuy6eJjQBHNzF7pnKKFjboPfd6wBCF1LVn1RQ",
            "6mH8scevHQJsyyp1qxu8kyAapHuzEE67mtjFDJZjSbQW",
            "3jziQYpnNe67yDduLX4VMNYL1VQai4kVHJKBfGKUHkK6",
            "ComputeBudget111111111111111111111111111111",
          ],
          recentBlockhash: "B2Fvtvvef4CVm8ANoHMTivxbwfs8814dVN4LjCdPEjGp",
          compiledInstructions: [
            {
              programIdIndex: 8,
              accountKeyIndexes: [
                0, 24, 9, 10, 11, 12, 1, 2, 16, 17, 25, 18, 3, 13, 26, 27, 28, 19, 14, 20, 29, 30,
                31, 32, 33,
              ],
              data: {
                type: "Buffer",
                data: [
                  221, 178, 184, 43, 247, 248, 90, 160, 248, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0,
                  0, 5, 0, 0, 0, 0, 0, 4, 108, 209, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 25, 51, 13,
                  16, 217, 204, 135, 81, 33, 142, 175, 81, 232, 136, 93, 5, 134, 66, 224, 138, 166,
                  95, 201, 67, 65, 154, 90, 213, 144, 4, 47, 214, 124, 151, 145, 253, 1, 90, 207,
                  83, 165, 76, 200, 35, 237, 184, 255, 129, 185, 237, 114, 46, 116, 231, 14, 213,
                  36, 100, 249, 151, 54, 155, 190, 253, 20, 29, 138, 45, 157, 211, 205, 21, 225,
                  242, 27, 55, 188, 225, 143, 69, 224, 233, 35, 178, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0, 175, 136, 208, 101, 231, 124, 140, 194, 35, 147, 39, 197, 237,
                  179, 164, 50, 38, 142, 88, 49, 244, 200, 71, 58, 14, 143, 176, 147, 202, 18, 151,
                  14, 214, 21, 219, 9, 247, 235, 187, 179, 208, 15, 64, 179, 226, 133, 225, 47, 64,
                  229, 201, 166, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0, 5, 253, 130, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 112, 40,
                  124, 121, 238, 65, 197, 209, 223, 130, 89, 205, 104, 186, 8, 144, 205, 56, 156,
                  71, 130, 0, 0, 0, 110, 4, 55, 128, 122, 40, 129, 221, 217, 112, 43, 137, 98, 153,
                  114, 48, 61, 17, 253, 98, 98, 220, 133, 134, 202, 116, 210, 111, 109, 4, 193, 70,
                  89, 75, 195, 243, 244, 109, 37, 106, 21, 219, 57, 122, 63, 77, 143, 211, 231, 177,
                  27, 86, 189, 98, 4, 215, 82, 111, 155, 5, 163, 232, 43, 163, 28, 203, 239, 197,
                  35, 143, 69, 135, 172, 35, 177, 65, 192, 187, 71, 4, 1, 233, 175, 42, 37, 219,
                  172, 29, 153, 181, 32, 75, 207, 5, 176, 79, 70, 71, 4, 172, 229, 167, 184, 11,
                  196, 152, 226, 111, 212, 113, 223, 17, 165, 25, 195, 2, 200, 52, 40, 95, 208, 145,
                  189, 197, 191, 93, 161, 141, 29, 27,
                ],
              },
            },
            {
              programIdIndex: 8,
              accountKeyIndexes: [
                0, 4, 5, 24, 6, 0, 1, 2, 7, 21, 22, 23, 34, 16, 35, 18, 26, 36, 28, 19, 29, 30, 31,
                32, 33, 37, 38,
              ],
              data: { type: "Buffer", data: [120, 236, 82, 121, 242, 118, 74, 161] },
            },
            {
              programIdIndex: 15,
              accountKeyIndexes: [],
              data: { type: "Buffer", data: [2, 32, 161, 7, 0] },
            },
            {
              programIdIndex: 15,
              accountKeyIndexes: [],
              data: { type: "Buffer", data: [3, 32, 161, 7, 0, 0, 0, 0, 0] },
            },
          ],
          addressTableLookups: [
            {
              accountKey: "4SeadipRDH6R1F4CmPdc1UqqWFADGMitT5FdyWX5X44t",
              readonlyIndexes: [4, 14, 10, 22, 11, 23, 13, 16, 17, 1, 9, 12, 24, 3, 2],
              writableIndexes: [18, 5, 15, 19, 20, 6, 7, 8],
            },
          ],
          accountKeys: [
            "HuHz5VBEdgx4Zp3ib54tUekUorkR4AvLQkMCASh5agvx",
            "EmTPkGLH2LmghdXTdKQi5jtVUNAkHFaFxfHEF2vCdQ25",
            "8DMeu4XAG4WAwk4orZRqL2buMAs67eBNCYge2hFQTnh6",
            "6mnJWykJokxo5azmMRiXZ4AXsVS8zEG9FcEquQepZKB3",
            "BQHoc2wSMMu8ZmMDbrdjK6G2U1epENgAPj1jkdpkWNaN",
            "BBybKRw2MSkqeXSNmp3RboGHkgE84Y5iLR4JK4McvYsk",
            "BvKLsJ3s3T6jki9XmvvZN33VRtzrKMykyqBwm9q8xTWp",
            "LSsg7o2CTUSLY8uf2NrwixFstRytUS2bx5jSt8ujNYu",
            "HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb",
            "DPmn6hk8Y4txMmpAEtfG5UiYzFzRxVsniymDqQiJNsYK",
            "GA6fw8zJxAMNBpBDidtGR2qRWsk4xAzm7tWUSKhnRGCu",
            "9fd1z3q2Ef9s2NR2tM3856eJ2caDLNjJNJvb9wNpCbKa",
            "B34tcEHPuy6eJjQBHNzF7pnKKFjboPfd6wBCF1LVn1RQ",
            "6mH8scevHQJsyyp1qxu8kyAapHuzEE67mtjFDJZjSbQW",
            "3jziQYpnNe67yDduLX4VMNYL1VQai4kVHJKBfGKUHkK6",
            "ComputeBudget111111111111111111111111111111",
          ],
        },
        signatures: [
          "3VXJGF7Xnxj5m167qVMqYb69GcqNAPraYJtZDWJMfKjCMwd3DbQu5wp81a9qgAcRPn2hpL41QHMncWGdxqkQE3gQ",
        ],
      },
      version: 0,
      chain: "solana",
      chainId: 1,
    } as any as solana.Transaction;

    const events = await solanaTransferRedeemedMapper(tx, { programs });

    expect(events).toHaveLength(1);
    expect(events[0].name).toBe("transfer-redeemed");
    expect(events[0].address).toBe("HtkeCDdYY4i9ncAxXKjYTx8Uu3WM8JbtiLRYjtHwaVXb");
    expect(events[0].chainId).toBe(1);
    expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
    expect(events[0].blockHeight).toBe(BigInt(tx.slot));
    expect(events[0].blockTime).toBe(tx.blockTime);
    expect(events[0].attributes.methodsByAddress).toBe("executeSlowOrderCctp");
    expect(events[0].attributes.status).toBe("completed");
    expect(events[0].attributes.emitterChain).toBe(23);
    expect(events[0].attributes.emitterAddress).toBe(
      "00000000000000000000000070287c79ee41c5d1df8259cd68ba0890cd389c47"
    );
    expect(events[0].attributes.sequence).toBe(11);
  });

  it("should map a mayan tx (e.g Mayan Swap)", async () => {
    const mockGetPostedMessage = getPostedMessage as jest.MockedFunction<typeof getPostedMessage>;
    mockGetPostedMessage.mockResolvedValueOnce({
      message: {
        sequence: 10364n,
        emitterChain: 23,
        emitterAddress: Buffer.from([
          0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 191, 95, 63, 101, 16, 42, 231, 69, 164, 139, 213, 33,
          209, 11, 171, 91, 240, 42, 158, 244,
        ]),
        payload: Buffer.from("41QVZTrdrRxb", "base64"),
      } as any,
    });

    const programs = {
      FC4eXxkyrMPTjiYUpp4EAnkmwMbQyZ6NDCh1kfLn6vsf: [
        {
          instructions: ["64"],
          vaaAccountIndex: 2,
        },
      ],
    };
    const tx = {
      blockTime: 1727886251,
      meta: {
        computeUnitsConsumed: 39778,
        err: null,
        fee: 205000,
        innerInstructions: [
          {
            index: 1,
            instructions: [
              {
                accounts: [0, 2],
                data: "111183uzQbvYqJhYsuUqDALMEpDcPoDyV3HRNAjKrwzyrnxKTJsDS2a637jUDf9G88hN8D",
                programIdIndex: 4,
                stackHeight: 2,
              },
              { accounts: [1, 3, 11], data: "3DYhamUpKWDM", programIdIndex: 16, stackHeight: 2 },
            ],
          },
        ],
        loadedAddresses: { readonly: [], writable: [] },
        logMessages: [
          "Program ComputeBudget111111111111111111111111111111 invoke [1]",
          "Program ComputeBudget111111111111111111111111111111 success",
          "Program FC4eXxkyrMPTjiYUpp4EAnkmwMbQyZ6NDCh1kfLn6vsf invoke [1]",
          "Program log: mayan-swap v0.2 (build  at 1708743373)",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: Transfer",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 164877 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program FC4eXxkyrMPTjiYUpp4EAnkmwMbQyZ6NDCh1kfLn6vsf consumed 39628 of 199850 compute units",
          "Program FC4eXxkyrMPTjiYUpp4EAnkmwMbQyZ6NDCh1kfLn6vsf success",
        ],
        postBalances: [
          10541196636, 2039280, 3814080, 2039280, 1, 25653059827, 2477760, 0, 3118080, 960480, 1,
          1000001, 1141440, 897840, 1169280, 1009200, 934087680,
        ],
        postTokenBalances: [
          {
            accountIndex: 1,
            mint: "3NZ9JMVBmGAqocybic2c7LQCJScmgsAZ6vQqTDzcqmJh",
            owner: "Dqfqb9Xr19tzQd5JjDwKnoK4RGrMa6AdaXzCjZJauhJP",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: { amount: "0", decimals: 8, uiAmount: null, uiAmountString: "0" },
          },
          {
            accountIndex: 3,
            mint: "3NZ9JMVBmGAqocybic2c7LQCJScmgsAZ6vQqTDzcqmJh",
            owner: "BSgCgeNT1nfrya8WJYBJcKGx4W7AQv6TiYHJGC46JAtV",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "60000000",
              decimals: 8,
              uiAmount: 0.6,
              uiAmountString: "0.6",
            },
          },
        ],
        preBalances: [
          10545215716, 2039280, 0, 2039280, 1, 25653059827, 2477760, 0, 3118080, 960480, 1, 1000001,
          1141440, 897840, 1169280, 1009200, 934087680,
        ],
        preTokenBalances: [
          {
            accountIndex: 1,
            mint: "3NZ9JMVBmGAqocybic2c7LQCJScmgsAZ6vQqTDzcqmJh",
            owner: "Dqfqb9Xr19tzQd5JjDwKnoK4RGrMa6AdaXzCjZJauhJP",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "60000000",
              decimals: 8,
              uiAmount: 0.6,
              uiAmountString: "0.6",
            },
          },
          {
            accountIndex: 3,
            mint: "3NZ9JMVBmGAqocybic2c7LQCJScmgsAZ6vQqTDzcqmJh",
            owner: "BSgCgeNT1nfrya8WJYBJcKGx4W7AQv6TiYHJGC46JAtV",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: { amount: "0", decimals: 8, uiAmount: null, uiAmountString: "0" },
          },
        ],
        rewards: [],
        status: { Ok: null },
      },
      slot: 293299620,
      transaction: {
        message: {
          header: {
            numReadonlySignedAccounts: 0,
            numReadonlyUnsignedAccounts: 13,
            numRequiredSignatures: 1,
          },
          accountKeys: [
            "7dm9am6Qx7cH64RB99Mzf7ZsLbEfmXM7ihXXCvMiT2X1",
            "Akoxxmb8M4bRXFUeoJKZSNDJfX4j6JXgnPCrBkYkdSX9",
            "BSgCgeNT1nfrya8WJYBJcKGx4W7AQv6TiYHJGC46JAtV",
            "DQVttusgHNk2WjLvDfNmiNDnqYvkusjGBSmY9cqvCKaB",
            "11111111111111111111111111111111",
            "3NZ9JMVBmGAqocybic2c7LQCJScmgsAZ6vQqTDzcqmJh",
            "42CxR7Lg2WtxRrH1uLzuEpqEZirkhLH5bSJejcGYZSr9",
            "461eEhHkU1vWAJfHRKFqLy27csP1dyBsU56EWRAM7h85",
            "6vLXjuyvimYA1FtHx33o8B9Hn7WZUFGAzietkN5PMws9",
            "74ouc5sSsfxjT6iDXwLeD3LB2ZyAP5vEgBui1JnSowCk",
            "ComputeBudget111111111111111111111111111111",
            "Dqfqb9Xr19tzQd5JjDwKnoK4RGrMa6AdaXzCjZJauhJP",
            "FC4eXxkyrMPTjiYUpp4EAnkmwMbQyZ6NDCh1kfLn6vsf",
            "mDE8wD13VPmMEsHFYJQfotytkKaioQpouuyZwzkUY2s",
            "SysvarC1ock11111111111111111111111111111111",
            "SysvarRent111111111111111111111111111111111",
            "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
          ],
          recentBlockhash: "DtdyoJFhevLweFzoXnNY7xYB3oo7dRWqU5SgEobChhTs",
          instructions: [
            { accounts: [], data: "3QCwqmHZ4mdq", programIdIndex: 10, stackHeight: null },
            {
              accounts: [0, 6, 8, 2, 11, 5, 5, 1, 3, 9, 7, 14, 13, 15, 4, 16],
              data: "21Khf38C9n467sgn3ay13Ae9Gy6mvhmNzixHawErdJ2zGF8etqrYvzQDuNc5tKP2gfNySVgLdZZF5DEhq5eWttYF6TL9PNJ",
              programIdIndex: 12,
              stackHeight: null,
            },
          ],
          indexToProgramIds: {},
          compiledInstructions: [
            {
              programIdIndex: 10,
              accountKeyIndexes: [],
              data: { type: "Buffer", data: [3, 64, 66, 15, 0, 0, 0, 0, 0] },
            },
            {
              programIdIndex: 12,
              accountKeyIndexes: [0, 6, 8, 2, 11, 5, 5, 1, 3, 9, 7, 14, 13, 15, 4, 16],
              data: {
                type: "Buffer",
                data: [
                  100, 255, 255, 254, 255, 49, 128, 145, 17, 86, 181, 86, 152, 153, 152, 28, 205,
                  60, 138, 178, 208, 160, 135, 38, 169, 173, 246, 133, 60, 62, 36, 236, 21, 192,
                  237, 182, 18, 29, 102, 224, 100, 86, 174, 3, 104, 37, 20, 206, 190, 28, 77, 135,
                  245, 63, 104, 186, 234, 115, 192, 44, 182, 83, 68, 140, 133, 167, 109, 36, 219,
                ],
              },
            },
          ],
        },
        signatures: [
          "3HxPToyFuStoBjecXNaavxeSuLhB1ViXbtGnvuTHPxUnS5Koc7ywbkMoaaLyD37URpvSYr4MFyNhrGUR9j7Zbumt",
        ],
      },
      version: "legacy",
      chain: "solana",
      chainId: 1,
    } as any as solana.Transaction;

    const events = await solanaTransferRedeemedMapper(tx, { programs });

    expect(events).toHaveLength(1);
    expect(events[0].name).toBe("transfer-redeemed");
    expect(events[0].address).toBe("FC4eXxkyrMPTjiYUpp4EAnkmwMbQyZ6NDCh1kfLn6vsf");
    expect(events[0].chainId).toBe(1);
    expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
    expect(events[0].blockHeight).toBe(BigInt(tx.slot));
    expect(events[0].blockTime).toBe(tx.blockTime);
    expect(events[0].attributes.methodsByAddress).toBe("MethodCreateAccount");
    expect(events[0].attributes.status).toBe("completed");
    expect(events[0].attributes.emitterChain).toBe(23);
    expect(events[0].attributes.emitterAddress).toBe(
      "000000000000000000000000bf5f3f65102ae745a48bd521d10bab5bf02a9ef4"
    );
    expect(events[0].attributes.sequence).toBe(10364);
  });
});
