import { TransactionFoundEvent } from "../../../domain/entities";
import { AptosEvent } from "../../../domain/entities/aptos";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "aptosRedeemedTransactionFoundMapper" });

const REDEEM_EVENT_TAIL = "::complete_transfer::TransferRedeemed";
const APTOS_CHAIN = "aptos";

export const aptosRedeemedTransactionFoundMapper = (
  tx: AptosEvent[]
): TransactionFoundEvent | undefined => {
  return undefined;
};
