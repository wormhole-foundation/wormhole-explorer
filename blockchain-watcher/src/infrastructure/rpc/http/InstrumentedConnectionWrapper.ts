import { Commitment, Connection, ConnectionConfig } from "@solana/web3.js";
import { ProviderHealthInstrumentation } from "@xlabs/rpc-pool";

export class InstrumentedConnectionWrapper extends Connection {
  health: ProviderHealthInstrumentation;
  private url: string;

  constructor(endpoint: string, commitment: Commitment | ConnectionConfig, timeout: number) {
    super(endpoint, commitment);
    this.health = new ProviderHealthInstrumentation(timeout, "solana");
    this.url = endpoint;
  }

  public setProviderOffline(): void {
    this.health.serviceOfflineSince = new Date();
  }

  public getUrl(): string {
    return this.url;
  }
}
