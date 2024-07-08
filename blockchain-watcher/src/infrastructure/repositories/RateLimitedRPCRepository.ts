import { Circuit, Ratelimit, Retry, RetryMode } from "mollitia";
import { Options } from "./common/rateLimitedOptions";
import winston from "winston";

export abstract class RateLimitedRPCRepository<T> {
  protected delegate: T;
  protected breaker: Circuit;
  protected logger: winston.Logger = winston.child({
    module: "RateLimitedRPCRepository",
  });

  constructor(delegate: T, opts: Options = { period: 10_000, limit: 1_000 }) {
    this.delegate = delegate;
    this.breaker = new Circuit({
      options: {
        modules: [
          new Ratelimit({ limitPeriod: opts.period, limitForPeriod: opts.limit }),
          new Retry({
            attempts: 10,
            interval: 1_000,
            fastFirst: false,
            mode: RetryMode.EXPONENTIAL,
            factor: 1,
            onRejection: (err: Error | any) => {
              if (err.message?.includes("429")) {
                this.logger.warn("Got 429 from RPC node. Retrying in 5 secs...");
                return 5_000; // Wait 5 secs if we get a 429
              } else if (err.message?.includes("healthy providers")) {
                this.logger.warn("Got no healthy providers from RPC node. Retrying in 5 secs...");
                return 5_000; // Wait 5 secs if we get a no healthy providers
              } else {
                this.logger.warn("Retry according to config...");
                return true; // Retry according to config
              }
            },
          }),
        ],
      },
    });
  }
}
