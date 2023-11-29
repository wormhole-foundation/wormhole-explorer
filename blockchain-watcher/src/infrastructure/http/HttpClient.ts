import axios, { AxiosError, AxiosInstance } from "axios";
import { setTimeout } from "timers/promises";
import { HttpClientError } from "../errors/HttpClientError";

/**
 * A simple HTTP client with exponential backoff retries and 429 handling.
 */
export class HttpClient {
  private initialDelay: number = 1_000;
  private maxDelay: number = 60_000;
  private retries: number = 0;
  private timeout: number = 5_000;
  private axios: AxiosInstance;

  constructor(options?: HttpClientOptions) {
    options?.initialDelay && (this.initialDelay = options.initialDelay);
    options?.maxDelay && (this.maxDelay = options.maxDelay);
    options?.retries && (this.retries = options.retries);
    options?.timeout && (this.timeout = options.timeout);
    this.axios = axios.create();
  }

  public async post<T>(url: string, body: any, opts?: HttpClientOptions): Promise<T> {
    return this.executeWithRetry(url, "POST", body, opts);
  }

  private async execute<T>(
    url: string,
    method: string,
    body?: any,
    opts?: HttpClientOptions
  ): Promise<T> {
    let response;
    try {
      response = await this.axios.request<T>({
        url: url,
        method: method,
        data: body,
        timeout: opts?.timeout ?? this.timeout,
        signal: AbortSignal.timeout(opts?.timeout ?? this.timeout),
      });
    } catch (err: AxiosError | any) {
      // Connection / timeout error:
      if (err instanceof AxiosError) {
        throw new HttpClientError(err.message ?? err.code, { status: err?.status ?? 0 }, err);
      }

      throw new HttpClientError(err.message ?? err.code, undefined, err);
    }

    if (!(response.status > 200) && !(response.status < 300)) {
      throw new HttpClientError(undefined, response, response.data);
    }

    return response.data;
  }

  private async executeWithRetry<T>(
    url: string,
    method: string,
    body?: any,
    opts?: HttpClientOptions
  ): Promise<T> {
    const maxRetries = opts?.retries ?? this.retries;
    let retries = 0;
    const initialDelay = opts?.initialDelay ?? this.initialDelay;
    const maxDelay = opts?.maxDelay ?? this.maxDelay;
    while (maxRetries >= 0) {
      try {
        return await this.execute(url, method, body, opts);
      } catch (err) {
        if (err instanceof HttpClientError) {
          if (retries < maxRetries) {
            const retryAfter = err.getRetryAfter(maxDelay, err);
            if (retryAfter) {
              await setTimeout(retryAfter, { ref: false });
            } else {
              const timeout = Math.min(initialDelay * 2 ** maxRetries, maxDelay);
              await setTimeout(timeout, { ref: false });
            }
            retries++;
            continue;
          }
        }
        throw err;
      }
    }

    throw new Error(`Failed to reach ${url}`);
  }
}

export type HttpClientOptions = {
  initialDelay?: number;
  maxDelay?: number;
  retries?: number;
  timeout?: number;
};
