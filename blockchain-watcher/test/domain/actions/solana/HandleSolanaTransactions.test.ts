import { describe, jest, it, expect } from "@jest/globals";
import {
  HandleSolanaTransactions,
  HandleSolanaTxConfig,
} from "../../../../src/domain/actions/solana/HandleSolanaTransactions";
import { solana } from "../../../../src/domain/entities";

let solanaTxs: solana.Transaction[];

describe("HandleSolanaTransactions", () => {
  let handleSolanaTransactions: HandleSolanaTransactions<any>;
  const mockConfig: HandleSolanaTxConfig = {
    programId: "mockProgramId",
  };

  it("should handle Solana transactions", async () => {
    givenSolanaTransactions();
    handleSolanaTransactions = new HandleSolanaTransactions<any>(
      mockConfig,
      async (tx: solana.Transaction) => {
        return [tx];
      }
    );

    const result = await handleSolanaTransactions.handle(solanaTxs);

    expect(result).toEqual(solanaTxs);
  });

  it("should handle Solana transactions with a target", async () => {
    givenSolanaTransactions();
    const mockTarget = jest.fn<(parsed: any[]) => Promise<void>>();
    handleSolanaTransactions = new HandleSolanaTransactions<any>(
      mockConfig,
      async (tx: solana.Transaction) => {
        return [tx];
      },
      mockTarget
    );
    const mockTransactions: solana.Transaction[] = await handleSolanaTransactions.handle(solanaTxs);

    expect(mockTarget).toHaveBeenCalledWith(mockTransactions);
  });
});

const givenSolanaTransactions = () =>
  (solanaTxs = [
    {
      slot: 1,
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
