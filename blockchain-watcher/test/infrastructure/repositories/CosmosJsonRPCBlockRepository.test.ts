import { mockRpcPool } from "../../mocks/mockRpcPool";
mockRpcPool();

import { CosmosJsonRPCBlockRepository } from "../../../src/infrastructure/repositories";
import { InstrumentedHttpProvider } from "../../../src/infrastructure/rpc/http/InstrumentedHttpProvider";
import { describe, it, expect } from "@jest/globals";
import axios from "axios";
import nock from "nock";

axios.defaults.adapter = "http"; // needed by nock
const rpc = "http://localhost";

let repo: CosmosJsonRPCBlockRepository;

// Mock pools
const cosmosPools: Map<number, any> = new Map([
  [3, () => new InstrumentedHttpProvider({ url: rpc, chain: "terra" })],
  [18, { get: () => new InstrumentedHttpProvider({ url: rpc, chain: "terra2" }) } as any],
  [32, { get: () => new InstrumentedHttpProvider({ url: rpc, chain: "sei" }) } as any],
]);

describe("CosmosJsonRPCBlockRepository", () => {
  it("should be able to return cosmos transactions", async () => {
    givenARepo(cosmosPools);
    givenTransactions();

    const result = await repo.getTransactions(
      18,
      {
        addresses: ["sei1smzlm9t79kur392nu9egl8p8je9j92q4gzguewj56a05kyxxra0qy0nuf3"],
      },
      20,
      "terra2"
    );

    expect(result).toBeTruthy();
    expect(2).toBe(result.length);
    expect("1C72CB1D4925D7BA7FB5484555C817FA58052F03FBDB1F192835E2158EDE67A4").toBe(result[0].hash);
    expect("terra2").toBe(result[0].chain);
    expect(18).toBe(result[0].chainId);
    expect(result[0].events).toBeTruthy();
    expect(result[0].data).toBeTruthy();
    expect(10252197n).toBe(result[0].height);
  });

  it("should be able to get block timestamp", async () => {
    givenARepo(cosmosPools);
    givenBlockHeightIs();

    const result = await repo.getBlockTimestamp(80542798n, 32, "sei");

    expect(1718722267).toBe(result); // '2024-06-18T14:51:07.000Z'
  });
});

const givenARepo = (cosmosPools: any = undefined) => {
  repo = new CosmosJsonRPCBlockRepository(cosmosPools);
};

