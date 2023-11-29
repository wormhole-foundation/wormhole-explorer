import { Config } from "../config";

export default class Context {
    static _bindings = new Map<string, any>();
     
    constructor () {}

    static setConfig(cfg: Config) {
      Context._bindings.set("cfg", cfg);
    }
      
    static set(key: string, value: any) : void {
      Context._bindings.set(key, value);
    }
      
    static get(key: string) : string | undefined {
      return Context._bindings.get(key) || undefined;
    }
  }