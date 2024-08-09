import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { describe, it, expect, afterEach, afterAll } from "@jest/globals";
import { NearJsonRPCBlockRepository } from "../../../src/infrastructure/repositories";
import { InstrumentedHttpProvider } from "../../../src/infrastructure/rpc/http/InstrumentedHttpProvider";
import nock from "nock";

let repo: NearJsonRPCBlockRepository;
const rpc = "http://localhost";

describe("NearJsonRPCBlockRepository", () => {
  afterAll(() => {
    nock.restore();
  });

  afterEach(() => {
    nock.cleanAll();
  });

  it("should be able to get block height", async () => {
    const expectedHeight = 125151310n;
    givenARepo();
    givenBlockHeightIs();

    const result = await repo.getBlockHeight("final");

    expect(result).toBe(expectedHeight);
  });

  it("should be able to get the transactions", async () => {
    const contract = "contract.portalbridge.near";
    givenARepo();
    givenBlockById();
    givenChunk();
    givenTxStatus();

    const result = await repo.getTransactions(contract, 125151310n, 125151310n);

    expect(result).toBeTruthy();
    expect(result[0].hash).toBe("DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj");
    expect(result[0].receiverId).toBe("contract.portalbridge.near");
    expect(result[0].blockHeight).toBe(125151310n);
    expect(result[0].timestamp).toBe(1722257013);
    expect(result[0].signerId).toBe("tkng.near");
  });
});

const givenARepo = () => {
  repo = new NearJsonRPCBlockRepository({
    get: () => new InstrumentedHttpProvider({ url: rpc, chain: "near" }),
  } as any);
};

const givenBlockHeightIs = () => {
  nock(rpc)
    .post("/", {
      jsonrpc: "2.0",
      method: "block",
      params: { finality: "final" },
      id: "",
    })
    .reply(200, {
      jsonrpc: "2.0",
      result: {
        author: "chorusone.poolv1.near",
        header: {
          height: 125151310,
        },
        chunks: [
          {
            chunk_hash: "4BVNQMQJNYqkeQHq45FyAw6z77mwWusCeMEA9YdU1v5Z",
            prev_block_hash: "CZiUeRsH5byWxMdG5K2e1mrwchrKVDrDXjovCPsYwp7t",
            outcome_root: "GCqf3GRpLfrv5uibXvrRzZAELxKsAHegC1SvyRbsYv34",
            prev_state_root: "4LyMFDWDdFDZb4Wc6GqDiJLxpF1YQxJTxfgHw6dSJttR",
            encoded_merkle_root: "4194YTenCcxGLdm1UqBdNwyVHdP3Td8kC4FppfbiCTqJ",
            encoded_length: 14506,
            height_created: 125151310,
            height_included: 125151310,
            shard_id: 0,
            gas_used: 34557238983670,
            gas_limit: 1000000000000000,
            rent_paid: "0",
            validator_reward: "0",
            balance_burnt: "2548856522443900000000",
            outgoing_receipts_root: "9hRYM7n6wEjneUL8VZw4ejoZJ1VxqPdrfe121ifMJ7BN",
            tx_root: "Ckkid68QZSZSQ52XyN56Cb2N2MUUfzGiVUxfBM1UU1gv",
            validator_proposals: [],
            signature:
              "ed25519:3whHDDY8VSF8T3VRTgBUWpXESW4fdXj8Esexk56nifAFEjaPFqPTjobDUyHWTqEew4xNdMQwGh5RsfiS4NvAZRhm",
          },
          {
            chunk_hash: "7JZ6dxPvx1Bwosm31zCf2EQ74caSPPc8nGZpn9wMhn8T",
            prev_block_hash: "CZiUeRsH5byWxMdG5K2e1mrwchrKVDrDXjovCPsYwp7t",
            outcome_root: "11111111111111111111111111111111",
            prev_state_root: "Fm22m72ANKFfir96krT322ecUtpYN95kmqcpTEeM9jdd",
            encoded_merkle_root: "9zYue7drR1rhfzEEoc4WUXzaYRnRNihvRoGt1BgK7Lkk",
            encoded_length: 8,
            height_created: 125151310,
            height_included: 125151310,
            shard_id: 1,
            gas_used: 0,
            gas_limit: 1000000000000000,
            rent_paid: "0",
            validator_reward: "0",
            balance_burnt: "0",
            outgoing_receipts_root: "AChfy3dXeJjgD2w5zXkUTFb6w8kg3AYGnyyjsvc7hXLv",
            tx_root: "11111111111111111111111111111111",
            validator_proposals: [],
            signature:
              "ed25519:3nVxLhsCxLRc6hvWXApKujvNub4nd1w73LTeimTXZZaYTc1dvzZT7cnXbj4WawEVVTBBbecMbxFvTgQUtyEEt2R8",
          },
        ],
      },
      id: "",
    });
};

