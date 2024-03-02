import { LogFoundEvent, LogMessagePublished } from "../../../domain/entities";
import { AptosEvent } from "../../../domain/entities/aptos";
import winston from "winston";

let logger: winston.Logger = winston.child({ module: "aptosLogMessagePublishedMapper" });

const SOURCE_EVENT_TAIL = "::state::WormholeMessage";

export const aptosLogMessagePublishedMapper = (
  tx: AptosEvent[]
): LogFoundEvent<LogMessagePublished> | undefined => {
  return undefined;
};
