import { algorandLogMessagePublishedMapper } from "../../../../src/infrastructure/mappers/algorand/algorandLogMessagePublishedMapper";
import { describe, it, expect } from "@jest/globals";

describe("algorandLogMessagePublishedMapper", () => {
  // 8/67e93fa6c8ac5c819990aa7340c0c16b508abb1178be9b30d024b8ac25193d45/10576
  it("should be able to map log to algorandLogMessagePublishedMapper", async () => {
    // When
    const result = algorandLogMessagePublishedMapper(tx);

    if (result) {
      // Then
      expect(result.name).toBe("log-message-published");
      expect(result.chainId).toBe(8);
      expect(result.txHash).toBe("SQA7S37MCLGHQRMFZHRNUNUFJ6PJKRZN5RO52NMEWJU5B365SINQ");
      expect(result.address).toBe("MG3DIJNS3JTVKUAQGFV5BQTDAK26OUM3SRXSLIFWVUS67V54VPKDUJQTOQ");
      expect(result.attributes.consistencyLevel).toBe(0);
      expect(result.attributes.nonce).toBe(0);
      expect(result.attributes.payload).toBe("AAAAADcXNho=");
      expect(result.attributes.sender).toBe(
        "67e93fa6c8ac5c819990aa7340c0c16b508abb1178be9b30d024b8ac25193d45"
      );
      expect(result.attributes.sequence).toBe(10576);
    }
  });
});

const tx = {
  payload: "AAAAADcXNho=",
  applicationId: "842126029",
  blockNumber: 40085318,
  timestamp: 1719311180,
  innerTxs: [
    {
      "application-transaction": {
        accounts: ["KGLLCLRLXRIANIGQCGBNSYITYAB2NEUQGQ6A5LN6SXTPDKWUEFDUUSEGDM"],
        "application-args": [
          "cHVibGlzaE1lc3NhZ2U=",
          "AQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMleRn4AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAADcXNhoACAAAAAAAAAAAAAAAAFvAqglWMsdtoL6YONlg2ppm/KMOAAUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
          "AAAAAAAAAAA=",
        ],
        "application-id": 842125965,
        "foreign-apps": [],
        "foreign-assets": [],
        "global-state-schema": { "num-byte-slice": 0, "num-uint": 0 },
        "local-state-schema": { "num-byte-slice": 0, "num-uint": 0 },
        "on-completion": "noop",
      },
      "close-rewards": 0,
      "closing-amount": 0,
      "confirmed-round": 40085318,
      fee: 0,
      "first-valid": 40085314,
      "intra-round-offset": 11,
      "last-valid": 40086314,
      "local-state-delta": [
        {
          address: "KGLLCLRLXRIANIGQCGBNSYITYAB2NEUQGQ6A5LN6SXTPDKWUEFDUUSEGDM",
          delta: [
            {
              key: "AA==",
              value: {
                action: 1,
                bytes:
                  "AAAAAAAAKVAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
                uint: 0,
              },
            },
          ],
        },
      ],
      logs: ["AAAAAAAAKVA="],
      note: "cHVibGlzaE1lc3NhZ2U=",
      "receiver-rewards": 0,
      "round-time": 1719311180,
      sender: "M7UT7JWIVROIDGMQVJZUBQGBNNIIVOYRPC7JWMGQES4KYJIZHVCRZEGFRQ",
      "sender-rewards": 0,
      "tx-type": "appl",
    },
  ],
  sender: "MG3DIJNS3JTVKUAQGFV5BQTDAK26OUM3SRXSLIFWVUS67V54VPKDUJQTOQ",
  hash: "SQA7S37MCLGHQRMFZHRNUNUFJ6PJKRZN5RO52NMEWJU5B365SINQ",
};