const givenBlockById = () => {
  nock(rpc)
    .post("/", {
      jsonrpc: "2.0",
      method: "block",
      params: { block_id: 125151310 },
      id: "",
    })
    .reply(200, {
      jsonrpc: "2.0",
      result: {
        author: "chorusone.poolv1.near",
        header: {
          height: 125151310,
          timestamp: 1722257013320891400,
        },
        chunks: [
          {
            chunk_hash: "4BVNQMQJNYqkeQHq45FyAw6z77mwWusCeMEA9YdU1v5Z",
            prev_block_hash: "CZiUeRsH5byWxMdG5K2e1mrwchrKVDrDXjovCPsYwp7t",
            outcome_root: "GCqf3GRpLfrv5uibXvrRzZAELxKsAHegC1SvyRbsYv34",
            prev_state_root: "4LyMFDWDdFDZb4Wc6GqDiJLxpF1YQxJTxfgHw6dSJttR",
            encoded_merkle_root: "4194YTenCcxGLdm1UqBdNwyVHdP3Td8kC4FppfbiCTqJ",
            encoded_length: 14506,
            height_created: 125151310,
            height_included: 125151310,
            shard_id: 0,
            gas_used: 34557238983670,
            gas_limit: 1000000000000000,
            rent_paid: "0",
            validator_reward: "0",
            balance_burnt: "2548856522443900000000",
            outgoing_receipts_root: "9hRYM7n6wEjneUL8VZw4ejoZJ1VxqPdrfe121ifMJ7BN",
            tx_root: "Ckkid68QZSZSQ52XyN56Cb2N2MUUfzGiVUxfBM1UU1gv",
            validator_proposals: [],
            signature:
              "ed25519:3whHDDY8VSF8T3VRTgBUWpXESW4fdXj8Esexk56nifAFEjaPFqPTjobDUyHWTqEew4xNdMQwGh5RsfiS4NvAZRhm",
          },
        ],
      },
      id: "",
    });
};

