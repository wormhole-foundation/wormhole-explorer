import { Handler, JobDefinition } from "./entities";
import { RunPollingJob } from "./actions";
import { RunRPCHealthcheck } from "./actions/RunRPCHealthcheck";

export interface Job {
  getJobDefinitions(): Promise<JobDefinition[]>;
  getRPCHealthcheck(jobsDef: JobDefinition[]): RunRPCHealthcheck;
  getPollingJob(jobDef: JobDefinition): RunPollingJob;
  getHandlers(jobDef: JobDefinition): Promise<Handler[]>;
}
