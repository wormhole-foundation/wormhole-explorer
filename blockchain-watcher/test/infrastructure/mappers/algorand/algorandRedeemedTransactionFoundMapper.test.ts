import { algorandRedeemedTransactionFoundMapper } from "../../../../src/infrastructure/mappers/algorand/algorandRedeemedTransactionFoundMapper";
import { describe, it, expect } from "@jest/globals";

describe("algorandRedeemedTransactionFoundMapper", () => {
  // 6/0000000000000000000000000e082f06ff657d94310cb8ce8b0d9a04541d8052/220162
  it("should be able to map log to algorandRedeemedTransactionFoundMapper", async () => {
    // When
    const result = algorandRedeemedTransactionFoundMapper(tx, filters);

    if (result) {
      // Then
      expect(result.name).toBe("transfer-redeemed");
      expect(result.chainId).toBe(22);
      expect(result.txHash).toBe("SERG7537SOJADJO5LC2J5SC6DD2VONL76B64YB5PDID2T3FONK5Q");
      expect(result.address).toBe("M7UT7JWIVROIDGMQVJZUBQGBNNIIVOYRPC7JWMGQES4KYJIZHVCRZEGFRQ");
      expect(result.attributes.from).toBe(
        "C3EXCPEEMYTIJ2EYUMEMLBDHIJ7J2KAHGKFHWD4GQX5MP7PYZO7O2C6YZE"
      );
      expect(result.attributes.emitterChain).toBe(6);
      expect(result.attributes.emitterAddress).toBe(
        "0000000000000000000000000E082F06FF657D94310CB8CE8B0D9A04541D8052"
      );
      expect(result.attributes.sequence).toBe(220162);
      expect(result.attributes.status).toBe("completed");
      expect(result.attributes.protocol).toBe("Token Bridge");
    }
  });
});

const tx = {
  payload:
    "AQAAAAQNAMvOqeysTxjYZtgYEOHVj77EPk0eVW0t1/xm5erC78GEJKVo/TJWgLpH+rPrv39ujSllqUNBG+4LCj1vVEO6HUYAASWlSS/cAIzz6mm+ZqpputkD25jFwEKycUGqjj39nP8yQzvR4PEPC75tIIgDLoj0HSzybbH1N7hFykOVE31Tg1cBAliHbjlP/OhEnOspo9WryTrmRmlQNkVGZ+5TRHkwy5WtVVBj5W0BE0M/hnkVAlny37QN99DRTl6cIs0IM4Bpw+QBA3gXwF6aqPF8aVp3vieQNweFLukFFqGXSdxV6KYOmfojc8/P7wLrD+4P9I2Y1XNfi9IoyihI+anU9IFDPpUkmuQABMJXlTMcms2Yit3JBLVjbiImzbtzPIMCRMLXG082cJVtevhCFtxR94pkOF75THhZy3vZ7v2oCQJXp6fssKGP2jgABihy7P7j+ovzB0/i+emkEXZMAoJcrviPbx11A+hAS36ubl+pk1sO1K4d7FM4HjP/f0WosNoCBXqfyD3jCrlZO/kACBZCCH+WTb6bYLPJC+nzzjvHA73QDH4lDkA4KjiJq3Z8MXRqRn6vFR7vjszj9NJCRw9xD6xscpMW0sNXcMFDDqsACYzV/6FXG/XZ7Kd4FYfoQ2ZHLSeRxRF4Xrt5YvsyhxOTHdj0S2b7eb7B41CxDUKzhYvWGf3V4WhYfvRgMFou944BDC8szu1vORgXN/EWBAJwBRmjOZXpHNH7GRIFupubGSQPIDtuom2/DHo2F77vxKC7HXVMVHbV7C3oOgB8w2lXfSABD9wQtyp+4hZAHFNOclrkTNgO/BRFaxXbTwKzKMv1ax3+O21cf3eSlYfdlwTz2DJrts/M/72v18EZriEFGPnFtHUAEOa6k9hnDJZys1ZJ4cSsd77DIFSRTc6GxbaWBTYXfMOjTMSArzdUl5XQKFTRIrDMHRY+sM4dz5dAWc7IM1rYcIoAES8aaahKohCHoR/d//Z0ZT1P4+LcgrF4rABIxki6mDbSaz+86v+derrWYHErb8+s5OpH5qwGbrftiXsDlebiyN8AEg6Rp7M4ooj+qsdqgzt4DUmLgUU2eqBVF1nWuEt0lJgNWAQ2ZPgqp2R+NcsxXEys/lIpSaYgxT3pGLbFWRdpLWgAZnqa+gADj5AABgAAAAAAAAAAAAAAAA4ILwb/ZX2UMQy4zosNmgRUHYBSAAAAAAADXAIBAwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACw7dfAAAAAAAAAAAAAAAAuX7574c0xxkE2AAvi2vGbdnEim4ABgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABK8EJrAAgAAAAAAAAAAAAAAABTZwZsNNSHRYgR77UGp6Kv2tuOi3dvcm1ob2xlRGVwb3NpdAAAAAAAAAAAAAAAAAurcc/Wfp8DdXO0rccmziZuKxssAAAAAAAAAAAAAAAASvBCawAXAAAAAAAAAAAAAAAAC6txz9Z+nwN1c7StxybOJm4rGyw=",
  applicationId: "842126029",
  blockNumber: 40085294,
  timestamp: 1719311110,
  innerTxs: [
    {
      sender: "PG56DVKH6F3RXJASLIR4AXIXDYTFK2DQWIEOLDY22HEX5IPRE47HNFINKY",
    },
  ],
  sender: "C3EXCPEEMYTIJ2EYUMEMLBDHIJ7J2KAHGKFHWD4GQX5MP7PYZO7O2C6YZE",
  hash: "SERG7537SOJADJO5LC2J5SC6DD2VONL76B64YB5PDID2T3FONK5Q",
};
const filters = [
  {
    applicationIds: "842126029",
    applicationAddress: "M7UT7JWIVROIDGMQVJZUBQGBNNIIVOYRPC7JWMGQES4KYJIZHVCRZEGFRQ",
  },
];