const givenChunk = () => {
  nock(rpc)
    .post("/", {
      jsonrpc: "2.0",
      method: "chunk",
      params: { chunk_id: "4BVNQMQJNYqkeQHq45FyAw6z77mwWusCeMEA9YdU1v5Z" },
      id: "",
    })
    .reply(200, {
      jsonrpc: "2.0",
      result: {
        author: "epic.poolv1.near",
        header: {
          chunk_hash: "85b9zR66MidmpdsdVhWmasQrHxpnMMva1AJuok9eaXB5",
          prev_block_hash: "HgnBPLuuFjXzWYErcGeNo7us7nyEezZNRBmvaBuha8Au",
          outcome_root: "7XjnSyzNJFEmPE49zP8boXoydZGWtUKi35fYUYivNgCL",
          prev_state_root: "HdMgACZNKJYFqDzrkUWpbnaj8FiyXPanNJrzSk6oWCUE",
          encoded_merkle_root: "4eNr86p6SZsCBPJ5t1Bx7qEbx3srn7fKtxSsqWoRo53z",
          encoded_length: 29531,
          height_created: 124531378,
          height_included: 124531378,
          shard_id: 0,
          gas_used: 213884082163045,
          gas_limit: 1000000000000000,
          rent_paid: "0",
          validator_reward: "0",
          balance_burnt: "18038248262220600000000",
          outgoing_receipts_root: "3tkzEumsvRfe3MVnkCs8hUQ85xsxWnLt796rwTEKvPY1",
          tx_root: "xKbcSq2qdruEM6UgKtsdKV1LJ2FmPRk3hJod5fk51be",
          validator_proposals: [],
          signature:
            "ed25519:3aUX2kwhwoMDmDytQCNL4zEkoWiinPP8YrRByidYF2EMtnVJNd8o4BHCSpVNRS1fV3UnKaAw8goZTKdw77EjZyJG",
        },
        transactions: [
          {
            signer_id: "tkng.near",
            public_key: "ed25519:3A1RfjLtRyXyhuNEg4SNhvqszQewBWEFG4nZSTUgvY6i",
            nonce: 115669127000059,
            receiver_id: "contract.portalbridge.near",
            actions: [
              {
                FunctionCall: {
                  method_name: "submit_vaa",
                  args: "eyJ2YWEiOiIwMTAwMDAwMDA0MGQwMDEzZTA2ZDFkNzgyMDFiZjQwNDhlMDg4MmJmZTJlNjQ5NjE4ODQzYjI1MDg4NjM3NDVlMTdiYTlhYmM4ZWFlMjAzOTA0MmE1ZTY5NmU3ZmU0ZjRiNjNkOGRjMGI4ZDdhNWZkNDNkNWViMjVhYWI2MmIwZDRlMTFjODZjZjUwMTliMDAwMmUxMjAxNzA5Y2Q4NmZiNDhjYWQyZjlhMTY3OGY5NzZlYmZiYzQ5NGZhZmYzNzZhNDRiMmQwNmY2NGQ5MzI5OWQ2MTQzMjJiYWIzZmJkMzFkZGFjMzZkZjQxYjI2MjM1OGI3MjRmYjg2ZDFjN2U0ODY3ZDk5MDkyYjFhZjgzMzY2MDAwNDE1M2VlMjIwZDAxNzlmNzY5ZmUxZjdjNmQ2ODk5ZWQ5NWRhZDAxNzZjMmFhM2FjZTk2ZjQyOWU5MzU5MjQ2NWY3OGM0MTBhM2FkNWMzZDFkNjZmNzkxZjBhZmE5M2QyZTEwMzcyOWMzOGZmMTg4MzQwNjhlOTU0NWMyZDUxNGE1MDAwNjYxMDUwMjk5ZmJkZjJjYmY5OTM0ZTRhMGE3OTk1OWQzMDhiNzE5MWI2OTNjMWVkODM5MWY4MzZiOTZkNWYxZmQ1ZGQ4ZDQyNmM1NWMwZTZhZDBkMTllNWEzYzVhMGQ4OTQ5MjY4ZTg3NDk3MmVhN2MxNjc4MmE3MDhjZTI5ZDNjMDAwN2Q2ZmZhZjYzODFmMGY2ZWI4NGNlODMyOTIwMGFmNTc5YWEyNTE0YjVmYzQxOTcxMDhkNDU3YmIwOTc4OGNkODk0MmYxNWQ4NWJkZDljOGFmNzBlNjIzZjgxOWM2M2IyNDZlZWQyZGYxZjcyZTI3MWJjYjM4MTNiNWNjNmI5MDFiMDAwOGZjNGNjODhiYTg3NzNhYTQ1Mzc2MjQ1OTY2Y2I3OTY5NDE4MWM5ZTRkYTE2ZTU0ZTU4ZjI4M2VmZTU2YmEyMTM0NjY0NWE1YzY1NzVlNWEyNTFjNDM5Yjk0Yzg3YzEwNzFiNjA5YWE0ZjkyMzU3ZjQ0YTU0NzgyM2FjYWRmNjAxMDAwOWFhYWVkNGY1NDZhMzQ4ZWRmYWIyOGExODMwZGNmZmU3ODIyOTU2ZTU1ZmU3OTFhMDQ5YTdkMDMyMThjY2U0OTMxN2IyZTE3MDk4OGY3ZDRkYWU0YzIyY2MwYjg1OWZmMGFiNTlmZGYzYzMzNTVlNDk3ODAzZjJjYzg0ZjM1YjVmMDEwYTRmZTFiMjcxZGU2ODc2YzIzYjQ5MTFkMTk1YmUwZjY1ZDljNmZiYzBiNjAzZThkNzU2ZDBjZjc5ZDM0YTA2YzI2N2IyODk3ZTY3MGU4M2Y3MzEwN2U3ZjQ4ZmNlMWM3YmFjYjAwNDJmN2M1OGZjNzcxYmMxNzA5ZDk1MmQ1NDY1MDEwZWI2ODM2MGRmNDNlM2U4ZTg1ZGVjODdlYzVmYzc3NDhhZWU2NDAxMzMwMzJhYmQxMTZiMjg0OTdmN2YxZGJlYTM2NjU4Yjg5N2FhZjI3NzEzODNlMWNhZDk4NmIyODcyYWRiYjhmNmQxZTI4ZTAwYjdlMGUyZGQ4ODk2MTZkMWJhMDAwZmYzMWNlYTMzZWRmZTBkNzYzOTI3NWE4Y2RkMmIyODRhYWZkMGRhOWFlMjU3NzBmOGEwMzU3NmQyNzEwNGRlNzc2ZDNmMzk3N2Y4NDk1OTAxZGRmYzQ0MGVkYTc1MDkyNGFiNWE1NWM3ZTA3MWI1N2M5ZjE0NWZhNzEyOWY5OTJlMDExMGYzM2JkNTlmZTc5MmFjOWFhNTlmMDNlNjBjYTU3M2VhM2I2MGQyYzVkNjkzZTY5ODM3ZGJlMTI2MDNhZDU1ZmE1YTIyYWNmMjQyNzc0YWJlMWRmNWI0N2E1NjcyNzBhNWRkZWE3YmQ3NTU5Mjg5OWNjMTA2YzYzOWQzMGJjNmIzMDAxMTg1MzY4NjAwYWRjZTBlMjcyZTU2MGNmMTUwNTAzMDNjMTcyOTg0YTcyZjgwNjYzZDkwNmNmZGVkNWZjNzcwYWMxODdkOTA2Y2RkN2IxM2RiMjE0Njc3YmYyNDRmYjk3N2ZjOTc3MDg5ODZjYzViZTdhMDBiMDE5MzNlZTU5ZDAzMDExMjgxN2U2ODRlYzhiZmQ0MWE5ZGU5NjM1ZjBhMTU3NzIzZDUwZjU5Yjg2ODAxNjkyN2ZjZjU2MDc3MWY0OGE4OTU3MjE2MzRjNDEyOTA0MzZhNWFkOWY1OTg2ZGI0NGI1Y2E5N2VkOTk5NWEzYmIxYzUwODAxNTViODQwNGJhNzRhMDE2NmE3OGUyMzAwMDA0MDIzMDAwMWVjNzM3Mjk5NWQ1Y2M4NzMyMzk3ZmIwYWQzNWMwMTIxZTBlYWE5MGQyNmY4MjhhNTM0Y2FiNTQzOTFiM2E0ZjUwMDAwMDAwMDAwMGUwMTFhMjAwMTAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMjNjMzQ2MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwZmUxM2MzMjA4ZDAyNGQ0MDIwMjEyZDU1OGNmNjAzMDc0ZTk4YmYzMTRkZmRmMTI3Y2MwZTEzOGZiYzcxMmFhZGYwMDBmMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMCJ9",
                  gas: 150000000000000,
                  deposit: "100000000000000000000000",
                },
              },
            ],
            signature:
              "ed25519:G6N2tzhYyacN3J9cudsuYADx6Uw3otKehtYVrPgbEVHPyg6cHqvdvFrUFZMavuowcKYVxyPmMTg7WkEPAeibxYA",
            hash: "DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj",
          },
        ],
        receipts: [
          {
            predecessor_id: "0-relay.hot.tg",
            receiver_id: "trinhvdung-7892.tg",
            receipt_id: "GnQc8bbm4PFKU4tRJiSWcBNnvz8wuxeM3dRjXEYuTcKX",
            receipt: {
              Action: {
                signer_id: "0-relay.hot.tg",
                signer_public_key: "ed25519:4AVPGFAXjrAb75QVRXn4zcpyxULtnMoGqMoLjRNJw8j2",
                gas_price: "122987387",
                output_data_receivers: [],
                input_data_ids: [],
                actions: [
                  {
                    Delegate: {
                      delegate_action: {
                        sender_id: "trinhvdung-7892.tg",
                        receiver_id: "game.hot.tg",
                        actions: [
                          {
                            FunctionCall: {
                              method_name: "l2_claim",
                              args: "eyJzaWduYXR1cmUiOiIxMTk1MTYzYjQ5MzM0MWUyOTQwOWUzMjY5NTczMzBmZDIyYjA5Y2NhNTIyYjI5N2JiN2I3NmI5OGZiMTlhYWFkIiwibWluaW5nX3RpbWUiOiI1NjI1NCIsIm1heF90cyI6IjE3MjIyNTcwMDI4NjkwMzgwODAiLCJib29tX3RpbWUiOjB9",
                              gas: 30000000000000,
                              deposit: "0",
                            },
                          },
                        ],
                        nonce: 115087788007760,
                        max_block_height: 124931370,
                        public_key: "ed25519:J4sNvt7jJxnUzbC7thPnPiQv1e6SJZQR1gVJPXLizv3V",
                      },
                      signature:
                        "ed25519:3LaCHY9yUSVnh9UacViEoFM5yJ6ZNh4kpLpTs2vQxzaS3myH1ASyp7qDpMPEorHpaaznXEFSF6cdr6nZbdRdhNRd",
                    },
                  },
                ],
                is_promise_yield: false,
              },
            },
          },
          {
            predecessor_id: "0-relay.hot.tg",
            receiver_id: "1915432157.tg",
            receipt_id: "Egdk5bRc6BVqvSYh89ChkmxgXJ6xkZUBe5PanTf42hNn",
            receipt: {
              Action: {
                signer_id: "0-relay.hot.tg",
                signer_public_key: "ed25519:9YQ7GNDZJXXWnEkjRe4N9q2im5AEixfD3gZxGnvEPXb6",
                gas_price: "138423388",
                output_data_receivers: [],
                input_data_ids: [],
                actions: [
                  {
                    Delegate: {
                      delegate_action: {
                        sender_id: "1915432157.tg",
                        receiver_id: "game.hot.tg",
                        actions: [
                          {
                            FunctionCall: {
                              method_name: "buy_asset",
                              args: "eyJhc3NldF9pZCI6MX0=",
                              gas: 50000000000000,
                              deposit: "0",
                            },
                          },
                        ],
                        nonce: 121317528000270,
                        max_block_height: 124931371,
                        public_key: "ed25519:HuB1f9iFr78999n75GJKNqQG6NewvdYrLY2L3UkqMi2d",
                      },
                      signature:
                        "ed25519:2zMuqXzUF5LGvsZ1DuvwNor17MYn92UXS7EaxxijZiquFC3ubmQ3Q7y395yhZqtsd6KT6Qg64hjaPP9cA62vMK8Z",
                    },
                  },
                ],
                is_promise_yield: false,
              },
            },
          },
        ],
      },
      id: "",
    });
};

