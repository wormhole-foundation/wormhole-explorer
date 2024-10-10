import { afterEach, describe, it, expect, jest } from "@jest/globals";
import { thenWaitForAssertion } from "../../../waitAssertion";
import { PollNear, PollNearConfig, PollNearMetadata } from "../../../../src/domain/actions";
import {
  MetadataRepository,
  NearRepository,
  StatRepository,
} from "../../../../src/domain/repositories";
import { NearTransaction } from "../../../../src/domain/entities/near";

let getBlockHeightSpy: jest.SpiedFunction<NearRepository["getBlockHeight"]>;
let getTransactionsSpy: jest.SpiedFunction<NearRepository["getTransactions"]>;
let metadataSaveSpy: jest.SpiedFunction<MetadataRepository<PollNearMetadata>["save"]>;

let handlerSpy: jest.SpiedFunction<(txs: NearTransaction[]) => Promise<void>>;

let metadataRepo: MetadataRepository<PollNearMetadata>;
let nearRepo: NearRepository;
let statsRepo: StatRepository;

let handlers = {
  working: (txs: NearTransaction[]) => Promise.resolve(),
  failing: (txs: NearTransaction[]) => Promise.reject(),
};

let pollNear: PollNear;

let cfg = new PollNearConfig({
  chain: "near",
  contracts: ["842125965"],
  chainId: 8,
  environment: "testnet",
  commitment: "final",
});

