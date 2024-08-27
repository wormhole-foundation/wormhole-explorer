import { expect, describe, it, jest, beforeEach } from "@jest/globals";
import { solanaLogCircleMessageSentMapper } from "../../../../src/infrastructure/mappers";
import { solana } from "../../../../src/domain/entities";

jest.mock("@coral-xyz/anchor", () => {
  return {
    web3: {
      PublicKey: jest.fn().mockImplementation((key) => {
        return {
          toString: () => key,
        };
      }),
    },
    Program: jest.fn().mockImplementation(() => {
      return {
        account: {
          messageSent: {
            fetch: jest.fn().mockImplementation(() => {
              return {
                message: Buffer.from([
                  0, 0, 0, 0, 0, 0, 0, 5, 0, 0, 0, 3, 0, 0, 0, 0, 0, 1, 8, 249, 166, 95, 201, 67,
                  65, 154, 90, 213, 144, 4, 47, 214, 124, 151, 145, 253, 1, 90, 207, 83, 165, 76,
                  200, 35, 237, 184, 255, 129, 185, 237, 114, 46, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                  0, 25, 51, 13, 16, 217, 204, 135, 81, 33, 142, 175, 81, 232, 136, 93, 5, 134, 66,
                  224, 138, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 198, 250, 122, 243, 190, 219, 173, 58, 61,
                  101, 243, 106, 171, 201, 116, 49, 177, 187, 228, 194, 210, 246, 224, 228, 124,
                  166, 2, 3, 69, 47, 93, 97, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 39, 54, 204, 38,
                  117, 99, 4, 129, 68, 112, 146, 224, 203, 180, 242, 130, 71, 65, 211, 92, 0, 0, 0,
                  0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 41, 32,
                  144, 128, 175, 52, 31, 83, 42, 213, 146, 59, 250, 101, 174, 113, 29, 159, 242, 70,
                  27, 253, 12, 190, 128, 107, 57, 220, 138, 221, 0, 9, 141, 162, 225, 204,
                ]),
              };
            }),
          },
        },
      };
    }),
  };
});