const givenTxStatus = () => {
  nock(rpc)
    .post("/", {
      jsonrpc: "2.0",
      method: "tx",
      params: {
        sender_account_id: "contract.portalbridge.near",
        tx_hash: "DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj",
      },
      id: "",
    })
    .reply(200, {
      jsonrpc: "2.0",
      result: {
        status: {
          SuccessValue: "",
        },
        transaction: {
          signer_id: "tkng.near",
          public_key: "ed25519:3A1RfjLtRyXyhuNEg4SNhvqszQewBWEFG4nZSTUgvY6i",
          nonce: 115669127000059,
          receiver_id: "contract.portalbridge.near",
          actions: [
            {
              FunctionCall: {
                method_name: "submit_vaa",
                args: "eyJ2YWEiOiIwMTAwMDAwMDA0MGQwMDEzZTA2ZDFkNzgyMDFiZjQwNDhlMDg4MmJmZTJlNjQ5NjE4ODQzYjI1MDg4NjM3NDVlMTdiYTlhYmM4ZWFlMjAzOTA0MmE1ZTY5NmU3ZmU0ZjRiNjNkOGRjMGI4ZDdhNWZkNDNkNWViMjVhYWI2MmIwZDRlMTFjODZjZjUwMTliMDAwMmUxMjAxNzA5Y2Q4NmZiNDhjYWQyZjlhMTY3OGY5NzZlYmZiYzQ5NGZhZmYzNzZhNDRiMmQwNmY2NGQ5MzI5OWQ2MTQzMjJiYWIzZmJkMzFkZGFjMzZkZjQxYjI2MjM1OGI3MjRmYjg2ZDFjN2U0ODY3ZDk5MDkyYjFhZjgzMzY2MDAwNDE1M2VlMjIwZDAxNzlmNzY5ZmUxZjdjNmQ2ODk5ZWQ5NWRhZDAxNzZjMmFhM2FjZTk2ZjQyOWU5MzU5MjQ2NWY3OGM0MTBhM2FkNWMzZDFkNjZmNzkxZjBhZmE5M2QyZTEwMzcyOWMzOGZmMTg4MzQwNjhlOTU0NWMyZDUxNGE1MDAwNjYxMDUwMjk5ZmJkZjJjYmY5OTM0ZTRhMGE3OTk1OWQzMDhiNzE5MWI2OTNjMWVkODM5MWY4MzZiOTZkNWYxZmQ1ZGQ4ZDQyNmM1NWMwZTZhZDBkMTllNWEzYzVhMGQ4OTQ5MjY4ZTg3NDk3MmVhN2MxNjc4MmE3MDhjZTI5ZDNjMDAwN2Q2ZmZhZjYzODFmMGY2ZWI4NGNlODMyOTIwMGFmNTc5YWEyNTE0YjVmYzQxOTcxMDhkNDU3YmIwOTc4OGNkODk0MmYxNWQ4NWJkZDljOGFmNzBlNjIzZjgxOWM2M2IyNDZlZWQyZGYxZjcyZTI3MWJjYjM4MTNiNWNjNmI5MDFiMDAwOGZjNGNjODhiYTg3NzNhYTQ1Mzc2MjQ1OTY2Y2I3OTY5NDE4MWM5ZTRkYTE2ZTU0ZTU4ZjI4M2VmZTU2YmEyMTM0NjY0NWE1YzY1NzVlNWEyNTFjNDM5Yjk0Yzg3YzEwNzFiNjA5YWE0ZjkyMzU3ZjQ0YTU0NzgyM2FjYWRmNjAxMDAwOWFhYWVkNGY1NDZhMzQ4ZWRmYWIyOGExODMwZGNmZmU3ODIyOTU2ZTU1ZmU3OTFhMDQ5YTdkMDMyMThjY2U0OTMxN2IyZTE3MDk4OGY3ZDRkYWU0YzIyY2MwYjg1OWZmMGFiNTlmZGYzYzMzNTVlNDk3ODAzZjJjYzg0ZjM1YjVmMDEwYTRmZTFiMjcxZGU2ODc2YzIzYjQ5MTFkMTk1YmUwZjY1ZDljNmZiYzBiNjAzZThkNzU2ZDBjZjc5ZDM0YTA2YzI2N2IyODk3ZTY3MGU4M2Y3MzEwN2U3ZjQ4ZmNlMWM3YmFjYjAwNDJmN2M1OGZjNzcxYmMxNzA5ZDk1MmQ1NDY1MDEwZWI2ODM2MGRmNDNlM2U4ZTg1ZGVjODdlYzVmYzc3NDhhZWU2NDAxMzMwMzJhYmQxMTZiMjg0OTdmN2YxZGJlYTM2NjU4Yjg5N2FhZjI3NzEzODNlMWNhZDk4NmIyODcyYWRiYjhmNmQxZTI4ZTAwYjdlMGUyZGQ4ODk2MTZkMWJhMDAwZmYzMWNlYTMzZWRmZTBkNzYzOTI3NWE4Y2RkMmIyODRhYWZkMGRhOWFlMjU3NzBmOGEwMzU3NmQyNzEwNGRlNzc2ZDNmMzk3N2Y4NDk1OTAxZGRmYzQ0MGVkYTc1MDkyNGFiNWE1NWM3ZTA3MWI1N2M5ZjE0NWZhNzEyOWY5OTJlMDExMGYzM2JkNTlmZTc5MmFjOWFhNTlmMDNlNjBjYTU3M2VhM2I2MGQyYzVkNjkzZTY5ODM3ZGJlMTI2MDNhZDU1ZmE1YTIyYWNmMjQyNzc0YWJlMWRmNWI0N2E1NjcyNzBhNWRkZWE3YmQ3NTU5Mjg5OWNjMTA2YzYzOWQzMGJjNmIzMDAxMTg1MzY4NjAwYWRjZTBlMjcyZTU2MGNmMTUwNTAzMDNjMTcyOTg0YTcyZjgwNjYzZDkwNmNmZGVkNWZjNzcwYWMxODdkOTA2Y2RkN2IxM2RiMjE0Njc3YmYyNDRmYjk3N2ZjOTc3MDg5ODZjYzViZTdhMDBiMDE5MzNlZTU5ZDAzMDExMjgxN2U2ODRlYzhiZmQ0MWE5ZGU5NjM1ZjBhMTU3NzIzZDUwZjU5Yjg2ODAxNjkyN2ZjZjU2MDc3MWY0OGE4OTU3MjE2MzRjNDEyOTA0MzZhNWFkOWY1OTg2ZGI0NGI1Y2E5N2VkOTk5NWEzYmIxYzUwODAxNTViODQwNGJhNzRhMDE2NmE3OGUyMzAwMDA0MDIzMDAwMWVjNzM3Mjk5NWQ1Y2M4NzMyMzk3ZmIwYWQzNWMwMTIxZTBlYWE5MGQyNmY4MjhhNTM0Y2FiNTQzOTFiM2E0ZjUwMDAwMDAwMDAwMGUwMTFhMjAwMTAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMjNjMzQ2MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwZmUxM2MzMjA4ZDAyNGQ0MDIwMjEyZDU1OGNmNjAzMDc0ZTk4YmYzMTRkZmRmMTI3Y2MwZTEzOGZiYzcxMmFhZGYwMDBmMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMCJ9",
                gas: 150000000000000,
                deposit: "100000000000000000000000",
              },
            },
          ],
          signature:
            "ed25519:G6N2tzhYyacN3J9cudsuYADx6Uw3otKehtYVrPgbEVHPyg6cHqvdvFrUFZMavuowcKYVxyPmMTg7WkEPAeibxYA",
          hash: "DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj",
        },
        transaction_outcome: {
          proof: [
            {
              hash: "5PuvDzh9yRemQ6s5LxfP7A6VKXgdaun7ZVhiEZeSnWZi",
              direction: "Left",
            },
            {
              hash: "2bZpa5bNNvHhhh9hY1Taj8ggtj5GwYHJVJbcZiXfGSAW",
              direction: "Right",
            },
            {
              hash: "9BcZqai1xh3d4niFCGcfodPU8AhgpyLob1vCVBy2xKRz",
              direction: "Right",
            },
            {
              hash: "9gXh2ZY1rzTf6YJ62U3GByoKbBKSpy53LLo2Ytpo3GBT",
              direction: "Right",
            },
            {
              hash: "ikWqTGYtq2sBQDSzku2Zj9Zir7pk7fVFmfx1dWkxxC3",
              direction: "Right",
            },
            {
              hash: "6uLVFvU9AtvRmMEZ6jWHLTsd9WndstemabVk4LyHCPW9",
              direction: "Right",
            },
            {
              hash: "4AhGoSU28VZUv7UVwQnDH2uXoJNFAKe7Pwf9AxAv3B36",
              direction: "Right",
            },
            {
              hash: "CWdhpsc7v7KbqPrbnf1bSATgdCUBJ2Rdq3fYEYL58eFq",
              direction: "Right",
            },
          ],
          block_hash: "kyhTyg7j9YpBeUna6AmbNEdEwL78JEjk9UJGvgNKvKg",
          id: "DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj",
          outcome: {
            logs: [],
            receipt_ids: ["3xJpKHibgUfkM87aifLsudT3Cn15JGgfv5L7gNiBFVWW"],
            gas_burnt: 312790736344,
            tokens_burnt: "31279073634400000000",
            executor_id: "tkng.near",
            status: {
              SuccessReceiptId: "3xJpKHibgUfkM87aifLsudT3Cn15JGgfv5L7gNiBFVWW",
            },
            metadata: {
              version: 1,
              gas_profile: null,
            },
          },
        },
        receipts_outcome: [
          {
            proof: [
              {
                hash: "D8VYAYJAMiCWCacv6E3xTUXpwPA2PjyJNzRoYPh53SFG",
                direction: "Left",
              },
              {
                hash: "9itDx2LLBVSfPLewvWHJ6SXNAjX56KQpdzNtCVZeMSbD",
                direction: "Right",
              },
              {
                hash: "6GZsLF5sGqywD6zispB9bxN89wfA94oVKTnivrKk32L1",
                direction: "Left",
              },
              {
                hash: "8AXsPrreyfZVahPex3jxr3ixFx8PtsEwinNasSPNPp9C",
                direction: "Left",
              },
              {
                hash: "X3xKP66DbBGkzoATAi95xbwD45vbckC9dTYuUyzGBbD",
                direction: "Right",
              },
            ],
            block_hash: "9JMURUrqknsbPC4wkzJkDjZxch9Yc4g6siEEkN6j41HK",
            id: "3xJpKHibgUfkM87aifLsudT3Cn15JGgfv5L7gNiBFVWW",
            outcome: {
              logs: [],
              receipt_ids: [
                "HJ1mtppg4wMDaYjfKVoGrXYKGpjhHWkSx1ccKMcLTNwQ",
                "5d48CgVXxc38ck5769qY79bVtTw7UvXA9JFttdy1uaoD",
                "45zwbg6ukpYDW8HCbkDSNbnHUgGTH1SY7zPgaNUvChMa",
                "DsXbfU7sgTenD1dX4RcSFwcUPYckqevWeFv3E5Lmc2Sz",
              ],
              gas_burnt: 4216651008975,
              tokens_burnt: "421665100897500000000",
              executor_id: "contract.portalbridge.near",
              status: {
                SuccessReceiptId: "45zwbg6ukpYDW8HCbkDSNbnHUgGTH1SY7zPgaNUvChMa",
              },
              metadata: {
                version: 3,
                gas_profile: [
                  {
                    cost_category: "ACTION_COST",
                    cost: "FUNCTION_CALL_BASE",
                    gas_used: "600000000000",
                  },
                ],
              },
            },
          },
          {
            proof: [
              {
                hash: "Efy8MQd5K9DhZNZ3dJGMT3EkgHKPfgD6C1d8r8Knq7wB",
                direction: "Right",
              },
            ],
            block_hash: "4esfpz9ni62jhADzzi4mXHD5APAD8ReA2HKnS9dT2bjV",
            id: "HJ1mtppg4wMDaYjfKVoGrXYKGpjhHWkSx1ccKMcLTNwQ",
            outcome: {
              logs: ["wormhole/src/lib.rs#373"],
              receipt_ids: ["DNmxVwQ3XSxyRfoK25NWoTMEQeayy7efcizfmtdyFsU4"],
              gas_burnt: 5934238528279,
              tokens_burnt: "593423852827900000000",
              executor_id: "contract.wormhole_crypto.near",
              status: {
                SuccessValue: "NA==",
              },
              metadata: {
                version: 3,
                gas_profile: [
                  {
                    cost_category: "ACTION_COST",
                    cost: "NEW_DATA_RECEIPT_BYTE",
                    gas_used: "34424022",
                  },
                ],
              },
            },
          },
        ],
        final_execution_status: "FINAL",
      },
      id: "",
    });
};
