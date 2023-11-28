import winston from "winston";
import { configuration } from "./config";

winston.remove(winston.transports.Console);
winston.configure({
  transports: [
    new winston.transports.Console({
      level: configuration.logLevel,
    }),
  ],
  format: winston.format.combine(
    winston.format.colorize(),
    winston.format.splat(),
    winston.format.errors({ stack: true }),
    winston.format.printf(
      ({ level, message, module, chain, label }) =>
        `${level} [${module ?? ""}]${chain ? `[${chain}]` : ""}${
          label ? `[${label}]` : ""
        } ${message}`
    )
  ),
});

export default winston;
