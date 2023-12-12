import { CronJob } from "cron";
import winston from "winston";

export class RunCronTask {
  private id: string;
  private schedule: CronJob;
  private logger: winston.Logger = winston.child({ module: "RunCronTask" });

  constructor(id: string, expression: string, task: () => Promise<any>) {
    this.id = id;
    this.schedule = CronJob.from({
      cronTime: expression,
      onTick: task,
      start: false,
    });
  }

  public async run(): Promise<void> {
    this.logger.info(`[run] Starting cron task ${this.id}`);
    this.schedule.start();
  }

  public async stop(): Promise<void> {
    this.schedule.stop();
    this.logger.info(`[run] Stopping cron task ${this.id}`);
  }
}
