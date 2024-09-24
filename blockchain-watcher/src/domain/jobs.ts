import { Handler, JobDefinition } from "./entities";
import { RunPollingJob } from "./actions";
import { RunPoolRpcs } from "./actions/RunPoolRpcs";

export interface Job {
  getJobDefinitions(): Promise<JobDefinition[]>;
  getRunPollingJob(jobDef: JobDefinition): RunPollingJob;
  getRunPoolRpcs(jobsDef: JobDefinition[]): RunPoolRpcs;
  getHandlers(jobDef: JobDefinition): Promise<Handler[]>;
}
