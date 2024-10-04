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
  private initialDelay: number = 1_000;
  private maxDelay: number = 60_000;
  private retries: number = 0;
  private timeout: number = 5_000;
  private url: string;
  private chain: string;
  health: ProviderHealthInstrumentation;

  private logger: winston.Logger = winston.child({ module: "InstrumentedHttpProvider" });

  constructor(options: InstrumentedHttpProviderOptions) {
    options?.initialDelay && (this.initialDelay = options.initialDelay);
    options?.maxDelay && (this.maxDelay = options.maxDelay);
    options?.retries && (this.retries = options.retries);
    options?.timeout && (this.timeout = options.timeout);

    if (!options.url) throw new Error("URL is required");
    this.url = options.url;

    if (!options.chain) throw new Error("Chain is required");
    this.chain = options.chain;

    this.health = new ProviderHealthInstrumentation(this.timeout, options.chain);
  }

  public async post<T>(body: any, opts?: HttpClientOptions): Promise<T> {
    return this.execute("POST", body, undefined, opts);
  }

  public async get<T>(endpoint: string, params?: any, opts?: HttpClientOptions): Promise<T> {
    const queryParamBuilder = new QueryParamBuilder().addParams(params).build();

    const endpointBuild = `${endpoint}${queryParamBuilder}`;

    return this.execute("GET", undefined, endpointBuild, opts);
  }

  public setProviderOffline(): void {
    this.health.serviceOfflineSince = new Date();
  }

  public getLatency(): number | undefined {
    const durations = this.health.lastRequestDurations;
    return durations.length > 0 ? durations[durations.length - 1] : undefined;
  }

  public isHealthy(): boolean {
    return this.health.isHealthy;
  }

  public getUrl(): string {
    return this.url;
  }

  private async execute<T>(
    method: string,
    body?: any,
    endpoint?: string,
    opts?: HttpClientOptions
  ): Promise<T> {
    let response;
    try {
      const requestOpts: RequestOpts = {
        method,
        signal: AbortSignal.timeout(opts?.timeout ?? this.timeout),
        headers: {
          "Content-Type": "application/json",
        },
      };

      if (method === "POST") {
        requestOpts.body = JSON.stringify(body);
      }

      const url = method === "POST" ? this.url : `${this.url}${endpoint}`;

      response = await this.health.fetch(url, requestOpts);
    } catch (e: AxiosError | any) {
      this.logger.error(
        `[${this.chain}][${body?.method ?? method}] Got error from ${this.url} rpc. ${
          e?.message ?? `${e}`
        }`
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
  initialDelay?: number;
  maxDelay?: number;
  retries?: number;
  timeout?: number;
};

type RequestOpts = {
  method: string;
  signal: AbortSignal;
  headers: {
    "Content-Type": string;
  };
  body?: string;
};

class QueryParamBuilder {
  private queryParams: Map<string, string>;

  constructor() {
    this.queryParams = new Map();
  }

  addParams(params: any): QueryParamBuilder {
    for (const key in params) {
      if (params[key]) {
        this.queryParams.set(key, params[key]);
      }
    }
    return this;
  }

  removeParam(key: string): QueryParamBuilder {
    this.queryParams.delete(key);
    return this;
  }

  build(): string {
    if (this.queryParams.size === 0) {
      return "";
    }

    const queryString = Array.from(this.queryParams.entries())
      .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
      .join("&");
    return `?${queryString}`;
  }
}
