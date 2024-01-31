import { describe, jest, it, expect } from "@jest/globals";
import {
  HandleSolanaTransactions,
  HandleSolanaTxConfig,
} from "../../../../src/domain/actions/solana/HandleSolanaTransactions";
import { solana } from "../../../../src/domain/entities";
import { StatRepository } from "../../../../src/domain/repositories";

let solanaTxs: solana.Transaction[];
let statsRepo: StatRepository;

describe("HandleSolanaTransactions", () => {
  let handleSolanaTransactions: HandleSolanaTransactions<any>;
  const mockConfig: HandleSolanaTxConfig = {
    programId: "mockProgramId",
    commitment: "finalized",
    chainId: 1,
    chain: "solana",
    abi: "",
    id: "poll-log-message-published-solana",
  };

  it("should handle Solana transactions", async () => {
    givenStatsRepository();
    givenSolanaTransactions();
    const mockTarget = jest.fn<(parsed: any[]) => Promise<void>>();
    handleSolanaTransactions = new HandleSolanaTransactions<any>(
      mockConfig,
      async (tx: solana.Transaction) => {
        return [tx];
      },
      mockTarget,
      statsRepo
    );

    const result = await handleSolanaTransactions.handle(solanaTxs);

    expect(result).toEqual(solanaTxs);
  });

  it("should handle Solana transactions with a target", async () => {
    givenStatsRepository();
    givenSolanaTransactions();
    const mockTarget = jest.fn<(parsed: any[]) => Promise<void>>();
    handleSolanaTransactions = new HandleSolanaTransactions<any>(
      mockConfig,
      async (tx: solana.Transaction) => {
        return [tx];
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
