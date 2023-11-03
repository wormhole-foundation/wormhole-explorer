
export class StreamingSource<Output, Cfg> implements Source<Output, Cfg> {
    getConfiguration(): Cfg {
        throw new Error("Method not implemented.");
    }
    getLastOutput(): Output[] {
        throw new Error("Method not implemented.");
    }
}

export class PollingSource<Output, Cfg> implements Source<Output, Cfg> {
    getConfiguration(): Cfg {
        throw new Error("Method not implemented.");
    }
    getLastOutput(): Output[] {
        throw new Error("Method not implemented.");
    }
}

export abstract class Source<Output, Cfg> {
    getConfiguration(): Cfg {
        throw new Error("Method not implemented.");
    }
   
    abstract getLastOutput(): Output[];
}

export interface Handler<Input, Output> {
    handle(input: Input[]): Promise<Output>;
}

export class Job {
    private name: string;
    private source: Source<any, any>;
    private handlers: Handler<any, any>[];

    constructor(name: string, source: Source<any, any>, handlers: Handler<any, any>[]) {
        this.name = name;
        this.source = source;
        this.handlers = handlers;
    }

    validate(): boolean {
        return true;
    }

}