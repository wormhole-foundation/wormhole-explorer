import { algorandLogMessagePublishedMapper } from "../../../../src/infrastructure/mappers/algorand/algorandLogMessagePublishedMapper";
import { describe, it, expect } from "@jest/globals";

describe("algorandLogMessagePublishedMapper", () => {
  // 8/67e93fa6c8ac5c819990aa7340c0c16b508abb1178be9b30d024b8ac25193d45/10578
  it("should be able to map log to algorandLogMessagePublishedMapper", async () => {
    // When
    const result = algorandLogMessagePublishedMapper(tx, filters);

    if (result) {
      // Then
      expect(result.name).toBe("log-message-published");
      expect(result.chainId).toBe(8);
      expect(result.txHash).toBe("WNXBBFRO2ZAWHPAC5RQOU2U3K7ZV5LWIY6LIAVYSRO2QGUDJOE6A");
      expect(result.address).toBe("BM26KC3NHYQ7BCDWVMP2OM6AWEZZ6ZGYQWKAQFC7XECOUBLP44VOYNBQTA");
      expect(result.attributes.consistencyLevel).toBe(0);
      expect(result.attributes.nonce).toBe(0);
      expect(result.attributes.payload).toBe("AAAAADwK+tc=");
      expect(result.attributes.sender).toBe(
        "67e93fa6c8ac5c819990aa7340c0c16b508abb1178be9b30d024b8ac25193d45"
      );
      expect(result.attributes.sequence).toBe(10578);
    }
  });
});

const tx = {
  payload: "AAAAADwK+tc=",
  method: "c2VuZFRyYW5zZmVy",
  applicationId: "842126029",
  blockNumber: 40095152,
  timestamp: 1719339547,
  innerTxs: [
    {
      applicationId: "842125965",
      payload:
        "AwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAF5FHqAAAAAAAAAAAAAAAAuX7574c0xxkE2AAvi2vGbdnEim4ABgAAAAAAAAAAAAAAAFNnBmw01IdFiBHvtQanoq/a246LAAYLNeULbT4h8Ih2qx+nM8CxM59k2IWUCBRfuQTqBW/nKmNjdHBXaXRoZHJhdwAAAAAAAAAAAAAAAB2P8AKWlWUx/XZJgCHETuO3hZY1ABcAAAAASvBCaw==",
      method: "cHVibGlzaE1lc3NhZ2U=",
      sender: "M7UT7JWIVROIDGMQVJZUBQGBNNIIVOYRPC7JWMGQES4KYJIZHVCRZEGFRQ",
      logs: ["AAAAAAAAKVI="],
    },
  ],
  sender: "BM26KC3NHYQ7BCDWVMP2OM6AWEZZ6ZGYQWKAQFC7XECOUBLP44VOYNBQTA",
  hash: "WNXBBFRO2ZAWHPAC5RQOU2U3K7ZV5LWIY6LIAVYSRO2QGUDJOE6A",
};
const filters = [
  {
    applicationIds: "842125965",
    applicationAddress: "J476J725L4JTOI2YU6DAI4E23LYUECLZR7RCYZ3LK6QFHX4M54ZI53SGXQ",
  },
];