describe("solanaLogCircleMessageSentMapper", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("should map a token bridge source tx circle-message-sent event", async () => {
    const programs = {
      programs: { CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd: { vaaAccountIndex: 3 } },
      environment: "mainnet",
    };

    const tx = {
      blockTime: 1724682919,
      meta: {
        computeUnitsConsumed: 61825,
        err: null,
        fee: 14604,
        innerInstructions: [
          {
            index: 0,
            instructions: [
              { accounts: [2, 5, 0], data: "6ufgx5CkpFRq", programIdIndex: 14, stackHeight: 2 },
              {
                accounts: [0, 15, 4, 1, 8, 6],
                data: "7CQpu95dc956sR51iTLa1ScZhCVNe41VGhDZRxTAe3cLzaYeJRFQPPjTfk9aTJVQUfQ7Jma3dh2JH5SndF9GX9j7vSQTnTK4vLpY6uaFXe8bHc6GJmF4AT48ARaShRzVHFNKP69gf7nLTFbe4Vza5TPZaEFiaBHF6LwEQNAQXuLmQ74AetCPachqpn3naEQRhuS6wt1Y2Kh3e1ALmVBKDK3eTKJcKZAF7wENdcGM72cDoHVd5Bj1uq",
                programIdIndex: 9,
                stackHeight: 2,
              },
              {
                accounts: [0, 1],
                data: "111184n6VJMYL8cUvJtKu66h1PA5AuP5c1YAiePkgYRo4DRGGzmGkRpowtZ9qczGP59b7j",
                programIdIndex: 6,
                stackHeight: 3,
              },
              {
                accounts: [10],
                data: "EVM9wLnauu9DWUq4iuSUfkpJwyNrVhAkQCNefj8Ka7Zvtkszn54sCqrgTwCqdwHBfrKyzkspEownHVuHBiD26RiNc4Pvb8fWPAV8n7ibM7R2KnQPjk53Dd6GTPjq7wHP1FAfyNLE3NLKT6Fahx1bPu2uDNJy77d8vyepkNyWtMkifPeT3QNyuqfcrahekiLwEzQJFhC62XBH3PGNyhzGZ7XE9LphtQ1dFQScXBAtiE4CKuNuYRcmZRBSfxYdXCiRBaMfuPpmTMxB",
                programIdIndex: 8,
                stackHeight: 2,
              },
            ],
          },
        ],
        loadedAddresses: { readonly: [], writable: [] },
        logMessages: [
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [1]",
          "Program log: Instruction: DepositForBurn",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: Burn",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4753 of 49121 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd invoke [2]",
          "Program log: Instruction: SendMessage",
          "Program 11111111111111111111111111111111 invoke [3]",
          "Program 11111111111111111111111111111111 success",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd consumed 16715 of 38666 compute units",
          "Program return: CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd ewgBAAAAAAA=",
          "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [2]",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 3632 of 17982 compute units",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 61525 of 73830 compute units",
          "Program return: CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 ewgBAAAAAAA=",
          "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
          "Program ComputeBudget111111111111111111111111111111 invoke [1]",
          "Program ComputeBudget111111111111111111111111111111 success",
          "Program ComputeBudget111111111111111111111111111111 invoke [1]",
          "Program ComputeBudget111111111111111111111111111111 success",
        ],
        postBalances: [
          44609455, 2923200, 2039280, 1795680, 2512560, 318749174587, 1, 1649520, 1141440, 1141440,
          0, 1, 1405920, 1197120, 934087680, 0,
        ],
        postTokenBalances: [
          {
            accountIndex: 2,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "3DBoK4xXbff387iNkVmJt18ZXbQHjaBUXR9Xwy7KJdPM",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "4",
              decimals: 6,
              uiAmount: 0.000004,
              uiAmountString: "0.000004",
            },
          },
        ],
        preBalances: [
          47547259, 0, 2039280, 1795680, 2512560, 318749174587, 1, 1649520, 1141440, 1141440, 0, 1,
          1405920, 1197120, 934087680, 0,
        ],
        preTokenBalances: [
          {
            accountIndex: 2,
            mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            owner: "3DBoK4xXbff387iNkVmJt18ZXbQHjaBUXR9Xwy7KJdPM",
            programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            uiTokenAmount: {
              amount: "30000000004",
              decimals: 6,
              uiAmount: 30000.000004,
              uiAmountString: "30000.000004",
            },
          },
        ],
        rewards: [],
        status: { Ok: null },
      },
      slot: 285963118,
      transaction: {
        message: {
          header: {
            numReadonlySignedAccounts: 0,
            numReadonlyUnsignedAccounts: 10,
            numRequiredSignatures: 2,
          },
          accountKeys: [
            "3DBoK4xXbff387iNkVmJt18ZXbQHjaBUXR9Xwy7KJdPM",
            "DD27CfXRS5ThRVevPNeqySBBEiMqiATDnZ8ch8vK1uzX",
            "3voNtvg2HzTEXBUxB2ZE41DEaz7L2VaFMdKtNVsM82Vc",
            "72bvEFk2Usi2uYc1SnaTNhBcQPc6tiJWXr9oKk7rkd4C",
            "BWrwSWjbikT3H7qHAkUEbLmwDQoB4ZDJ4wcSEhSPTZCu",
            "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
            "11111111111111111111111111111111",
            "Afgq3BHEfCE7d78D2XE9Bfyu2ieDqvE24xX8KDwreBms",
            "CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3",
            "CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd",
            "CNfZLeeL4RUxwfPnjA3tLiQt4y43jp4V7bMpga673jf9",
            "ComputeBudget111111111111111111111111111111",
            "DBD8hAwLDRQkTsu6EqviaYNGKPnsAMmQonxf7AH8ZcFY",
            "Hazwi3jFQtLKc2ughi7HFXPkpDeso7DQaMR9Ks4afh3j",
            "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            "X5rMYSBWMqeWULSdDKXXATBjqk9AJF8odHpYJYeYA9H",
          ],
          recentBlockhash: "2LPpMmbQfFGXWkienmFfPesRiZtFYzGFe5zZYUCuWGNe",
          instructions: [
            {
              accounts: [0, 0, 15, 2, 4, 7, 13, 12, 3, 5, 1, 9, 8, 14, 6, 10, 8],
              data: "tfYNBY81Rd8gjyPMmiEKJAibRSTV9Pv824pPmwUnc2tKbBgi49WiceonFsPTby6APgPyqUv",
              programIdIndex: 8,
              stackHeight: null,
            },
            { accounts: [], data: "GhU9Ww", programIdIndex: 11, stackHeight: null },
            { accounts: [], data: "3dhkZTgNbFGb", programIdIndex: 11, stackHeight: null },
          ],
          indexToProgramIds: {},
          compiledInstructions: [
            {
              programIdIndex: 8,
              accountKeyIndexes: [0, 0, 15, 2, 4, 7, 13, 12, 3, 5, 1, 9, 8, 14, 6, 10, 8],
              data: {
                type: "Buffer",
                data: [
                  215, 60, 61, 46, 114, 55, 128, 176, 0, 172, 35, 252, 6, 0, 0, 0, 0, 0, 0, 0, 0, 0,
                  0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 107, 67, 248, 94, 254, 89, 162, 157, 203, 189, 122,
                  127, 47, 141, 224, 84, 254, 36, 213, 147,
                ],
              },
            },
            {
              programIdIndex: 11,
              accountKeyIndexes: [],
              data: { type: "Buffer", data: [2, 102, 32, 1, 0] },
            },
            {
              programIdIndex: 11,
              accountKeyIndexes: [],
              data: { type: "Buffer", data: [3, 144, 243, 0, 0, 0, 0, 0, 0] },
            },
          ],
        },
        signatures: [
          "3UZLAda9fz1oxxMc7Z4sPM5TyyJ7W2rEGF2tPnfWb9dY6swYrKkKi4GTF7Ps312nZXXGTpFzDPRV7hdvi3TYFmoL",
          "5F55EMDuuiToV7Zqz2Ymw23bvhnTfrSGR1ADH6UAVhTyLPeHtm2CT1j8ocwu2JZYuTZkukDmwLrw1mtW33wZBRg9",
        ],
      },
      version: "legacy",
      chain: "solana",
      chainId: 1,
    } as any as solana.Transaction;

    const events = await solanaLogCircleMessageSentMapper(tx, programs);

    expect(events).toHaveLength(1);
    expect(events[0].name).toBe("circle-message-sent");
    expect(events[0].address).toBe("CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd");
    expect(events[0].chainId).toBe(1);
    expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
    expect(events[0].blockHeight).toBe(BigInt(tx.slot));
    expect(events[0].blockTime).toBe(tx.blockTime);

    expect(events[0].attributes.sourceDomain).toBe("Solana");
    expect(events[0].attributes.destinationDomain).toBe("Arbitrum");
    expect(events[0].attributes.messageSender).toBe(
      "0xaf341f532ad5923bfa65ae711d9ff2461bfd0cbe806b39dc8add00098da2e1cc"
    );
    expect(events[0].attributes.destinationCaller).toBe(
      "0x0000000000000000000000000000000000000000000000000000000000000000"
    );
    expect(events[0].attributes.sender).toBe(
      "0xa65fc943419a5ad590042fd67c9791fd015acf53a54cc823edb8ff81b9ed722e"
    );
  });
});
