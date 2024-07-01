import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { describe, it, expect, afterEach, afterAll } from "@jest/globals";
import { AlgorandJsonRPCBlockRepository } from "../../../src/infrastructure/repositories";
import { InstrumentedHttpProvider } from "../../../src/infrastructure/rpc/http/InstrumentedHttpProvider";
import nock from "nock";

let repo: AlgorandJsonRPCBlockRepository;
const rpc = "http://localhost";

describe("AlgorandJsonRPCBlockRepository", () => {
  afterAll(() => {
    nock.restore();
  });

  afterEach(() => {
    nock.cleanAll();
  });

  it("should be able to get block height", async () => {
    const expectedHeight = 40087333n;
    givenARepo();
    givenBlockHeightIs();

    const result = await repo.getBlockHeight();

    expect(result).toBe(expectedHeight);
  });

  it("should be able to get the transactions", async () => {
    givenARepo();
    givenTransactions();
    const applicationId = "842125965";
    const result = await repo.getTransactions(applicationId, 40085294n, 40085299n);

    expect(result).toBeTruthy();
    expect(result[0].applicationId).toBe(842125965);
    expect(result[0].blockNumber).toBe(40085294);
    expect(result[0].hash).toBe("Y2PTYVGAJYDALKNN4KVIJ4HBJNY5ZZO3BMYT7AR3KBBYQVXHP3JA");
    expect(result[0].payload).toBe(
      "AMvOqeysTxjYZtgYEOHVj77EPk0eVW0t1/xm5erC78GEJKVo/TJWgLpH+rPrv39ujSllqUNBG+4LCj1vVEO6HUYAASWlSS/cAIzz6mm+ZqpputkD25jFwEKycUGqjj39nP8yQzvR4PEPC75tIIgDLoj0HSzybbH1N7hFykOVE31Tg1cBAliHbjlP/OhEnOspo9WryTrmRmlQNkVGZ+5TRHkwy5WtVVBj5W0BE0M/hnkVAlny37QN99DRTl6cIs0IM4Bpw+QBA3gXwF6aqPF8aVp3vieQNweFLukFFqGXSdxV6KYOmfojc8/P7wLrD+4P9I2Y1XNfi9IoyihI+anU9IFDPpUkmuQABMJXlTMcms2Yit3JBLVjbiImzbtzPIMCRMLXG082cJVtevhCFtxR94pkOF75THhZy3vZ7v2oCQJXp6fssKGP2jgABihy7P7j+ovzB0/i+emkEXZMAoJcrviPbx11A+hAS36ubl+pk1sO1K4d7FM4HjP/f0WosNoCBXqfyD3jCrlZO/kA"
    );
    expect(result[0].sender).toBe("EZATROXX2HISIRZDRGXW4LRQ46Z6IUJYYIHU3PJGP7P5IQDPKVX42N767A");
    expect(result[0].timestamp).toBe(1719311110);
  });
});

const givenARepo = () => {
  repo = new AlgorandJsonRPCBlockRepository(
    { get: () => new InstrumentedHttpProvider({ url: rpc, chain: "algorand" }) } as any,
    { get: () => new InstrumentedHttpProvider({ url: rpc, chain: "algorand" }) } as any
  );
};

const givenBlockHeightIs = () => {
  nock(rpc).post("/v2/status").reply(200, { "last-round": 40087333 });
};

