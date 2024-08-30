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

    expect(events[0].tags?.sourceDomain).toBe("Solana");
    expect(events[0].tags?.destinationDomain).toBe("Arbitrum");
    expect(events[0].tags?.protocol).toBe("cctp");
    expect(events[0].tags?.messageProtocol).toBe("wormhole");
  });
});

const tx = {
  blockTime: 1724793729,
  meta: {
    computeUnitsConsumed: 130356,
    err: null,
    fee: 45000,
    innerInstructions: [
      {
        index: 1,
        instructions: [
          { accounts: [4, 12, 3], data: "3NY1P7iid9M9", programIdIndex: 22, stackHeight: 2 },
          {
            accounts: [3, 0, 15, 4, 5, 6, 16, 17, 7, 8, 1, 25, 26, 22, 23, 18, 26, 21],
            data: "CyFvRa11cBLXge5zukvdZExjWNfQ3FQrcZJPztDYK8tP1BWSQSiVXC2DfbqTvoii1wgr7mDU1TaPmqJppfqCdaasfKZ6hiGB5CaFDZxsXgDi1SHS76T",
            programIdIndex: 26,
            stackHeight: 2,
          },
          { accounts: [4, 8, 3], data: "7XzBQ1xWXYnK", programIdIndex: 22, stackHeight: 3 },
          {
            accounts: [0, 15, 5, 1, 26, 23],
            data: "7qMYLq4qUQBC94ph5cgKMqaBwGddVPprVRr9KKBQAFxVnE7xEpunQHhaUyhMpYx5dRAHjpA6pfSVbM5ADVthvdGmZgYPq6ECqKe6gv79yyVHNekCsBxRAyu23yFqJffmekww4GxPFnQ1AgNtNcWW36ZoGC4fggGSyASExVRhj5ZLe6jjxNPjHSaRXfWPGj8PXa5ywvR1bgwbGy9jPaVBpZEPheNR6ySe97uYkoPGRUHE7EKNbW95nhUmPCKB4scJdZSGFCBMmmMFetpZzyryieKU5NeCZUBgYT",
            programIdIndex: 25,
            stackHeight: 3,
          },
          {
            accounts: [0, 1],
            data: "111184n6VJMYL8cUvJtKu66h1PA5AuP5c1YAiePkgYRo4DRGGzmGkRpowtZ9qczGP59b7j",
            programIdIndex: 23,
            stackHeight: 4,
          },
          {
            accounts: [18],
            data: "EVM9wLnauu9DWUq4iuSUfkvhztqu941cDavBD3d3nLhLjdj6UGFnocYeHQ596sZQRnYyNoZS87iv3DEHaeRMj5gbcU81JkYN6iK9JS2SrsK2PcV21YuPesaMPZN3AHUHg95XCd7JeNkJQHcYh13uDGKcGfHgRokabeJCncTdb7jo7y5ZSrzuvqpPgJBvaKDowh3drugLoux1v3ayS9LJXA9PzYxsBAHA74TKWtpjdrVUqvoJKLbzX9ACeew9tcB945SwyyihW4EP",
            programIdIndex: 26,
            stackHeight: 3,
          },
          { accounts: [0, 11], data: "3Bxs4HanWsHUZCbH", programIdIndex: 23, stackHeight: 2 },
          {
            accounts: [10, 2, 19, 9, 0, 11, 20, 21, 23],
            data: "rr9xssinGcGUEqXJFUA6kC6LAQvjvKMqnEWJBrQnyiXqBtJZnaV55QHUnMTvoG8WjP8sDgoqyVVFyjJAxt9aH2rfEkjJKLXmS",
            programIdIndex: 24,
            stackHeight: 2,
          },
          { accounts: [0, 2], data: "3Bxs4KdnJed8ya55", programIdIndex: 23, stackHeight: 3 },
          { accounts: [2], data: "9krTDSgjc5yvN3MR", programIdIndex: 23, stackHeight: 3 },
          {
            accounts: [2],
            data: "SYXsBvR59WTsF4KEVN8LCQ1X9MekXCGPPNo3Af36taxCQBED",
            programIdIndex: 23,
            stackHeight: 3,
          },
          { accounts: [4, 0, 3], data: "A", programIdIndex: 22, stackHeight: 2 },
        ],
      },
    ],
    loadedAddresses: { readonly: [], writable: [] },
    logMessages: [
      "Program ComputeBudget111111111111111111111111111111 invoke [1]",
      "Program ComputeBudget111111111111111111111111111111 success",
      "Program Awm2zSgzMGTRraAVjRvshqLehy7mJ2Qr3maURDsoDmwi invoke [1]",
      "Program log: mctp-swap v0.2 (build 24 at 1722037923)",
      "Program log: ./src/mctpswap/ctx.h:144: parse clock account",
      "Program log: ./src/mctpswap/ctx.h:107: parse rent account",
      "Program log: ./src/mctpswap/ctx.h:110: rent account cursor move forward",
      "Program log: ./src/mctpswap/ctx.h:116: rent address is correct",
      "Program log: ./src/mctpswap/bridgewithfee.c:177: bridge_with_fee acc parsed",
      "Program log: ./src/mctpswap/bridgewithfee.c:185: bridge_with_fee acc validated",
      "Program log: ./src/mctpswap/bridgewithfee.c:201: transfer fee solana",
      "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
      "Program log: Instruction: Transfer",
      "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4645 of 172953 compute units",
      "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
      "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [2]",
      "Program log: Instruction: DepositForBurnWithCaller",
      "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [3]",
      "Program log: Instruction: Burn",
      "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4753 of 141246 compute units",
      "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
      "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd invoke [3]",
      "Program log: Instruction: SendMessageWithCaller",
      "Program 11111111111111111111111111111111 invoke [4]",
      "Program 11111111111111111111111111111111 success",
      "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd consumed 16752 of 130734 compute units",
      "Program return: CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd tgoBAAAAAAA=",
      "Program CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd success",
      "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 invoke [3]",
      "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 3632 of 110014 compute units",
      "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
      "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 consumed 62294 of 166574 compute units",
      "Program return: CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 tgoBAAAAAAA=",
      "Program CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3 success",
      "Program log: ./src/mctpswap/bridgewithfee.c:223: cctp deposit done",
      "Program log: ./src/mctpswap/wormhole.c:209: post bridge_with_fee msg",
      "Program 11111111111111111111111111111111 invoke [2]",
      "Program 11111111111111111111111111111111 success",
      "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth invoke [2]",
      "Program log: Sequence: 16399",
      "Program 11111111111111111111111111111111 invoke [3]",
      "Program 11111111111111111111111111111111 success",
      "Program 11111111111111111111111111111111 invoke [3]",
      "Program 11111111111111111111111111111111 success",
      "Program 11111111111111111111111111111111 invoke [3]",
      "Program 11111111111111111111111111111111 success",
      "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth consumed 27447 of 101211 compute units",
      "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth success",
      "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
      "Program log: Instruction: CloseAccount",
      "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 3015 of 72692 compute units",
      "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
      "Program Awm2zSgzMGTRraAVjRvshqLehy7mJ2Qr3maURDsoDmwi consumed 130206 of 199850 compute units",
      "Program Awm2zSgzMGTRraAVjRvshqLehy7mJ2Qr3maURDsoDmwi success",
    ],
    postBalances: [
      12940241291, 2923200, 1983600, 0, 0, 2512560, 1649520, 1795680, 318749174587, 946560, 1057920,
      226687224, 2039280, 1, 1398960, 0, 1197120, 1405920, 0, 0, 1169280, 1009200, 934087680, 1,
      1141440, 1141440, 1141440,
    ],
    postTokenBalances: [
      {
        accountIndex: 12,
        mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        owner: "7dm9am6Qx7cH64RB99Mzf7ZsLbEfmXM7ihXXCvMiT2X1",
        programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
        uiTokenAmount: {
          amount: "37152381",
          decimals: 6,
          uiAmount: 37.152381,
          uiAmountString: "37.152381",
        },
      },
    ],
    preBalances: [
      12941142471, 0, 0, 2011440, 2039280, 2512560, 1649520, 1795680, 318749174587, 946560, 1057920,
      226687124, 2039280, 1, 1398960, 0, 1197120, 1405920, 0, 0, 1169280, 1009200, 934087680, 1,
      1141440, 1141440, 1141440,
    ],
    preTokenBalances: [
      {
        accountIndex: 4,
        mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        owner: "9emSZvnJyAjfSfqqgv6WUJU573h4NPgjUt8kG3LQJ1Sq",
        programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
        uiTokenAmount: {
          amount: "81553003279",
          decimals: 6,
          uiAmount: 81553.003279,
          uiAmountString: "81553.003279",
        },
      },
      {
        accountIndex: 12,
        mint: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        owner: "7dm9am6Qx7cH64RB99Mzf7ZsLbEfmXM7ihXXCvMiT2X1",
        programId: "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
        uiTokenAmount: {
          amount: "37134919",
          decimals: 6,
          uiAmount: 37.134919,
          uiAmountString: "37.134919",
        },
      },
    ],
    rewards: [],
    status: { Ok: null },
  },
  slot: 286217622,
  transaction: {
    message: {
      header: {
        numReadonlySignedAccounts: 0,
        numReadonlyUnsignedAccounts: 14,
        numRequiredSignatures: 3,
      },
      staticAccountKeys: [
        "7dm9am6Qx7cH64RB99Mzf7ZsLbEfmXM7ihXXCvMiT2X1",
        "AoDggctWU3Gnyov6nbKXtnyrKow1Thcyo78HKrc5RTKj",
        "4MxdcNfR8dZENCtNcDbQWVeXUVzBZ87McsuPAXKsNaVa",
        "9emSZvnJyAjfSfqqgv6WUJU573h4NPgjUt8kG3LQJ1Sq",
        "4k9LMAcFxk2DAgFbBt48uXecjctzLTXeeAM9EZ6jhqMx",
        "BWrwSWjbikT3H7qHAkUEbLmwDQoB4ZDJ4wcSEhSPTZCu",
        "Afgq3BHEfCE7d78D2XE9Bfyu2ieDqvE24xX8KDwreBms",
        "72bvEFk2Usi2uYc1SnaTNhBcQPc6tiJWXr9oKk7rkd4C",
        "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        "Cp2FeYSYeCRAETJsk7twWJ7aLPMCXsdJ3vL9Say6VjzF",
        "2yVjuQwpsvdsrywzsJJVs9Ueh4zayyo5DYJbBNc3DDpn",
        "9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy",
        "67hMagLUiATDtprRkLur73FKmnNVXqwUBoQZkpPjrEs6",
        "ComputeBudget111111111111111111111111111111",
        "Awm2zSgzMGTRraAVjRvshqLehy7mJ2Qr3maURDsoDmwi",
        "X5rMYSBWMqeWULSdDKXXATBjqk9AJF8odHpYJYeYA9H",
        "REzxi9nX3Eqseha5fBiaJhTC6SFJx4qJhP83U4UCrtc",
        "DBD8hAwLDRQkTsu6EqviaYNGKPnsAMmQonxf7AH8ZcFY",
        "CNfZLeeL4RUxwfPnjA3tLiQt4y43jp4V7bMpga673jf9",
        "8tFsB9BjMSiRtRJLmunbDJ3gtvkAr8szgzhAz8ZnBG11",
        "SysvarC1ock11111111111111111111111111111111",
        "SysvarRent111111111111111111111111111111111",
        "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
        "11111111111111111111111111111111",
        "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
        "CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd",
        "CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3",
      ],
      recentBlockhash: "58Dp5gCXYhhG26sKvpvmEUFvTDaq6XpFvgg7jhAzvgnW",
      compiledInstructions: [
        {
          programIdIndex: 13,
          accountKeyIndexes: [],
          data: { type: "Buffer", data: [3, 240, 73, 2, 0, 0, 0, 0, 0] },
        },
        {
          programIdIndex: 14,
          accountKeyIndexes: [
            3, 15, 4, 5, 6, 16, 17, 7, 8, 1, 18, 19, 9, 2, 10, 11, 0, 12, 20, 21, 22, 23, 24, 25,
            26,
          ],
          data: { type: "Buffer", data: [11, 3, 0, 0, 0] },
        },
      ],
      addressTableLookups: [],
      accountKeys: [
        "7dm9am6Qx7cH64RB99Mzf7ZsLbEfmXM7ihXXCvMiT2X1",
        "AoDggctWU3Gnyov6nbKXtnyrKow1Thcyo78HKrc5RTKj",
        "4MxdcNfR8dZENCtNcDbQWVeXUVzBZ87McsuPAXKsNaVa",
        "9emSZvnJyAjfSfqqgv6WUJU573h4NPgjUt8kG3LQJ1Sq",
        "4k9LMAcFxk2DAgFbBt48uXecjctzLTXeeAM9EZ6jhqMx",
        "BWrwSWjbikT3H7qHAkUEbLmwDQoB4ZDJ4wcSEhSPTZCu",
        "Afgq3BHEfCE7d78D2XE9Bfyu2ieDqvE24xX8KDwreBms",
        "72bvEFk2Usi2uYc1SnaTNhBcQPc6tiJWXr9oKk7rkd4C",
        "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
        "Cp2FeYSYeCRAETJsk7twWJ7aLPMCXsdJ3vL9Say6VjzF",
        "2yVjuQwpsvdsrywzsJJVs9Ueh4zayyo5DYJbBNc3DDpn",
        "9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy",
        "67hMagLUiATDtprRkLur73FKmnNVXqwUBoQZkpPjrEs6",
        "ComputeBudget111111111111111111111111111111",
        "Awm2zSgzMGTRraAVjRvshqLehy7mJ2Qr3maURDsoDmwi",
        "X5rMYSBWMqeWULSdDKXXATBjqk9AJF8odHpYJYeYA9H",
        "REzxi9nX3Eqseha5fBiaJhTC6SFJx4qJhP83U4UCrtc",
        "DBD8hAwLDRQkTsu6EqviaYNGKPnsAMmQonxf7AH8ZcFY",
        "CNfZLeeL4RUxwfPnjA3tLiQt4y43jp4V7bMpga673jf9",
        "8tFsB9BjMSiRtRJLmunbDJ3gtvkAr8szgzhAz8ZnBG11",
        "SysvarC1ock11111111111111111111111111111111",
        "SysvarRent111111111111111111111111111111111",
        "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
        "11111111111111111111111111111111",
        "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
        "CCTPmbSD7gX1bxKPAmg77w8oFzNFpaQiQUWD43TKaecd",
        "CCTPiPYPc6AsJuwueEnWgSgucamXDZwBd53dQ11YiKX3",
      ],
    },
    signatures: [
      "5UJQJfjNeN3XBTR6WgDxyHpMRMnPGTKoQJfkGthKrqALqE81oNozfo7dkAHbsAA2BezDdTrmzbzGH6YgBo8DdJ7G",
      "3bkE261yHwMQwuai6d5a4Hi7HhovK91RtEPxzoGCQmSmsRX8V1dn3mNiSLzFgrgtaP4zCRjvPTu8p6jLbmCwMFSB",
      "3jwmnCRBnkZnDHPJMJJ8Hn6YYaPPkgQsJmBQq2J9RpNTWM66mBavXU5H8rgHgtE5mmTZciZtXe1sQ5Rbr2GgUhMR",
    ],
  },
  version: 0,
  chain: "solana",
  chainId: 1,
} as any as solana.Transaction;
