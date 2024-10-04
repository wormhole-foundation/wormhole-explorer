import { InstrumentedSuiClient, ProviderHealthInstrumentation } from "@xlabs/rpc-pool";

export class InstrumentedSuiClientWrapper extends InstrumentedSuiClient {
  health: ProviderHealthInstrumentation;
  url: string;

  constructor(endpoint: string, timeout: number, chain: string = "sui") {
    super(endpoint, timeout);
    this.health = new ProviderHealthInstrumentation(timeout, chain);
    this.url = endpoint;
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
}