const givenTransactions = () => {
  nock(rpc)
    .post(
      "/tx_search?query=%22wasm._contract_address=%27sei1smzlm9t79kur392nu9egl8p8je9j92q4gzguewj56a05kyxxra0qy0nuf3%27%22&page=1&per_page=20"
    )
    .reply(200, {
      jsonrpc: "2.0",
      id: -1,
      result: {
        txs: [
          {
            hash: "1C72CB1D4925D7BA7FB5484555C817FA58052F03FBDB1F192835E2158EDE67A4",
            height: "10252197",
            index: 1,
            tx_result: {
              code: 0,
              data: "Ei4KLC9jb3Ntd2FzbS53YXNtLnYxLk1zZ0V4ZWN1dGVDb250cmFjdFJlc3BvbnNl",
              log: '[{"msg_index":0,"events":[{"type":"message","attributes":[{"key":"action","value":"/cosmwasm.wasm.v1.MsgExecuteContract"},{"key":"sender","value":"terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx"},{"key":"module","value":"wasm"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx"},{"key":"amount","value":"1000uluna"}]},{"type":"coin_received","attributes":[{"key":"receiver","value":"terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9"},{"key":"amount","value":"1000uluna"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9"},{"key":"sender","value":"terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx"},{"key":"amount","value":"1000uluna"}]},{"type":"execute","attributes":[{"key":"_contract_address","value":"terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9"}]},{"type":"wasm","attributes":[{"key":"_contract_address","value":"terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9"},{"key":"chain_id","value":"34"},{"key":"chain_address","value":"00000000000000000000000024850c6f61c438823f01b7a3bf2b89b72174fa9d"}]}]}]',
              info: "",
              gas_wanted: "603219",
              gas_used: "366430",
              events: [
                {
                  type: "coin_spent",
                  attributes: [
                    {
                      key: "spender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "amount", value: "17086200uluna", index: true },
                  ],
                },
                {
                  type: "coin_received",
                  attributes: [
                    {
                      key: "receiver",
                      value: "terra17xpfvakm2amg962yls6f84z3kell8c5lkaeqfa",
                      index: true,
                    },
                    { key: "amount", value: "17086200uluna", index: true },
                  ],
                },
                {
                  type: "transfer",
                  attributes: [
                    {
                      key: "recipient",
                      value: "terra17xpfvakm2amg962yls6f84z3kell8c5lkaeqfa",
                      index: true,
                    },
                    {
                      key: "sender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "amount", value: "17086200uluna", index: true },
                  ],
                },
                {
                  type: "message",
                  attributes: [
                    {
                      key: "sender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                  ],
                },
                {
                  type: "tx",
                  attributes: [
                    { key: "fee", value: "17086200uluna", index: true },
                    {
                      key: "fee_payer",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                  ],
                },
                {
                  type: "tx",
                  attributes: [
                    {
                      key: "acc_seq",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx/3",
                      index: true,
                    },
                  ],
                },
                {
                  type: "tx",
                  attributes: [
                    {
                      key: "signature",
                      value:
                        "v+D5taqVhgfTBJycR5Jclsyo/HeGSpEBpipViK29OLNcKqyieC9qugfU90cBd2heIZttflDpSfVI2HNrMmFNgw==",
                      index: true,
                    },
                  ],
                },
                {
                  type: "message",
                  attributes: [
                    { key: "action", value: "/cosmwasm.wasm.v1.MsgExecuteContract", index: true },
                    {
                      key: "sender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "module", value: "wasm", index: true },
                  ],
                },
                {
                  type: "coin_spent",
                  attributes: [
                    {
                      key: "spender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "amount", value: "1000uluna", index: true },
                  ],
                },
                {
                  type: "coin_received",
                  attributes: [
                    {
                      key: "receiver",
                      value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
                      index: true,
                    },
                    { key: "amount", value: "1000uluna", index: true },
                  ],
                },
                {
                  type: "transfer",
                  attributes: [
                    {
                      key: "recipient",
                      value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
                      index: true,
                    },
                    {
                      key: "sender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "amount", value: "1000uluna", index: true },
                  ],
                },
                {
                  type: "execute",
                  attributes: [
                    {
                      key: "_contract_address",
                      value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
                      index: true,
                    },
                  ],
                },
                {
                  type: "wasm",
                  attributes: [
                    {
                      key: "_contract_address",
                      value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
                      index: true,
                    },
                    { key: "chain_id", value: "34", index: true },
                    {
                      key: "chain_address",
                      value: "00000000000000000000000024850c6f61c438823f01b7a3bf2b89b72174fa9d",
                      index: true,
                    },
                  ],
                },
              ],
              codespace: "",
            },
            tx: "CugLCuULCiQvY29zbXdhc20ud2FzbS52MS5Nc2dFeGVjdXRlQ29udHJhY3QSvAsKLHRlcnJhMThreTZ2ZzZxdXZlcmZlM3ZobWxtbGtqdDRnOG56OHA4ejBsMHJ4EkB0ZXJyYTE1MzM2NnE1MGs3dDhubjdnZWMwMGhnNjZjcm5oa2RnZ3BnZHRheGx0YXE2eHJ1dGtrejNzOTkyZnc5GroKeyJzdWJtaXRfdmFhIjp7ImRhdGEiOiJBUUFBQUFRTkFkL0wrVkMvb2JJaU1uTnJpTjVhc0FwTVZOV1FIYjQ0YU82YmJKUndRWXhMWkd2MHU3b0d6cWFlQjQ3aTFGMkFrUHVSQXYyMDVYN2NQMXRlZzk2ODVha0JBaEdoUXFHQmgzVmh6NGdNM0tUdGRaRjIyalNSNDcySEVGQ0RESHJsVWR2TEVpQkN2MlBoaVIwd1lSbE9BVGRQZlZqbmF0M3dBbEhNNkV2bjBLUTVGVllCQTZRQ0ZWWlpvL28zU3laSU03Q3hDeGVXNCtFOXRSb0RWK2wxZ1E2WjVERGRFYklPenBkSzkzM1A3K3RzVmJjWGxaaUlOSVdMVnB2dDJ3MS9JQ2liZVBZQUJNS2tWZG85SUZzWFZ1Y2NmSWRLV0pTTjIxK3YyOHBnQ1YyRktzV3MxOGxtS1NBbklGTGwrVUUrKzhza0xEQ3ZEellFenh4Mnduc3Z0YVhUb1RHUU9jOEFCZVVKMVBHRTNydWdZaERDMEI1TE8yRGhMamJ2ZmZld0Z5Qjd4c1lYWThURlhjYklReE1FWlBzSElkTlZUaHlWRTVYZ3RLNEhrcjVLRjNsR094TU80RU1BQnVRSDV5RDVWTmdndG9JcUVJUXA3TGdlbXlQemVQWG9PUjRCbXhQNldkRDJhQkdZR0l2ZXZXQXdOY0dGbTBnVkZZVWZON2lzU1p6R0VYUHgwV2JiRmc4QkJ5OTlSZ1JsS3p6eUtNdjdJQWZvVTl5VldiWkI3QXAxSjdYZFp2NlFPVkZEZDRmVkVwZUVMVVpnTE4rMmYrd01HeWYzdThxSk5KeFZLQnRyclRUSi80c0JDUVlXNVo3L0h6Sm10TjA5eDdTN3NDNWVyY0o3WmdxRFU4QUZKdmdWWEZuMFVSOW5pRStEQ21vYTQxY0NVQWEyYVR1QWMwN3VrYjhMa0hFdmIvcVJWdHdCQ3k2ayszYjZmOFNKQ28wYmN0dUt6eXRWMmpmbURrYU9QYmJWby9YOWhKS3dDZlQrL1Z1RCtzbjhZRi95d0Zyd1JPSXpuZ1ZzK2RsTzkwV09aeDlNK3FNQURmLzBqM2s1N21Mei85Nzl0K0MxZTNQWlRwQmVnTFJTelptWTVyOS8xR1BzYTc5bElkaElJY2hhK3lBU3hZZm9TdTFqOWE1SVNOWFEzZUVVTGtVQXZlVUJEczIyQ2RwUGFuc0toS2gwK3QzNklOTStaRXZYaEdGK2NiejVUQ3BJV0RTSUV5aTJ5b1BTYzQxc0FUb052Kyt6bkN2cW5lcGNDL3BoR3Q3VnhoWktFZUlBRC9vQUJ2V1JYc3RkVEJYKzBRVzBhcWVtVDVJSTNMRGd1ZXRsU2xMQWE3Rm9aWm9RVVRQamRMNVl2b3BhWmlCWEt4Q2k1N2NyMnd0azc1bHQ5NnhOUVBZQUVHL0k2L3EwNnYrc0hmVXZqN1BKeDVNemtsNFMwYVFYeE1USFFiRlAvSXhpS2x3dzBtZU90clRxZ2F3Y21WUDQ5TjlzMVhBb2liV0ljK0kyV1dXem9COEJBQUFBQUQveXdBOEFBUUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBRW1HeitIeWxBR3NjZ0FBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFWRzlyWlc1Q2NtbGtaMlVCQUFBQUlnQUFBQUFBQUFBQUFBQUFBQ1NGREc5aHhEaUNQd0czbzc4cmliY2hkUHFkIn19Kg0KBXVsdW5hEgQxMDAwEmsKUApGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQMvrTVjJn2CqTyvFnhVuMifkn4uGR0G9eT8Llme7hIGfBIECgIIARgDEhcKEQoFdWx1bmESCDE3MDg2MjAwENPoJBpAv+D5taqVhgfTBJycR5Jclsyo/HeGSpEBpipViK29OLNcKqyieC9qugfU90cBd2heIZttflDpSfVI2HNrMmFNgw==",
          },
          {
            hash: "7A6D586EE1124858B40B2D77734BA8CC0FFFD931C915620F62A399DB59510919",
            height: "10252321",
            index: 0,
            tx_result: {
              code: 0,
              data: "Ei4KLC9jb3Ntd2FzbS53YXNtLnYxLk1zZ0V4ZWN1dGVDb250cmFjdFJlc3BvbnNl",
              log: '[{"msg_index":0,"events":[{"type":"message","attributes":[{"key":"action","value":"/cosmwasm.wasm.v1.MsgExecuteContract"},{"key":"sender","value":"terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx"},{"key":"module","value":"wasm"}]},{"type":"coin_spent","attributes":[{"key":"spender","value":"terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx"},{"key":"amount","value":"1000uluna"}]},{"type":"coin_received","attributes":[{"key":"receiver","value":"terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9"},{"key":"amount","value":"1000uluna"}]},{"type":"transfer","attributes":[{"key":"recipient","value":"terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9"},{"key":"sender","value":"terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx"},{"key":"amount","value":"1000uluna"}]},{"type":"execute","attributes":[{"key":"_contract_address","value":"terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9"}]},{"type":"wasm","attributes":[{"key":"_contract_address","value":"terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9"},{"key":"chain_id","value":"36"},{"key":"chain_address","value":"00000000000000000000000024850c6f61c438823f01b7a3bf2b89b72174fa9d"}]}]}]',
              info: "",
              gas_wanted: "654880",
              gas_used: "369135",
              events: [
                {
                  type: "coin_spent",
                  attributes: [
                    {
                      key: "spender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "amount", value: "9824uluna", index: true },
                  ],
                },
                {
                  type: "coin_received",
                  attributes: [
                    {
                      key: "receiver",
                      value: "terra17xpfvakm2amg962yls6f84z3kell8c5lkaeqfa",
                      index: true,
                    },
                    { key: "amount", value: "9824uluna", index: true },
                  ],
                },
                {
                  type: "transfer",
                  attributes: [
                    {
                      key: "recipient",
                      value: "terra17xpfvakm2amg962yls6f84z3kell8c5lkaeqfa",
                      index: true,
                    },
                    {
                      key: "sender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "amount", value: "9824uluna", index: true },
                  ],
                },
                {
                  type: "message",
                  attributes: [
                    {
                      key: "sender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                  ],
                },
                {
                  type: "tx",
                  attributes: [
                    { key: "fee", value: "9824uluna", index: true },
                    {
                      key: "fee_payer",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                  ],
                },
                {
                  type: "tx",
                  attributes: [
                    {
                      key: "acc_seq",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx/4",
                      index: true,
                    },
                  ],
                },
                {
                  type: "tx",
                  attributes: [
                    {
                      key: "signature",
                      value:
                        "RliBm5HavkanE/63buiEeAiKabPpQXQXDVZIQ/tA99QRgkIQILaxzg++gLQbAe+0h3tIKSXOZKHfsVTNBxCjuA==",
                      index: true,
                    },
                  ],
                },
                {
                  type: "message",
                  attributes: [
                    { key: "action", value: "/cosmwasm.wasm.v1.MsgExecuteContract", index: true },
                    {
                      key: "sender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "module", value: "wasm", index: true },
                  ],
                },
                {
                  type: "coin_spent",
                  attributes: [
                    {
                      key: "spender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "amount", value: "1000uluna", index: true },
                  ],
                },
                {
                  type: "coin_received",
                  attributes: [
                    {
                      key: "receiver",
                      value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
                      index: true,
                    },
                    { key: "amount", value: "1000uluna", index: true },
                  ],
                },
                {
                  type: "transfer",
                  attributes: [
                    {
                      key: "recipient",
                      value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
                      index: true,
                    },
                    {
                      key: "sender",
                      value: "terra18ky6vg6quverfe3vhmlmlkjt4g8nz8p8z0l0rx",
                      index: true,
                    },
                    { key: "amount", value: "1000uluna", index: true },
                  ],
                },
                {
                  type: "execute",
                  attributes: [
                    {
                      key: "_contract_address",
                      value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
                      index: true,
                    },
                  ],
                },
                {
                  type: "wasm",
                  attributes: [
                    {
                      key: "_contract_address",
                      value: "terra153366q50k7t8nn7gec00hg66crnhkdggpgdtaxltaq6xrutkkz3s992fw9",
                      index: true,
                    },
                    { key: "chain_id", value: "36", index: true },
                    {
                      key: "chain_address",
                      value: "00000000000000000000000024850c6f61c438823f01b7a3bf2b89b72174fa9d",
                      index: true,
                    },
                  ],
                },
              ],
              codespace: "",
            },
            tx: "CugLCuULCiQvY29zbXdhc20ud2FzbS52MS5Nc2dFeGVjdXRlQ29udHJhY3QSvAsKLHRlcnJhMThreTZ2ZzZxdXZlcmZlM3ZobWxtbGtqdDRnOG56OHA4ejBsMHJ4EkB0ZXJyYTE1MzM2NnE1MGs3dDhubjdnZWMwMGhnNjZjcm5oa2RnZ3BnZHRheGx0YXE2eHJ1dGtrejNzOTkyZnc5GroKeyJzdWJtaXRfdmFhIjp7ImRhdGEiOiJBUUFBQUFRTkFWU2REckNqWlk2SGVGVU82YnZla3VHNEJabjhtVTNZYlhhaFF3ak9lK0tNTUxPdHRCQ3ptbXYvYmtHNjJzaTlxWExRRXRSZVVBa1crbklDdElCSzVXNEJBcDZuaG9yNkpOYXNqVklXak1ZWXp0YTdSQTBEV1dJWno4cGx5cG54RW82dUk1Rm5RWkRiVGdINCtYcHpSeHlvY2FIcW9oVkhjK2FwV3l1R1JxcG1TQ0lCQXdacDRhQ2tGKzVnWU41T2xSTUVOMmNaWkJTRjZxMEJyZC9ENGMxNlNqNWdHSEd0QkFRbEN4L0VRRG1wRjFaMHh6cmk5d2VDd3RjR1lLaWZSUTljNktjQkJQY2F2dWFMMEl1OStDVVlUMzMxajFXbWlmOW8wem5FQ3RIdUxSSVptTUMxRk1FRFhpZFdtbnU3TGNYaHZ1NUtnd1RrSTh2Z1ZtdTF1WUxBd2JTQzR3Z0FCV3gyQ3NDVUtEUkFjejdBRDJWTkRmcC9ycEg2VFdNSGE4QXJHSFBuSG53eFNuamdJUjRuWVBSQkFaZ0c0bVdINkhaSDFoTkJPcmZNdUFGSXRXS3pQT0VBQmx3NVNVeVRrU3hsako2SUxWdDBGbjBEL0xmZGFVeUFwREoxVlBjWUZkNXhQWWYxMzBudXVjUXJvaVZRcEFSVWMwNWcyU2hiQmtrcXE0cEV2OEE2SzVZQUIvNFNuRlFDM2xrNi84bU41ZXFOVWFiRlVSNTZtSjZlWkFVUVZ1ejBvYXErZDRvU1ppNEY4WXk2dUJwY1hvbFAzb3hvU0twRXVqRGtlc05yOXJwdTg0OEFDYjg4ZFBDUXd2WWw5LzJxa1lVSGYvVmM4Q0cxa3lvdUJGTjJ4b0pMeG90Z0h2VlZSNEpsMHBIVTZmcTAvVndWQlc3QWxqVU5BR1p3SVhhbk51U2JzZ3dCQzA0UUtNTlp5QzdlekIvSUpyMU5NRERkRTZ4Y0kycHRwdUxPM3NuLzZyQzdVcy83UFJsSGZxRkxZRWxqRm9aWEhzRGlONU43aTJmbEx2VzlJOVUvanc0QkRTWkdob0dSV0lUVExBQ0ZENW42OHNmWk9MdGdOSjgrcjhaLzZRVW9CZ1FiQmV1Y01rdXhkSS95UTNnaGJCVE5kN09QL1YvT3hsdE1wZjMreTdNcVJPVUFEblJndDlGQ2tIQ2ZiNmwvSkdSN2piUXFuVWVwcmZNOEF6WXRqNjR3U0JVeENFSno1cWJvT281RTdsem1rSW5oMXE5SmR5SDd2Zk9veEY4bkR0aHQvb1VBRCt2VHdISTdqMlpFdU9jeEZScUVLVURGK1hnejRxZnIzclBucXZmUkhFVU1XWjdxd2xBaUFkYmJucXM2VUROQ3VnSWNNdkpJWVdpbVZpRkxDVS9ydXJVQUVJOXpqOW5RNHpnYU9TU2lhR2NqdmNuZVpPVU1IZ0E4Y0xXRG1KaTZPM0drRS9zWmJLTDl1aW9oY2IySG91VnFyL2Uvc21wUHRJVnJCMk55YTl0VjZBc0FBQUFBQUZYTFRXc0FBUUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBRVdrS1R4YklrRXFRZ0FBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFBQUFWRzlyWlc1Q2NtbGtaMlVCQUFBQUpBQUFBQUFBQUFBQUFBQUFBQ1NGREc5aHhEaUNQd0czbzc4cmliY2hkUHFkIn19Kg0KBXVsdW5hEgQxMDAwEmcKUApGCh8vY29zbW9zLmNyeXB0by5zZWNwMjU2azEuUHViS2V5EiMKIQMvrTVjJn2CqTyvFnhVuMifkn4uGR0G9eT8Llme7hIGfBIECgIIARgEEhMKDQoFdWx1bmESBDk4MjQQoPwnGkBGWIGbkdq+RqcT/rdu6IR4CIpps+lBdBcNVkhD+0D31BGCQhAgtrHOD76AtBsB77SHe0gpJc5kod+xVM0HEKO4",
          },
        ],
        total_count: "63",
      },
    });
};

const givenBlockHeightIs = () => {
  nock(rpc)
    .post("/block?height=80542798")
    .reply(200, {
      jsonrpc: "2.0",
      id: -1,
      result: {
        block_id: {
          hash: "8E241F058343F48F934C9DF5E0BA311142FF7030F154BD6F08E37BBDBB814440",
          parts: {
            total: 1,
            hash: "97C2A3D3E279E36733F35BD13B1C9EB1172A8881D41A8117AA2AF590AB466A6D",
          },
        },
        block: {
          header: {
            version: { block: "11" },
            chain_id: "phoenix-1",
            height: "10823140",
            time: "2024-06-18T14:51:07.396526491Z",
            last_block_id: {
              hash: "BD8154BD0984C61FB22BAEC33040AE7E22361203DDC447A0916B0F2CCEC1C639",
              parts: {
                total: 1,
                hash: "DEE8E86658AB965C31890B8A5A7EA70923D0157227B1C0A588527A1B30E11B97",
              },
            },
            last_commit_hash: "C3DBF041B4AFF44EC622E9360BD22BF51B1D9A362AEBBFAA8D54E64C8F8D6700",
            data_hash: "2998B57A3A93052F9A7E110D05388331B543A38969D940BCE892421DC514A969",
            validators_hash: "207731F082E90C7405D884B7819705744B8514CE888C92792D22A463C068756E",
            next_validators_hash:
              "207731F082E90C7405D884B7819705744B8514CE888C92792D22A463C068756E",
            consensus_hash: "E660EF14A95143DB0F3EAF2F31F177DE334DE5AB650579FD093A10CBAE86D5A6",
            app_hash: "5AEBD7E54F9FC7D05A477E99199F2AC468AA54225705FD86C9335B668DBA9511",
            last_results_hash: "D47A43587D6B730AE26585A8500AB9117DC057C11C1CA760F465AD45FE291E83",
            evidence_hash: "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855",
            proposer_address: "F8D93F465274628E94F3DAB43C05F3BA691D800E",
          },
        },
      },
    });
};
