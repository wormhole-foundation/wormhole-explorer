import { Circuit, Ratelimit, Retry, RetryMode } from "mollitia";
import { Options } from "./common/rateLimitedOptions";
import winston from "winston";

export abstract class RateLimitedRPCRepository {
  protected delegate: any;
  protected breaker: Circuit;
  protected logger: winston.Logger = winston.child({
    module: "RateLimitedRPCRepository",
  });

  constructor(delegate: any, opts: Options = { period: 10_000, limit: 1000 }) {
    this.delegate = delegate;
    this.breaker = new Circuit({
      options: {
        modules: [
          new Ratelimit({ limitPeriod: opts.period, limitForPeriod: opts.limit }),
          new Retry({
            attempts: 2,
            interval: 1_000,
            fastFirst: false,
            mode: RetryMode.EXPONENTIAL,
            factor: 1,
            onRejection: (err: Error | any) => {
              if (err.message?.includes("429")) {
                this.logger.warn("Got 429 from evm RPC node. Retrying in 10 secs...");
                return 10_000; // Wait 10 secs if we get a 429
              } else {
                return true; // Retry according to config
              }
            },
          }),
        ],
      },
    });
  }
}
