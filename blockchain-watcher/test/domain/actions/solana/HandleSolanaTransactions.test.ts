import { describe, jest, it, expect } from "@jest/globals";
import { StatRepository } from "../../../../src/domain/repositories";
import { solana } from "../../../../src/domain/entities";
import {
  HandleSolanaTransactions,
  HandleSolanaTxConfig,
} from "../../../../src/domain/actions/solana/HandleSolanaTransactions";

let solanaTxs: solana.Transaction[];
let statsRepo: StatRepository;

describe("HandleSolanaTransactions", () => {
  let handleSolanaTransactions: HandleSolanaTransactions<any>;
  const mockConfig: HandleSolanaTxConfig = {
    environment: "mainnet",
    metricName: "process_source_solana_event",
    programId: "mockProgramId",
    programs: { mockProgramId: ["0a"] },
    commitment: "finalized",
    chainId: 1,
    chain: "solana",
    abi: "",
    id: "poll-log-message-published-solana",
  };

  it("should handle Solana transactions", async () => {
    givenStatsRepository();
    givenSolanaTransactions();
    const txMapped = {
      name: "transfer-redeemed",
      address: "1231231231234312412312",
      chainId: 1,
      txHash: "fasifoasfoasojfjasjdjasdaksdkad",
      blockHeight: BigInt("124123".toString()),
      blockTime: 1231211,
      attributes: {
        methodsByAddress: "unknownInstruction",
        status: "completed",
        emitterChain: 1232,
        emitterAddress: "asdasdSS222sdasSDSD2231232",
        sequence: 1232,
        protocol: "unknown",
      },
    };
    const mockTarget = jest.fn<(parsed: any[]) => Promise<void>>();
    handleSolanaTransactions = new HandleSolanaTransactions<any>(
      mockConfig,
      async () => {
        return [txMapped];
      },
      mockTarget,
      statsRepo
    );

    const result = await handleSolanaTransactions.handle(solanaTxs);

    expect(result).toEqual([txMapped]);
  });

  it("should handle Solana transactions with a target", async () => {
    givenStatsRepository();
    givenSolanaTransactions();
    const txMapped = {
      name: "transfer-redeemed",
      address: "1231231231234312412312",
      chainId: 1,
      txHash: "fasifoasfoasojfjasjdjasdaksdkad",
      blockHeight: BigInt("124123".toString()),
      blockTime: 1231211,
      attributes: {
        methodsByAddress: "completeNativeInstruction",
        status: "completed",
        emitterChain: 1232,
        emitterAddress: "asdasdSS222sdasSDSD2231232",
        sequence: 1232,
        protocol: "Token Bridge",
      },
    };
    const mockTarget = jest.fn<(parsed: any[]) => Promise<void>>();
    handleSolanaTransactions = new HandleSolanaTransactions<any>(
      mockConfig,
      async () => {
        return [txMapped];
      },
      mockTarget,
      statsRepo
    );
    const mockTransactions: solana.Transaction[] = await handleSolanaTransactions.handle(solanaTxs);

    expect(mockTarget).toHaveBeenCalledWith(mockTransactions);
  });
});

const givenStatsRepository = () => {
  statsRepo = {
    count: () => {},
    measure: () => {},
    report: () => Promise.resolve(""),
  };
};

const givenSolanaTransactions = () =>
  (solanaTxs = [
    {
      slot: 1,
      chainId: 1,
      chain: "solana",
      transaction: {
        message: {
          accountKeys: [],
          instructions: [],
          compiledInstructions: [],
        },
        signatures: [],
      },
    },
  ]);
