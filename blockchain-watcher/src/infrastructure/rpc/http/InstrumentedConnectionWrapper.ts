import { Commitment, Connection, ConnectionConfig } from "@solana/web3.js";
import { ProviderHealthInstrumentation } from "@xlabs/rpc-pool";

export class InstrumentedConnectionWrapper extends Connection {
  health: ProviderHealthInstrumentation;
  private url: string;

  constructor(
    endpoint: string,
    commitment: Commitment | ConnectionConfig,
    timeout: number,
    chain: string
  ) {
    super(endpoint, commitment);
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

  public getUrl(): string {
    return this.url;
  }

  public isHealthy(): boolean {
    return this.health.isHealthy;
  }
}
