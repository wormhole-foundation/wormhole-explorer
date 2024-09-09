import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { thenWaitForAssertion } from "../../../waitAssertion";
import { AlgorandTransaction } from "../../../../src/domain/entities/algorand";
import {
  PollAlgorandMetadata,
  PollAlgorandConfig,
  PollAlgorand,
} from "../../../../src/domain/actions";
import {
  AlgorandRepository,
  MetadataRepository,
  StatRepository,
} from "../../../../src/domain/repositories";

let getBlockHeightSpy: jest.SpiedFunction<AlgorandRepository["getBlockHeight"]>;
let getTransactionsSpy: jest.SpiedFunction<AlgorandRepository["getTransactions"]>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollAlgorandMetadata>["save"]>;

let handlerSpy: jest.SpiedFunction<(txs: AlgorandTransaction[]) => Promise<void>>;

let metadataRepo: MetadataRepository<PollAlgorandMetadata>;
let algorandRepo: AlgorandRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (txs: AlgorandTransaction[]) => Promise.resolve(),
  failing: (txs: AlgorandTransaction[]) => Promise.reject(),
};

let pollAlgorand: PollAlgorand;

let cfg = new PollAlgorandConfig({
  chain: "algorand",
  applicationIds: ["842125965"],
  chainId: 8,
  environment: "testnet",
});

describe("GetAlgorandTransactions", () => {
  afterEach(async () => {
    await pollAlgorand.stop();
  });

  it("should be use from and batch size cfg, and process tx because is a wormhole redeem", async () => {
    // Given
    const txs = [
      {
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
      },
    ];

    givenAlgorandRepository(402222n, txs);
    givenMetadataRepository();
    givenStatsRepository();
    givenPollAlgorandTxs();

    // Whem
    await whenPollAlgorandStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getBlockHeightSpy).toHaveReturnedTimes(1),
      () => expect(getTransactionsSpy).toBeCalledWith("842125965", 402222n, 402222n)
    );
  });
});

const givenAlgorandRepository = (height?: bigint, txs: any = []) => {
  algorandRepo = {
    getBlockHeight: () => Promise.resolve(height),
    getTransactions: () => Promise.resolve(txs),
  };

  getBlockHeightSpy = jest.spyOn(algorandRepo, "getBlockHeight");
  getTransactionsSpy = jest.spyOn(algorandRepo, "getTransactions");
  handlerSpy = jest.spyOn(handlers, "working");
};

const givenMetadataRepository = (data?: PollAlgorandMetadata) => {
  metadataRepo = {
    get: () => Promise.resolve(data),
    save: () => Promise.resolve(),
  };
  metadataSaveSpy = jest.spyOn(metadataRepo, "save");
};

const givenStatsRepository = () => {
  statsRepo = {
    count: () => {},
    measure: () => {},
    report: () => Promise.resolve(""),
  };
};

const givenPollAlgorandTxs = (from?: bigint) => {
  cfg.setFromBlock(from);
  pollAlgorand = new PollAlgorand(algorandRepo, metadataRepo, statsRepo, cfg);
};

const whenPollAlgorandStarts = async () => {
  pollAlgorand.run([handlers.working]);
};
