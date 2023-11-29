
export class HttpClientError extends Error {
    public readonly status?: number;
    public readonly data?: any;
    public readonly headers?: any;
  
    constructor(message?: string, response?: { status: number; headers?: any }, data?: any) {
      super(message ?? `Unexpected status code: ${response?.status}`);
      this.status = response?.status;
      this.data = data;
      this.headers = response?.headers;
      Error.captureStackTrace(this, this.constructor);
    }
  
    /**
     * Parses the Retry-After header and returns the value in milliseconds.
     * @param maxDelay
     * @param error
     * @throws {HttpClientError} if retry-after is bigger than maxDelay.
     * @returns the retry-after value in milliseconds.
     */
    public getRetryAfter(maxDelay: number, error: HttpClientError): number | undefined {
      const retryAfter = this.headers?.get("Retry-After");
      if (retryAfter) {
        const value = parseInt(retryAfter) * 1000; // header value is in seconds
        if (value <= maxDelay) {
          return value;
        }
  
        throw error;
      }
    }
  }