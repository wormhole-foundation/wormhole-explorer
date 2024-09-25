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

  public getUrl(): string {
    return this.url;
  }

  public isHealthy(): boolean {
    return this.health.isHealthy;
  }
}