describe("GetNearTransactions", () => {
  afterEach(async () => {
    await pollNear.stop();
  });

  it("should be use from and batch size cfg, and process tx because is a wormhole redeem", async () => {
    // Given
    const txs = [
      {
        receiverId: "contract.portalbridge.near",
        signerId: "tkng.near",
        timestamp: 1722257013,
        blockHeight: "124531378",
        chainId: 15,
        hash: "DMqXkWDFGv59x5z3QpdmtPM1aYZCCKyeMGDasZgVdRj",
        logs: [
          {
            proof: [
              { hash: "D8VYAYJAMiCWCacv6E3xTUXpwPA2PjyJNzRoYPh53SFG", direction: "Left" },
              { hash: "9itDx2LLBVSfPLewvWHJ6SXNAjX56KQpdzNtCVZeMSbD", direction: "Right" },
              { hash: "6GZsLF5sGqywD6zispB9bxN89wfA94oVKTnivrKk32L1", direction: "Left" },
              { hash: "8AXsPrreyfZVahPex3jxr3ixFx8PtsEwinNasSPNPp9C", direction: "Left" },
              { hash: "X3xKP66DbBGkzoATAi95xbwD45vbckC9dTYuUyzGBbD", direction: "Right" },
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
              status: { SuccessReceiptId: "45zwbg6ukpYDW8HCbkDSNbnHUgGTH1SY7zPgaNUvChMa" },
              metadata: {
                version: 3,
                gas_profile: [
                  {
                    cost_category: "ACTION_COST",
                    cost: "FUNCTION_CALL_BASE",
                    gas_used: "600000000000",
                  },
                  {
                    cost_category: "ACTION_COST",
                    cost: "FUNCTION_CALL_BYTE",
                    gas_used: "5249973032",
                  },
                  {
                    cost_category: "ACTION_COST",
                    cost: "NEW_ACTION_RECEIPT",
                    gas_used: "470125429248",
                  },
                  { cost_category: "WASM_HOST_COST", cost: "BASE", gas_used: "8207811441" },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "CONTRACT_LOADING_BASE",
                    gas_used: "35445963",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "CONTRACT_LOADING_BYTES",
                    gas_used: "1223704199345",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "KECCAK256_BASE",
                    gas_used: "5879491275",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "KECCAK256_BYTE",
                    gas_used: "3950683320",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "PROMISE_RETURN",
                    gas_used: "560152386",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_CACHED_TRIE_NODE",
                    gas_used: "11400000000",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_MEMORY_BASE",
                    gas_used: "46977537600",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_MEMORY_BYTE",
                    gas_used: "10860408381",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_REGISTER_BASE",
                    gas_used: "15102991116",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_REGISTER_BYTE",
                    gas_used: "231029328",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_HAS_KEY_BASE",
                    gas_used: "54039896625",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_HAS_KEY_BYTE",
                    gas_used: "1139261265",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_READ_BASE",
                    gas_used: "56356845750",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_READ_KEY_BYTE",
                    gas_used: "154762665",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_READ_VALUE_BYTE",
                    gas_used: "813595725",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_BASE",
                    gas_used: "64196736000",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_EVICTED_BYTE",
                    gas_used: "4657009515",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_KEY_BYTE",
                    gas_used: "352414335",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_VALUE_BYTE",
                    gas_used: "4497688155",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "TOUCHING_TRIE_NODE",
                    gas_used: "322039118520",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "UTF8_DECODING_BASE",
                    gas_used: "9335337183",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "UTF8_DECODING_BYTE",
                    gas_used: "23618018799",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WASM_INSTRUCTION",
                    gas_used: "319104269088",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_MEMORY_BASE",
                    gas_used: "25234153749",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_MEMORY_BYTE",
                    gas_used: "6515262624",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_REGISTER_BASE",
                    gas_used: "20058657402",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_REGISTER_BYTE",
                    gas_used: "9462092796",
                  },
                ],
              },
            },
          },
          {
            proof: [
              { hash: "3RZbxRiPbAQ8JsMDkknnuvRg54WxcjLUmhSJsYRoQVqH", direction: "Left" },
              { hash: "CuuCGepQBLpFi1u68Z9qSH35RcHN2DeGrTLPq8knYMmc", direction: "Right" },
              { hash: "5Swz8ixMCS2uyBazaSxxXfEP7zK6ovxgosJh6maX8ZiN", direction: "Left" },
              { hash: "5N6uVA1uC72Xim2Ec7icLVn2yDUTjGq6GzR5H7j9n8bz", direction: "Left" },
              { hash: "3DFAY9VEEcV1LT2vV2fukL6EZGeLmMEFLQWAPHDt7iR2", direction: "Left" },
              { hash: "8ekpFsNXDTaGBkrrjKrgyCwqv2XDY241S4AxstN1TQEJ", direction: "Right" },
            ],
            block_hash: "34cTwPWkjxuTG9aJ5oGF9Pp8ioDrA1KV7Ah3qapmnLMU",
            id: "5d48CgVXxc38ck5769qY79bVtTw7UvXA9JFttdy1uaoD",
            outcome: {
              logs: [
                "token-bridge/src/lib.rs#1265: refunding 99220000000000000000000 to tkng.near?",
              ],
              receipt_ids: [
                "5AkS8bx76bMGGCjHgJMwp2KnJdaV8Rc9yL6EDvZqYKQg",
                "8wiGQ74DoecsRfa2bqvGXRgph6CM87pNbqvGwssXuPoH",
              ],
              gas_burnt: 3289563780780,
              tokens_burnt: "328956378078000000000",
              executor_id: "contract.portalbridge.near",
              status: { SuccessReceiptId: "5AkS8bx76bMGGCjHgJMwp2KnJdaV8Rc9yL6EDvZqYKQg" },
              metadata: {
                version: 3,
                gas_profile: [
                  {
                    cost_category: "ACTION_COST",
                    cost: "NEW_ACTION_RECEIPT",
                    gas_used: "108059500000",
                  },
                  {
                    cost_category: "ACTION_COST",
                    cost: "NEW_DATA_RECEIPT_BYTE",
                    gas_used: "137696088",
                  },
                  { cost_category: "ACTION_COST", cost: "TRANSFER", gas_used: "115123062500" },
                  { cost_category: "WASM_HOST_COST", cost: "BASE", gas_used: "6883970886" },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "CONTRACT_LOADING_BASE",
                    gas_used: "35445963",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "CONTRACT_LOADING_BYTES",
                    gas_used: "1223704199345",
                  },
                  { cost_category: "WASM_HOST_COST", cost: "LOG_BASE", gas_used: "3543313050" },
                  { cost_category: "WASM_HOST_COST", cost: "LOG_BYTE", gas_used: "1016306907" },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "PROMISE_RETURN",
                    gas_used: "560152386",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_CACHED_TRIE_NODE",
                    gas_used: "54720000000",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_MEMORY_BASE",
                    gas_used: "26098632000",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_MEMORY_BYTE",
                    gas_used: "1277247888",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_REGISTER_BASE",
                    gas_used: "12585825930",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_REGISTER_BYTE",
                    gas_used: "34201014",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_HAS_KEY_BASE",
                    gas_used: "54039896625",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_HAS_KEY_BYTE",
                    gas_used: "1139261265",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_READ_BASE",
                    gas_used: "56356845750",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_READ_KEY_BYTE",
                    gas_used: "154762665",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_READ_VALUE_BYTE",
                    gas_used: "813595725",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_BASE",
                    gas_used: "128393472000",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_EVICTED_BYTE",
                    gas_used: "4657009515",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_KEY_BYTE",
                    gas_used: "2960280414",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_VALUE_BYTE",
                    gas_used: "4528706694",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "TOUCHING_TRIE_NODE",
                    gas_used: "483058677780",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "UTF8_DECODING_BASE",
                    gas_used: "6223558122",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "UTF8_DECODING_BYTE",
                    gas_used: "25075921194",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WASM_INSTRUCTION",
                    gas_used: "43072099356",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_MEMORY_BASE",
                    gas_used: "16822769166",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_MEMORY_BYTE",
                    gas_used: "988729236",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_REGISTER_BASE",
                    gas_used: "17193134916",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_REGISTER_BYTE",
                    gas_used: "1870369488",
                  },
                ],
              },
            },
          },
          {
            proof: [
              { hash: "5MxoQCShchs3M2cXk5z9UcKUw7TX2gDmVCXSZocpgo28", direction: "Left" },
              { hash: "3fg8uteb9DidnYXG3SzRNgNpFEu6J6fpruVrvSUW2gwd", direction: "Right" },
              { hash: "7e1BxpoVVTNMkxtWoWGPb37wKgHX9NfmL93Gnz89jrPc", direction: "Right" },
              { hash: "5cTvEzpnuyA2Z2oMNSYLvMtPjDLjUFfq6ncaWmMSWrkp", direction: "Right" },
            ],
            block_hash: "J3YkMeNyw1eMhywA7Ww5RcGieS8dULXU6bymMhUTxQmd",
            id: "45zwbg6ukpYDW8HCbkDSNbnHUgGTH1SY7zPgaNUvChMa",
            outcome: {
              logs: [],
              receipt_ids: ["EmdLrha47UTz26GCA2sK4F6UJXBPWzNtejHHrqqc7NPb"],
              gas_burnt: 2710728822931,
              tokens_burnt: "271072882293100000000",
              executor_id: "contract.portalbridge.near",
              status: { SuccessValue: "" },
              metadata: {
                version: 3,
                gas_profile: [
                  { cost_category: "WASM_HOST_COST", cost: "BASE", gas_used: "4236289776" },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "CONTRACT_LOADING_BASE",
                    gas_used: "35445963",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "CONTRACT_LOADING_BYTES",
                    gas_used: "1223704199345",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_CACHED_TRIE_NODE",
                    gas_used: "2280000000",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_MEMORY_BASE",
                    gas_used: "7829589600",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_MEMORY_BYTE",
                    gas_used: "589206615",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_REGISTER_BASE",
                    gas_used: "10068660744",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "READ_REGISTER_BYTE",
                    gas_used: "24936186",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_READ_BASE",
                    gas_used: "56356845750",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_READ_KEY_BYTE",
                    gas_used: "154762665",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_READ_VALUE_BYTE",
                    gas_used: "813595725",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_BASE",
                    gas_used: "64196736000",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_EVICTED_BYTE",
                    gas_used: "4657009515",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_KEY_BYTE",
                    gas_used: "352414335",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "STORAGE_WRITE_VALUE_BYTE",
                    gas_used: "4497688155",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "TOUCHING_TRIE_NODE",
                    gas_used: "386446942224",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WASM_INSTRUCTION",
                    gas_used: "22824074196",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_MEMORY_BASE",
                    gas_used: "14018974305",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_MEMORY_BYTE",
                    gas_used: "732694668",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_REGISTER_BASE",
                    gas_used: "17193134916",
                  },
                  {
                    cost_category: "WASM_HOST_COST",
                    cost: "WRITE_REGISTER_BYTE",
                    gas_used: "1513022472",
                  },
                ],
              },
            },
          },
        ],
        actions: [
          {
            functionCall: {
              method: "submit_vaa",
              args: "eyJ2YWEiOiIwMTAwMDAwMDA0MGQwMDEzZTA2ZDFkNzgyMDFiZjQwNDhlMDg4MmJmZTJlNjQ5NjE4ODQzYjI1MDg4NjM3NDVlMTdiYTlhYmM4ZWFlMjAzOTA0MmE1ZTY5NmU3ZmU0ZjRiNjNkOGRjMGI4ZDdhNWZkNDNkNWViMjVhYWI2MmIwZDRlMTFjODZjZjUwMTliMDAwMmUxMjAxNzA5Y2Q4NmZiNDhjYWQyZjlhMTY3OGY5NzZlYmZiYzQ5NGZhZmYzNzZhNDRiMmQwNmY2NGQ5MzI5OWQ2MTQzMjJiYWIzZmJkMzFkZGFjMzZkZjQxYjI2MjM1OGI3MjRmYjg2ZDFjN2U0ODY3ZDk5MDkyYjFhZjgzMzY2MDAwNDE1M2VlMjIwZDAxNzlmNzY5ZmUxZjdjNmQ2ODk5ZWQ5NWRhZDAxNzZjMmFhM2FjZTk2ZjQyOWU5MzU5MjQ2NWY3OGM0MTBhM2FkNWMzZDFkNjZmNzkxZjBhZmE5M2QyZTEwMzcyOWMzOGZmMTg4MzQwNjhlOTU0NWMyZDUxNGE1MDAwNjYxMDUwMjk5ZmJkZjJjYmY5OTM0ZTRhMGE3OTk1OWQzMDhiNzE5MWI2OTNjMWVkODM5MWY4MzZiOTZkNWYxZmQ1ZGQ4ZDQyNmM1NWMwZTZhZDBkMTllNWEzYzVhMGQ4OTQ5MjY4ZTg3NDk3MmVhN2MxNjc4MmE3MDhjZTI5ZDNjMDAwN2Q2ZmZhZjYzODFmMGY2ZWI4NGNlODMyOTIwMGFmNTc5YWEyNTE0YjVmYzQxOTcxMDhkNDU3YmIwOTc4OGNkODk0MmYxNWQ4NWJkZDljOGFmNzBlNjIzZjgxOWM2M2IyNDZlZWQyZGYxZjcyZTI3MWJjYjM4MTNiNWNjNmI5MDFiMDAwOGZjNGNjODhiYTg3NzNhYTQ1Mzc2MjQ1OTY2Y2I3OTY5NDE4MWM5ZTRkYTE2ZTU0ZTU4ZjI4M2VmZTU2YmEyMTM0NjY0NWE1YzY1NzVlNWEyNTFjNDM5Yjk0Yzg3YzEwNzFiNjA5YWE0ZjkyMzU3ZjQ0YTU0NzgyM2FjYWRmNjAxMDAwOWFhYWVkNGY1NDZhMzQ4ZWRmYWIyOGExODMwZGNmZmU3ODIyOTU2ZTU1ZmU3OTFhMDQ5YTdkMDMyMThjY2U0OTMxN2IyZTE3MDk4OGY3ZDRkYWU0YzIyY2MwYjg1OWZmMGFiNTlmZGYzYzMzNTVlNDk3ODAzZjJjYzg0ZjM1YjVmMDEwYTRmZTFiMjcxZGU2ODc2YzIzYjQ5MTFkMTk1YmUwZjY1ZDljNmZiYzBiNjAzZThkNzU2ZDBjZjc5ZDM0YTA2YzI2N2IyODk3ZTY3MGU4M2Y3MzEwN2U3ZjQ4ZmNlMWM3YmFjYjAwNDJmN2M1OGZjNzcxYmMxNzA5ZDk1MmQ1NDY1MDEwZWI2ODM2MGRmNDNlM2U4ZTg1ZGVjODdlYzVmYzc3NDhhZWU2NDAxMzMwMzJhYmQxMTZiMjg0OTdmN2YxZGJlYTM2NjU4Yjg5N2FhZjI3NzEzODNlMWNhZDk4NmIyODcyYWRiYjhmNmQxZTI4ZTAwYjdlMGUyZGQ4ODk2MTZkMWJhMDAwZmYzMWNlYTMzZWRmZTBkNzYzOTI3NWE4Y2RkMmIyODRhYWZkMGRhOWFlMjU3NzBmOGEwMzU3NmQyNzEwNGRlNzc2ZDNmMzk3N2Y4NDk1OTAxZGRmYzQ0MGVkYTc1MDkyNGFiNWE1NWM3ZTA3MWI1N2M5ZjE0NWZhNzEyOWY5OTJlMDExMGYzM2JkNTlmZTc5MmFjOWFhNTlmMDNlNjBjYTU3M2VhM2I2MGQyYzVkNjkzZTY5ODM3ZGJlMTI2MDNhZDU1ZmE1YTIyYWNmMjQyNzc0YWJlMWRmNWI0N2E1NjcyNzBhNWRkZWE3YmQ3NTU5Mjg5OWNjMTA2YzYzOWQzMGJjNmIzMDAxMTg1MzY4NjAwYWRjZTBlMjcyZTU2MGNmMTUwNTAzMDNjMTcyOTg0YTcyZjgwNjYzZDkwNmNmZGVkNWZjNzcwYWMxODdkOTA2Y2RkN2IxM2RiMjE0Njc3YmYyNDRmYjk3N2ZjOTc3MDg5ODZjYzViZTdhMDBiMDE5MzNlZTU5ZDAzMDExMjgxN2U2ODRlYzhiZmQ0MWE5ZGU5NjM1ZjBhMTU3NzIzZDUwZjU5Yjg2ODAxNjkyN2ZjZjU2MDc3MWY0OGE4OTU3MjE2MzRjNDEyOTA0MzZhNWFkOWY1OTg2ZGI0NGI1Y2E5N2VkOTk5NWEzYmIxYzUwODAxNTViODQwNGJhNzRhMDE2NmE3OGUyMzAwMDA0MDIzMDAwMWVjNzM3Mjk5NWQ1Y2M4NzMyMzk3ZmIwYWQzNWMwMTIxZTBlYWE5MGQyNmY4MjhhNTM0Y2FiNTQzOTFiM2E0ZjUwMDAwMDAwMDAwMGUwMTFhMjAwMTAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMjNjMzQ2MDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwZmUxM2MzMjA4ZDAyNGQ0MDIwMjEyZDU1OGNmNjAzMDc0ZTk4YmYzMTRkZmRmMTI3Y2MwZTEzOGZiYzcxMmFhZGYwMDBmMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMDAwMCJ9",
            },
          },
        ],
      },
    ];

    givenNearRepository(402222n, txs);
    givenMetadataRepository();
    givenStatsRepository();
    givenPollNearTxs();

    // Whem
    await whenPollNearStarts();

    // Then
    await thenWaitForAssertion(
      () => expect(getBlockHeightSpy).toHaveReturnedTimes(1),
      () => expect(getTransactionsSpy).toBeCalledWith("842125965", 402222n, 402222n)
    );
  });
});

const givenNearRepository = (height?: bigint, txs: any = []) => {
  nearRepo = {
    getBlockHeight: () => Promise.resolve(height),
    getTransactions: () => Promise.resolve(txs),
    healthCheck: () => Promise.resolve([]),
  };

  getBlockHeightSpy = jest.spyOn(nearRepo, "getBlockHeight");
  getTransactionsSpy = jest.spyOn(nearRepo, "getTransactions");
  handlerSpy = jest.spyOn(handlers, "working");
};

const givenMetadataRepository = (data?: PollNearMetadata) => {
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

const givenPollNearTxs = (from?: bigint) => {
  cfg.setFromBlock(from);
  pollNear = new PollNear(nearRepo, metadataRepo, statsRepo, cfg);
};

const whenPollNearStarts = async () => {
  pollNear.run([handlers.working]);
};
