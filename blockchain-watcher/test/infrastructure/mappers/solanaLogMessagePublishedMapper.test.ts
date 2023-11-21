import { expect, describe, it } from "@jest/globals";
import { solana } from "../../../src/domain/entities";
import { solanaLogMessagePublishedMapper } from "../../../src/infrastructure/mappers/solanaLogMessagePublishedMapper";

describe("solanaLogMessagePublishedMapper", () => {
  it("should map a solana transaction to a log-message-published event", async () => {
    const programId = "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth";
    const tx = {
      blockTime: 1700571926,
      meta: {
        computeUnitsConsumed: 38108,
        err: null,
        fee: 45000,
        innerInstructions: [
          {
            index: 1,
            instructions: [
              { accounts: [0, 1], data: "3Bxs43gDisiYM5ts", programIdIndex: 2, stackHeight: 2 },
              { accounts: [1], data: "9krTDAJ1gisg2T8j", programIdIndex: 2, stackHeight: 2 },
              {
                accounts: [1],
                data: "SYXsBvR59WTsF4KEVN8LCQ1X9MekXCGPPNo3Af36taxCQBED",
                programIdIndex: 2,
                stackHeight: 2,
              },
            ],
          },
        ],
        loadedAddresses: { readonly: [], writable: [] },
        logMessages: [
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth invoke [1]",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [2]",
          "Program 11111111111111111111111111111111 success",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth consumed 38108 of 400000 compute units",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth success",
        ],
        postBalances: [548150063, 1301520, 1, 3647040, 1, 0, 1009200, 1141440],
        postTokenBalances: [],
        preBalances: [549496583, 0, 1, 3647040, 1, 0, 1009200, 1141440],
        preTokenBalances: [],
        rewards: [],
        status: { Ok: null },
      },
      slot: 231368528,
      transaction: {
        message: {
          header: {
            numReadonlySignedAccounts: 0,
            numReadonlyUnsignedAccounts: 6,
            numRequiredSignatures: 2,
          },
          accountKeys: [
            "DPpi7Sv1BgH8rgNLfcF5hQXccjfHzDWCutH6HyVJXo2c",
            "6YWeHdUSFyciT7EwDcT2XnBHQykJv4EiH8j63jS3gKSL",
            "11111111111111111111111111111111",
            "6d3w8mGjJauf6gCAg7WfLezbaPmUHYGuoNutnfYF1RYM",
            "KeccakSecp256k11111111111111111111111111111",
            "Sysvar1nstructions1111111111111111111111111",
            "SysvarRent111111111111111111111111111111111",
            "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
          ],
          recentBlockhash: "BFHcrzxbqH93hinicAiJ4qFUDn3teurRnVSAV3zgyiSt",
          instructions: [
            {
              accounts: [],
              data: "hm62hycs7UDQNrAfDW2YdfVAnH86Ezqr5FRGQQc5VbBR84unD6iT4dneGMkbvseGehcBCs36x7cb1aGui81EvGWnwgAkYRmo4RE4ccABa5hZHwTJY9sEfzTTirfKHCxRpohqAVj2fd13CnJot9hnqRfVTyFS25nFMwnWpTay16BcFcgFBmwWAoSv5EVfcuquRkVh9XMLyTXX1LqwY2ubFrViXSf7vctQroqvVaBf8KF8AiQ1NfkyQqXpiac1Fdz7NdfmdNxabaurJi8Q4Li2s6x6Yhd5qjpRWFsvpThs3HXjm3qRfN2biY5HTy5abmdyz8NR3jxybgzsehUTtwTDRQTdvdszUDWKYmjViqLuEtesTg68bkGk2Sw8vWzFpw8wHueEG4RW3XVPvmyLYsuKE2UXiyntUczc14FHUuTRuMpFDcoajhHk9kPFSqFYMEBiK5MQj9xK1dEn32J5WMQwbvUfxgBpLW6pCymWTPCecHDsCWYxG3DpSmrsbBCuzxdVN1ZFWF2WbG53gVCYNCksaHZmzCkXjo985bgZ9yByaYhtkNc5mWMffYdnynUWTwcB7iW8qN9YbVK44QMqgsK8s6V2GHw6S2v78pUw2a6HtVFyoNThN77yyn7YhHEKroXKWiNXgMnGxLjNTUB8DHUYFg7Ak9wB2C6whkPowkDreUovSXoncC6ovrgT97g7tWmj8PcwAD7iPgDc2AFYFqV6fbj5gJwDxT38sheRYUHoCGREmpmDyJveCsgo74HpimwyZhnauK4pZi4M4aWNLTvdrEdHZ4AVvnzbst6WBkg62nNAiz2dkqUgVLBh1oAnk5rFTuqowBuoMa3gVVKo758wT9qksVgYbZDJKrcqr89g6huQ7VNDjwxaueYbtjmRKxvwb9v6mdbDtGApmtSYNroWznQ18KF1SW422VHdAbfY1NXbd9PM5u2c4mv7MYdFxKsGiN",
              programIdIndex: 4,
              stackHeight: null,
            },
            {
              accounts: [0, 3, 1, 5, 6, 2],
              data: "6f4i6QEQuN9Xu1k2omxzJbdsSVL",
              programIdIndex: 7,
              stackHeight: null,
            },
          ],
          indexToProgramIds: {},
          compiledInstructions: [
            {
              programIdIndex: 4,
              accountKeyIndexes: [],
              data: {
                type: "Buffer",
                data: [
                  7, 78, 0, 0, 143, 0, 0, 161, 2, 32, 0, 0, 163, 0, 0, 228, 0, 0, 161, 2, 32, 0, 0,
                  248, 0, 0, 57, 1, 0, 161, 2, 32, 0, 0, 77, 1, 0, 142, 1, 0, 161, 2, 32, 0, 0, 162,
                  1, 0, 227, 1, 0, 161, 2, 32, 0, 0, 247, 1, 0, 56, 2, 0, 161, 2, 32, 0, 0, 76, 2,
                  0, 141, 2, 0, 161, 2, 32, 0, 0, 59, 51, 204, 135, 248, 215, 99, 74, 107, 66, 65,
                  74, 160, 39, 31, 10, 158, 178, 19, 247, 139, 113, 78, 152, 216, 104, 158, 173,
                  117, 242, 60, 47, 111, 11, 39, 15, 162, 101, 151, 167, 65, 70, 40, 10, 189, 243,
                  205, 153, 220, 47, 188, 222, 232, 11, 218, 184, 255, 231, 127, 221, 37, 61, 206,
                  230, 0, 88, 204, 58, 229, 192, 151, 178, 19, 206, 60, 129, 151, 158, 27, 159, 149,
                  112, 116, 106, 165, 3, 210, 147, 221, 8, 171, 179, 168, 188, 218, 112, 226, 92,
                  49, 49, 27, 124, 46, 134, 151, 125, 137, 207, 247, 242, 138, 32, 193, 210, 41,
                  171, 208, 67, 188, 141, 100, 185, 141, 64, 53, 165, 10, 209, 137, 107, 18, 181,
                  66, 93, 26, 81, 67, 162, 46, 193, 145, 96, 135, 120, 216, 22, 191, 65, 123, 1,
                  255, 108, 185, 82, 88, 155, 222, 134, 44, 37, 239, 67, 146, 19, 47, 185, 212, 164,
                  33, 87, 156, 94, 21, 252, 196, 69, 191, 21, 115, 155, 193, 198, 202, 177, 243, 53,
                  80, 54, 244, 147, 184, 30, 84, 210, 14, 117, 87, 25, 29, 59, 26, 213, 61, 160,
                  214, 67, 51, 168, 83, 98, 116, 109, 214, 78, 101, 113, 149, 63, 88, 8, 254, 217,
                  15, 201, 154, 197, 127, 33, 176, 22, 72, 196, 210, 245, 0, 17, 77, 232, 70, 1,
                  147, 189, 243, 162, 252, 248, 31, 134, 160, 151, 101, 244, 118, 47, 209, 4, 63,
                  253, 158, 86, 175, 163, 112, 215, 156, 208, 139, 218, 88, 193, 237, 42, 85, 42,
                  173, 196, 121, 102, 233, 227, 133, 147, 146, 60, 101, 80, 239, 64, 241, 154, 63,
                  198, 250, 78, 190, 167, 90, 86, 181, 207, 125, 34, 79, 14, 127, 156, 245, 84, 246,
                  183, 164, 176, 57, 5, 146, 83, 195, 128, 93, 0, 16, 122, 0, 134, 179, 45, 122, 9,
                  119, 146, 106, 32, 81, 49, 216, 115, 29, 57, 203, 235, 153, 210, 200, 156, 182,
                  223, 166, 149, 149, 170, 47, 151, 192, 41, 129, 137, 93, 215, 178, 181, 225, 104,
                  186, 127, 119, 236, 109, 99, 178, 176, 72, 114, 96, 227, 217, 155, 138, 221, 154,
                  70, 143, 112, 32, 185, 30, 197, 4, 158, 50, 149, 239, 237, 42, 169, 127, 253, 164,
                  207, 76, 40, 145, 125, 108, 252, 0, 140, 130, 178, 253, 130, 250, 237, 39, 17,
                  213, 154, 240, 242, 73, 157, 22, 231, 38, 246, 178, 13, 25, 229, 42, 53, 107, 116,
                  76, 12, 67, 37, 29, 182, 248, 226, 60, 3, 119, 195, 199, 147, 102, 74, 41, 57,
                  110, 123, 60, 184, 30, 119, 252, 104, 157, 39, 179, 190, 11, 134, 145, 209, 198,
                  123, 98, 62, 94, 159, 195, 14, 25, 115, 56, 229, 58, 140, 237, 185, 145, 202, 243,
                  155, 117, 248, 181, 1, 84, 206, 91, 77, 52, 143, 183, 75, 149, 142, 137, 102, 226,
                  236, 61, 189, 73, 88, 167, 205, 210, 49, 221, 81, 99, 210, 175, 81, 22, 34, 101,
                  20, 219, 184, 39, 130, 193, 30, 173, 193, 14, 220, 163, 55, 219, 125, 17, 160,
                  206, 143, 46, 224, 124, 192, 116, 60, 223, 75, 34, 151, 183, 117, 193, 233, 202,
                  147, 252, 71, 146, 15, 70, 232, 4, 137, 128, 7, 186, 220, 206, 235, 238, 107, 212,
                  98, 1, 116, 163, 191, 145, 57, 83, 214, 149, 38, 13, 136, 188, 26, 162, 90, 78,
                  238, 54, 62, 240, 145, 62, 159, 160, 23, 234, 98, 198, 39, 201, 106, 33, 183, 112,
                  38, 215, 224, 190, 86, 146, 31, 102, 155, 202, 73, 138, 117, 235, 0, 141, 115, 75,
                ],
              },
            },
            {
              programIdIndex: 7,
              accountKeyIndexes: [0, 3, 1, 5, 6, 2],
              data: {
                type: "Buffer",
                data: [
                  7, 0, 1, 2, 3, 4, 255, 5, 255, 6, 255, 255, 255, 255, 255, 255, 255, 255, 255,
                  255,
                ],
              },
            },
          ],
        },
      },
      version: "legacy",
    } as any as solana.Transaction;

    const events = await solanaLogMessagePublishedMapper(tx, { programId });

    expect(events).toHaveLength(1);
    expect(events[0].name).toBe("log-message-published");
    expect(events[0].address).toBe(programId);
    expect(events[0].chainId).toBe(1);
    expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
    expect(events[0].blockHeight).toBe(tx.slot.toString());
    expect(events[0].blockTime).toBe(tx.blockTime);
  });
});
