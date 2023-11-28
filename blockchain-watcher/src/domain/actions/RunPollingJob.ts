import { setTimeout } from "timers/promises";
import winston from "winston";
import { Handler } from "../entities";

export abstract class RunPollingJob {
  private interval: number;
  private running: boolean = false;
  protected abstract logger: winston.Logger;
  protected abstract preHook(): Promise<void>;
  protected abstract hasNext(): Promise<boolean>;
  protected abstract get(): Promise<any[]>;
  protected abstract persist(): Promise<void>;

  constructor(interval: number) {
    this.interval = interval;
    this.running = true;
  }

  public async run(handlers: Handler[]): Promise<void> {
    this.logger.info("Starting polling job");
    await this.preHook();
    while (this.running) {
      if (!(await this.hasNext())) {
        this.logger.info("Finished processing");
        await this.stop();
        break;
      }

      let items: any[];

      try {
        items = await this.get();
        await Promise.all(handlers.map((handler) => handler(items)));
      } catch (e: Error | any) {
        this.logger.error("Error processing items", e.stack ?? e);
        await setTimeout(this.interval);
        continue;
      }

      await this.persist();
      await setTimeout(this.interval);
    }
  }

  public async stop(): Promise<void> {
    this.running = false;
  }
}
