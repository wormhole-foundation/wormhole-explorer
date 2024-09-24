import { Handler, JobDefinition } from "./entities";
import { RunPollingJob } from "./actions";
import { RunPoolRpcs } from "./actions/RunPoolRpcs";

export interface Job {
  getJobDefinitions(): Promise<JobDefinition[]>;
  getPollingJob(jobDef: JobDefinition): RunPollingJob;
  getPoolRpcs(jobsDef: JobDefinition[]): RunPoolRpcs;
  getHandlers(jobDef: JobDefinition): Promise<Handler[]>;
}
