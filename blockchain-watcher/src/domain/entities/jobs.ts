export class JobDefinition {
  id: string;
  chain: string;
  chainId: number;
  source: {
    action: string;
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
    source: { action: string; config: Record<string, any> },
    handlers: { action: string; target: string; mapper: string; config: Record<string, any> }[]
  ) {
    this.id = id;
    this.chain = chain;
    this.chainId = chainId;
    this.source = source;
    this.handlers = handlers;
  }
}

export type Handler = (items: any[]) => Promise<any>;