const givenTransactions = () => {
  nock(rpc)
    .post("/v2/transactions?application-id=842125965&min-round=40085294&max-round=40085299")
    .reply(200, {
      "current-round": 40087557,
      "next-token": "LqdjAgAAAAA6AAAA",
      transactions: [
        {
          "application-transaction": {
            accounts: [
              "22DBCQI25XZ52JB5QPBQ72BCMHIYGCEJ3ODIRVOD5MRQTIV6IQUHAT7T5A",
              "XJC32PG73M4VIWAAZQZX6LRHPDAX2DMGVGQJJZSWJUZ5LECEVFMGJENX2M",
            ],
            "application-args": [
              "dmVyaWZ5U2lncw==",
              "AMvOqeysTxjYZtgYEOHVj77EPk0eVW0t1/xm5erC78GEJKVo/TJWgLpH+rPrv39ujSllqUNBG+4LCj1vVEO6HUYAASWlSS/cAIzz6mm+ZqpputkD25jFwEKycUGqjj39nP8yQzvR4PEPC75tIIgDLoj0HSzybbH1N7hFykOVE31Tg1cBAliHbjlP/OhEnOspo9WryTrmRmlQNkVGZ+5TRHkwy5WtVVBj5W0BE0M/hnkVAlny37QN99DRTl6cIs0IM4Bpw+QBA3gXwF6aqPF8aVp3vieQNweFLukFFqGXSdxV6KYOmfojc8/P7wLrD+4P9I2Y1XNfi9IoyihI+anU9IFDPpUkmuQABMJXlTMcms2Yit3JBLVjbiImzbtzPIMCRMLXG082cJVtevhCFtxR94pkOF75THhZy3vZ7v2oCQJXp6fssKGP2jgABihy7P7j+ovzB0/i+emkEXZMAoJcrviPbx11A+hAS36ubl+pk1sO1K4d7FM4HjP/f0WosNoCBXqfyD3jCrlZO/kA",
              "WJO1p2w/c5ZFZIiFvczAbNcKPNP/bLlSWJvehiwl70OSEy+51KQhVxFN6EYBk73zovz4H4agl2X0di/REHoAhrMtegl3kmogUTHYcx05y+uMgrL9gvrtJxHVmvDySZ0W5yb2slTOW000j7dLlY6JZuLsPb1JWKfN",
              "ya4vUYDhuV1IzK9Kb9jilGSjYp5N0/v88OdQsvQmkWw=",
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
          "confirmed-round": 40085294,
          fee: 0,
          "first-valid": 40085291,
          "genesis-hash": "wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8=",
          "genesis-id": "mainnet-v1.0",
          group: "G2RHZhcMt6mtjYBl4tE6CP0Lfw+KNLf3Lj8ciEKby3U=",
          id: "Y2PTYVGAJYDALKNN4KVIJ4HBJNY5ZZO3BMYT7AR3KBBYQVXHP3JA",
          "intra-round-offset": 55,
          "last-valid": 40086291,
          "receiver-rewards": 0,
          "round-time": 1719311110,
          sender: "EZATROXX2HISIRZDRGXW4LRQ46Z6IUJYYIHU3PJGP7P5IQDPKVX42N767A",
          "sender-rewards": 0,
          signature: {
            logicsig: {
              args: [],
              logic:
                "BiAEAQAgFCYBADEgMgMSRDEBIxJEMRCBBhJENhoBNhoDNhoCiAADRCJDNQI1ATUAKDXwKDXxNAAVNQUjNQMjNQQ0AzQFDEEARDQBNAA0A4FBCCJYFzQANAMiCCRYNAA0A4EhCCRYBwA18TXwNAI0BCVYNPA08VACVwwUEkQ0A4FCCDUDNAQlCDUEQv+0Iok=",
            },
          },
          "tx-type": "appl",
        },
        {
          "application-transaction": {
            accounts: [
              "22DBCQI25XZ52JB5QPBQ72BCMHIYGCEJ3ODIRVOD5MRQTIV6IQUHAT7T5A",
              "XJC32PG73M4VIWAAZQZX6LRHPDAX2DMGVGQJJZSWJUZ5LECEVFMGJENX2M",
            ],
            "application-args": [
              "dmVyaWZ5U2lncw==",
              "CBZCCH+WTb6bYLPJC+nzzjvHA73QDH4lDkA4KjiJq3Z8MXRqRn6vFR7vjszj9NJCRw9xD6xscpMW0sNXcMFDDqsACYzV/6FXG/XZ7Kd4FYfoQ2ZHLSeRxRF4Xrt5YvsyhxOTHdj0S2b7eb7B41CxDUKzhYvWGf3V4WhYfvRgMFou944BDC8szu1vORgXN/EWBAJwBRmjOZXpHNH7GRIFupubGSQPIDtuom2/DHo2F77vxKC7HXVMVHbV7C3oOgB8w2lXfSABD9wQtyp+4hZAHFNOclrkTNgO/BRFaxXbTwKzKMv1ax3+O21cf3eSlYfdlwTz2DJrts/M/72v18EZriEFGPnFtHUAEOa6k9hnDJZys1ZJ4cSsd77DIFSRTc6GxbaWBTYXfMOjTMSArzdUl5XQKFTRIrDMHRY+sM4dz5dAWc7IM1rYcIoAES8aaahKohCHoR/d//Z0ZT1P4+LcgrF4rABIxki6mDbSaz+86v+derrWYHErb8+s5OpH5qwGbrftiXsDlebiyN8A",
              "dKO/kTlT1pUmDYi8GqJaTu42PvAACsAHZyezX76i2sKP7lzLD+p2jtLMN6TcA2qNIytI9izdRzFBL0iQgZK25zh8zXaCd8F9qxt6UCfAs88XjiGtLneuBnEVSc+7H5x6nYCW6F4Uh/NVFdAqknU1BKjXVHG59J7b",
              "ya4vUYDhuV1IzK9Kb9jilGSjYp5N0/v88OdQsvQmkWw=",
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
          "confirmed-round": 40085294,
          fee: 0,
          "first-valid": 40085291,
          "genesis-hash": "wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8=",
          "genesis-id": "mainnet-v1.0",
          group: "G2RHZhcMt6mtjYBl4tE6CP0Lfw+KNLf3Lj8ciEKby3U=",
          id: "KXFFIB7C7Z7SYMAFN4BVJEYMUT6DGWFNGYJKK2V2PGMUDJ6I3XHA",
          "intra-round-offset": 56,
          "last-valid": 40086291,
          "receiver-rewards": 0,
          "round-time": 1719311110,
          sender: "EZATROXX2HISIRZDRGXW4LRQ46Z6IUJYYIHU3PJGP7P5IQDPKVX42N767A",
          "sender-rewards": 0,
          signature: {
            logicsig: {
              args: [],
              logic:
                "BiAEAQAgFCYBADEgMgMSRDEBIxJEMRCBBhJENhoBNhoDNhoCiAADRCJDNQI1ATUAKDXwKDXxNAAVNQUjNQMjNQQ0AzQFDEEARDQBNAA0A4FBCCJYFzQANAMiCCRYNAA0A4EhCCRYBwA18TXwNAI0BCVYNPA08VACVwwUEkQ0A4FCCDUDNAQlCDUEQv+0Iok=",
            },
          },
          "tx-type": "appl",
        },
        {
          "application-transaction": {
            accounts: [
              "22DBCQI25XZ52JB5QPBQ72BCMHIYGCEJ3ODIRVOD5MRQTIV6IQUHAT7T5A",
              "XJC32PG73M4VIWAAZQZX6LRHPDAX2DMGVGQJJZSWJUZ5LECEVFMGJENX2M",
            ],
            "application-args": [
              "dmVyaWZ5U2lncw==",
              "Eg6Rp7M4ooj+qsdqgzt4DUmLgUU2eqBVF1nWuEt0lJgNWAQ2ZPgqp2R+NcsxXEys/lIpSaYgxT3pGLbFWRdpLWgA",
              "b768iY9APkdz6V/rFegMmpnINI0=",
              "ya4vUYDhuV1IzK9Kb9jilGSjYp5N0/v88OdQsvQmkWw=",
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
          "confirmed-round": 40085294,
          fee: 0,
          "first-valid": 40085291,
          "genesis-hash": "wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8=",
          "genesis-id": "mainnet-v1.0",
          group: "G2RHZhcMt6mtjYBl4tE6CP0Lfw+KNLf3Lj8ciEKby3U=",
          id: "WKV3GMESISL2DV3CQJZQ365CIZDPRQ4YJ2Q2J775YTYVTFPMWQQQ",
          "intra-round-offset": 57,
          "last-valid": 40086291,
          "receiver-rewards": 0,
          "round-time": 1719311110,
          sender: "EZATROXX2HISIRZDRGXW4LRQ46Z6IUJYYIHU3PJGP7P5IQDPKVX42N767A",
          "sender-rewards": 0,
          signature: {
            logicsig: {
              args: [],
              logic:
                "BiAEAQAgFCYBADEgMgMSRDEBIxJEMRCBBhJENhoBNhoDNhoCiAADRCJDNQI1ATUAKDXwKDXxNAAVNQUjNQMjNQQ0AzQFDEEARDQBNAA0A4FBCCJYFzQANAMiCCRYNAA0A4EhCCRYBwA18TXwNAI0BCVYNPA08VACVwwUEkQ0A4FCCDUDNAQlCDUEQv+0Iok=",
            },
          },
          "tx-type": "appl",
        },
        {
          "application-transaction": {
            accounts: [
              "22DBCQI25XZ52JB5QPBQ72BCMHIYGCEJ3ODIRVOD5MRQTIV6IQUHAT7T5A",
              "XJC32PG73M4VIWAAZQZX6LRHPDAX2DMGVGQJJZSWJUZ5LECEVFMGJENX2M",
            ],
            "application-args": [
              "dmVyaWZ5VkFB",
              "AQAAAAQNAMvOqeysTxjYZtgYEOHVj77EPk0eVW0t1/xm5erC78GEJKVo/TJWgLpH+rPrv39ujSllqUNBG+4LCj1vVEO6HUYAASWlSS/cAIzz6mm+ZqpputkD25jFwEKycUGqjj39nP8yQzvR4PEPC75tIIgDLoj0HSzybbH1N7hFykOVE31Tg1cBAliHbjlP/OhEnOspo9WryTrmRmlQNkVGZ+5TRHkwy5WtVVBj5W0BE0M/hnkVAlny37QN99DRTl6cIs0IM4Bpw+QBA3gXwF6aqPF8aVp3vieQNweFLukFFqGXSdxV6KYOmfojc8/P7wLrD+4P9I2Y1XNfi9IoyihI+anU9IFDPpUkmuQABMJXlTMcms2Yit3JBLVjbiImzbtzPIMCRMLXG082cJVtevhCFtxR94pkOF75THhZy3vZ7v2oCQJXp6fssKGP2jgABihy7P7j+ovzB0/i+emkEXZMAoJcrviPbx11A+hAS36ubl+pk1sO1K4d7FM4HjP/f0WosNoCBXqfyD3jCrlZO/kACBZCCH+WTb6bYLPJC+nzzjvHA73QDH4lDkA4KjiJq3Z8MXRqRn6vFR7vjszj9NJCRw9xD6xscpMW0sNXcMFDDqsACYzV/6FXG/XZ7Kd4FYfoQ2ZHLSeRxRF4Xrt5YvsyhxOTHdj0S2b7eb7B41CxDUKzhYvWGf3V4WhYfvRgMFou944BDC8szu1vORgXN/EWBAJwBRmjOZXpHNH7GRIFupubGSQPIDtuom2/DHo2F77vxKC7HXVMVHbV7C3oOgB8w2lXfSABD9wQtyp+4hZAHFNOclrkTNgO/BRFaxXbTwKzKMv1ax3+O21cf3eSlYfdlwTz2DJrts/M/72v18EZriEFGPnFtHUAEOa6k9hnDJZys1ZJ4cSsd77DIFSRTc6GxbaWBTYXfMOjTMSArzdUl5XQKFTRIrDMHRY+sM4dz5dAWc7IM1rYcIoAES8aaahKohCHoR/d//Z0ZT1P4+LcgrF4rABIxki6mDbSaz+86v+derrWYHErb8+s5OpH5qwGbrftiXsDlebiyN8AEg6Rp7M4ooj+qsdqgzt4DUmLgUU2eqBVF1nWuEt0lJgNWAQ2ZPgqp2R+NcsxXEys/lIpSaYgxT3pGLbFWRdpLWgAZnqa+gADj5AABgAAAAAAAAAAAAAAAA4ILwb/ZX2UMQy4zosNmgRUHYBSAAAAAAADXAIBAwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACw7dfAAAAAAAAAAAAAAAAuX7574c0xxkE2AAvi2vGbdnEim4ABgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABK8EJrAAgAAAAAAAAAAAAAAABTZwZsNNSHRYgR77UGp6Kv2tuOi3dvcm1ob2xlRGVwb3NpdAAAAAAAAAAAAAAAAAurcc/Wfp8DdXO0rccmziZuKxssAAAAAAAAAAAAAAAASvBCawAXAAAAAAAAAAAAAAAAC6txz9Z+nwN1c7StxybOJm4rGyw=",
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
          "confirmed-round": 40085294,
          fee: 4000,
          "first-valid": 40085291,
          "genesis-hash": "wGHE2Pwdvd7S12BL5FaOP20EGYesN73ktiC1qzkkit8=",
          "genesis-id": "mainnet-v1.0",
          group: "G2RHZhcMt6mtjYBl4tE6CP0Lfw+KNLf3Lj8ciEKby3U=",
          id: "3E7KILAIPB4HE4XF5TWUBDHIRXGECZZSLKV3ERPZF64OJXEPRBSA",
          "intra-round-offset": 58,
          "last-valid": 40086291,
          "receiver-rewards": 0,
          "round-time": 1719311110,
          sender: "C3EXCPEEMYTIJ2EYUMEMLBDHIJ7J2KAHGKFHWD4GQX5MP7PYZO7O2C6YZE",
          "sender-rewards": 0,
          signature: {
            sig: "hSaeNt/qY/+QVVDWc44yYcYlt0SejQMLPs/HJp73Io1KzW/0OvKLvWchVu+9YGZdaEc+6xwq8kHMLBrlohpIAA==",
          },
          "tx-type": "appl",
        },
      ],
    });
};
