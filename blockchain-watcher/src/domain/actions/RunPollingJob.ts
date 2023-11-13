import { setTimeout } from "timers/promises";
import * as log from "winston";

export abstract class RunPollingJob {
  private interval: number;
  private steps: ((items: any[]) => Promise<any>)[];
  private running: boolean = false;

  protected abstract hasNext(): Promise<boolean>;
  protected abstract get(): Promise<any[]>;

  constructor(interval: number, steps: ((items: any[]) => Promise<void>)[]) {
    this.steps = steps;
    this.interval = interval;
    this.running = true;
  }

  public async run(): Promise<void> {
    while (this.running && (await this.hasNext())) {
      const items = await this.get();
      const stepItems: any[][] = [];
      for (const step of this.steps) {
        const intermediateItems = await step(items);
        stepItems.push(intermediateItems);
      }

      await setTimeout(this.interval);
    }

    log.info("Polling job finished");
  }
}
