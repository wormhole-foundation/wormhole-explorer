export class JobDefinition {
  id: string;
  chain: string;
  chainId: number;
  source: {
    action: string;
    records: string;
    config: Record<string, any>;
  };
  handlers: {
    action: string;
    target: string;
    mapper: string;
    config: Record<string, any>;
  }[];

  constructor(
    id: string,
    chain: string,
    chainId: number,
    source: { action: string; records: string; config: Record<string, any> },
    handlers: { action: string; target: string; mapper: string; config: Record<string, any> }[]
  ) {
    this.id = id;
    this.chain = chain;
    this.source = source;
    this.chainId = chainId;
    this.handlers = handlers;
  }
}

export type Handler = (items: any[]) => Promise<any>;
