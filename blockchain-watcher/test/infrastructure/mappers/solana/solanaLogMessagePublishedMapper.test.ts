import { expect, describe, it, jest } from "@jest/globals";
import { solana } from "../../../../src/domain/entities";
import { solanaLogMessagePublishedMapper } from "../../../../src/infrastructure/mappers/solana/solanaLogMessagePublishedMapper";
import { getPostedMessage } from "@certusone/wormhole-sdk/lib/cjs/solana/wormhole";

jest.mock("@certusone/wormhole-sdk/lib/cjs/solana/wormhole");

describe("solanaLogMessagePublishedMapper", () => {
  it("should map a solana transaction to a log-message-published event", async () => {
    const mockGetPostedMessage = getPostedMessage as jest.MockedFunction<typeof getPostedMessage>;
    mockGetPostedMessage.mockResolvedValueOnce({
      message: {
        emitterChain: 1,
        sequence: 1n,
        emitterAddress: Buffer.from("7dm9am6Qx7cH64RB99Mzf7ZsLbEfmXM7ihXXCvMiT2X1", "hex"),
        submissionTime: 1700571923,
        nonce: 1,
        consistencyLevel: 1,
        payload: Buffer.from("41QVZTrdrRxb", "base64"),
      } as any,
    });

    const programId = "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth";
    const tx = {
      blockTime: 1700571923,
      meta: {
        computeUnitsConsumed: 114965,
        err: null,
        fee: 10000,
        innerInstructions: [
          {
            index: 0,
            instructions: [
              { accounts: [8, 12, 11], data: "41QVZTrdrRxb", programIdIndex: 19, stackHeight: 2 },
              {
                accounts: [0, 14, 8, 11, 5, 16, 12, 2, 1, 15, 6, 3, 17, 0, 18, 10, 20, 19],
                data: "A3ktuLMbmSRxRSrkzhPwxk77nFqnyxKr5uiQ83okUWL4V6Lv6FFoqn5uSFxSzyGCvJb9TdY29txdrxJMbFCH7jBsU5Pvq7WAAYHoHe43ZmFd7fARFJsQNWknoTNRAbmmUHezG4rVo5H",
                programIdIndex: 21,
                stackHeight: 2,
              },
              { accounts: [8, 5, 12], data: "6xgaMiT8w5vF", programIdIndex: 19, stackHeight: 3 },
              { accounts: [0, 3], data: "3Bxs4HanWsHUZCbH", programIdIndex: 10, stackHeight: 3 },
              {
                accounts: [2, 1, 15, 6, 0, 3, 17, 10, 18],
                data: "5WRA5BK5zWbhP4hcFCuyykwdemimeRAzp1cmpuNA8xAiatnD1R65GWoYK5cZjrKJYacibisYizNq2Yu1NuBjJBqqB4bt6FjtMu7dXbEnm4o2KJrQGBC5rFiRxMiJsqED4kYz3zzuPvWuiyjC5gjhKp1CCNYEZxcCByC42UrkMmKwfT1HacaRNcesnyrchKE1vHrvoFPmPfAb4rzHNnqnq4R63wXd9e4YjyNBkEz935Y1a4L2zEiep7opJVFGMD55NTYmmVi",
                programIdIndex: 20,
                stackHeight: 3,
              },
              { accounts: [0, 1], data: "3Bxs46EF5kSf8Vhh", programIdIndex: 10, stackHeight: 4 },
              { accounts: [1], data: "9krTD476zgsrbsPV", programIdIndex: 10, stackHeight: 4 },
              {
                accounts: [1],
                data: "SYXsBvR59WTsF4KEVN8LCQ1X9MekXCGPPNo3Af36taxCQBED",
                programIdIndex: 10,
                stackHeight: 4,
              },
              { accounts: [9, 4, 11], data: "3j3xf5aqFS8P", programIdIndex: 19, stackHeight: 2 },
            ],
          },
        ],
        logMessages: [
          "Program 8LPjGDbxhW4G2Q8S6FvdvUdfGWssgtqmvsc63bwNFA7E invoke [1]",
          "Program log: mayan-swap v0.2 (build  at 1693428062)",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: Approve",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 2904 of 176787 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb invoke [2]",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [3]",
          "Program log: Instruction: Burn",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4790 of 135281 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program 11111111111111111111111111111111 invoke [3]",
          "Program 11111111111111111111111111111111 success",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth invoke [3]",
          "Program log: Sequence: 328338",
          "Program 11111111111111111111111111111111 invoke [4]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [4]",
          "Program 11111111111111111111111111111111 success",
          "Program 11111111111111111111111111111111 invoke [4]",
          "Program 11111111111111111111111111111111 success",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth consumed 27438 of 119637 compute units",
          "Program worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth success",
          "Program wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb consumed 81223 of 172075 compute units",
          "Program wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb success",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA invoke [2]",
          "Program log: Instruction: Transfer",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA consumed 4736 of 89780 compute units",
          "Program TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA success",
          "Program 8LPjGDbxhW4G2Q8S6FvdvUdfGWssgtqmvsc63bwNFA7E consumed 114965 of 200000 compute units",
          "Program 8LPjGDbxhW4G2Q8S6FvdvUdfGWssgtqmvsc63bwNFA7E success",
        ],
        status: { Ok: null },
      },
      slot: 231368523,
      transaction: {
        message: {
          header: {
            numReadonlySignedAccounts: 0,
            numReadonlyUnsignedAccounts: 12,
            numRequiredSignatures: 2,
          },
          accountKeys: [
            "7dm9am6Qx7cH64RB99Mzf7ZsLbEfmXM7ihXXCvMiT2X1",
            "DBQF5sQK9VWh7oSGcygNEBnHkkBqnEjaswJQ1fzkQmKm",
            "2yVjuQwpsvdsrywzsJJVs9Ueh4zayyo5DYJbBNc3DDpn",
            "9bFNrXNb2WTx8fMHXCheaZqkLZ3YCCaiqTftHxeintHy",
            "BKJnouHX2xuqZLpDjXYbdjdjdSZ8QAtjdwyH43Eiu6uD",
            "CSD6JQMvLi46psjHdpfFdr826mF336pEVMJgjwcoS1m4",
            "GF2ghkjwsR9CHkGk1RvuZrApPZGBZynxMm817VNi51Nf",
            "GoCsLpg3sWWAmQDBUpE9pPiyqsuXFzZNJnCXvCXiD5P2",
            "GYTw6AtTSQjWJsBqdTsMxtKd1797zsfSpDawG64Ryo1y",
            "HjFhpdX7VyAW8Kq4fhughu6MHFRVwdCwVjZLVrVaag3k",
            "11111111111111111111111111111111",
            "5yZiE74sGLCT4uRoyeqz4iTYiUwX5uykiPRggCVih9PN",
            "7oPa2PHQdZmjSPqvpZN7MQxnC7Dcf3uL4oLqknGLk2S3",
            "8LPjGDbxhW4G2Q8S6FvdvUdfGWssgtqmvsc63bwNFA7E",
            "DapiQYH3BGonhN8cngWcXQ6SrqSm3cwysoznoHr6Sbsx",
            "Gv1KWf8DT1jKv5pKBmGaTmVszqa56Xn8YGx2Pg7i7qAk",
            "HZ3rxK31XBHQGyBG95RAix1JeFyF2YMSaYjm2th9LxsB",
            "SysvarC1ock11111111111111111111111111111111",
            "SysvarRent111111111111111111111111111111111",
            "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA",
            "worm2ZoG2kUd4vFXhvjh93UUH596ayRfgQ2MgjNMTth",
            "wormDTUJ6AWPNvk59vGQbDvGJmqbDTdgWgAqcLBCgUb",
          ],
          recentBlockhash: "D8ZWuRAhdPg2cvdy13W3nZFtPtYpkhq8S88yMMrALTcU",
          instructions: [
            {
              accounts: [0, 7, 11, 4, 9, 14, 12, 15, 2, 6, 3, 5, 16, 8, 1, 18, 17, 10, 20, 19, 21],
              data: "4Qxwod2ZmVmMFkX2cD8Uw",
              programIdIndex: 13,
              stackHeight: null,
            },
          ],
          indexToProgramIds: {},
          compiledInstructions: [
            {
              programIdIndex: 13,
              accountKeyIndexes: [
                0, 7, 11, 4, 9, 14, 12, 15, 2, 6, 3, 5, 16, 8, 1, 18, 17, 10, 20, 19, 21,
              ],
              data: {
                type: "Buffer",
                data: [121, 255, 255, 243, 31, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0],
              },
            },
          ],
        },
        signatures: [
          "1NP3vjTMDQnu94JrBKtQqyZLrg9Ep6bgmHbsGQJjUmdLAkfnrkTgRWPM5nLBLYhbGPNJQMv3gMhtwWrW6QHk6iv",
          "Sxz3ipEz8SU5tkEGTp97q9weZfgaBVz6bZNUBDNDGLhRUnrVAYDJ2vFHognsGyZ9YsA1YrcMxxYpmWjzNfyX4E2",
        ],
      },
      version: "legacy",
    } as any as solana.Transaction;

    const events = await solanaLogMessagePublishedMapper(tx, { programId });

    expect(events).toHaveLength(1);
    expect(events[0].name).toBe("log-message-published");
    expect(events[0].address).toBe(programId);
    expect(events[0].chainId).toBe(1);
    expect(events[0].txHash).toBe(tx.transaction.signatures[0]);
    expect(events[0].blockHeight).toBe(BigInt(tx.slot));
    expect(events[0].blockTime).toBe(tx.blockTime);
  });
});
