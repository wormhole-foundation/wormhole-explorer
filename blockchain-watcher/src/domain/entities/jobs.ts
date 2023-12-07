export class JobDefinition {
  id: string;
  name?: string;
  chain: string;
  paused?: boolean = false;
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
    source: { action: string; config: Record<string, any> },
    handlers: { action: string; target: string; mapper: string; config: Record<string, any> }[],
    name?: string
  ) {
    this.id = id;
    this.name = name ?? id;
    this.chain = chain;
    this.source = source;
    this.handlers = handlers;
  }
}

export type Handler = (items: any[]) => Promise<any>;

export interface Runnable {
  run(handlers: Handler[]): Promise<void>;
  stop(): Promise<void>;
}

export type JobExecution = {
  id: string;
  job: JobDefinition;
  status: string;
  error?: Error;
  startedAt: Date;
  finishedAt?: Date;
};
