import { ProviderHealthInstrumentation } from "@xlabs/rpc-pool";
import { HttpClientError } from "../../errors/HttpClientError";
import { AxiosError } from "axios";
import winston from "winston";

// make url and chain required
type InstrumentedHttpProviderOptions = Required<Pick<HttpClientOptions, "url" | "chain">> &
  HttpClientOptions;

/**
 * A simple HTTP client with exponential backoff retries and 429 handling.
 */
export class InstrumentedHttpProvider {
  private timeout: number = 5_000;
  private url: string;
  health: ProviderHealthInstrumentation;

  private logger: winston.Logger = winston.child({ module: "InstrumentedHttpProvider" });

  constructor(options: InstrumentedHttpProviderOptions) {
    options?.timeout && (this.timeout = options.timeout);

    if (!options.url) throw new Error("URL is required");
    this.url = options.url;

    if (!options.chain) throw new Error("Chain is required");

    this.health = new ProviderHealthInstrumentation(this.timeout, options.chain);
  }

  public async post<T>(chain: string, body: any, opts?: HttpClientOptions): Promise<T> {
    return this.execute(chain, "POST", body, opts);
  }

  private async execute<T>(
    chain: string,
    method: string,
    body?: any,
    opts?: HttpClientOptions
  ): Promise<T> {
    let response;
    try {
      response = await this.health.fetch(this.url, {
        method: method,
        body: JSON.stringify(body),
        signal: AbortSignal.timeout(opts?.timeout ?? this.timeout),
        headers: {
          "Content-Type": "application/json",
        },
      });
    } catch (e: AxiosError | any) {
      this.logger.error(
        `[${chain}][${body?.method}] Got error from ${this.url} rpc. ${e?.message ?? `${e}`}`
      );

      // Connection / timeout error:
      if (e instanceof AxiosError) {
        throw new HttpClientError(e.message ?? e.code, { status: e?.status ?? 0 }, e);
      }

      throw new HttpClientError(e.message ?? e.code, undefined, e);
    }

    if (!(response.status > 200) && !(response.status < 300)) {
      throw new HttpClientError(undefined, response, response.json());
    }

    return response.json() as T;
  }
}

export type HttpClientOptions = {
  chain?: string;
  url?: string;
  timeout?: number;
};
