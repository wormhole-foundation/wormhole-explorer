import { Handler, JobDefinition } from "./entities";
import { RunPoolRpcs } from "./actions/RunPoolRpcs";
import { RunPollingJob } from "./actions";

export interface Job {
  getJobDefinitions(): Promise<JobDefinition[]>;
  getRunPollingJob(jobDef: JobDefinition): RunPollingJob;
  getHandlers(jobDef: JobDefinition): Promise<Handler[]>;
  getRunPoolRpcs(jobsDef: JobDefinition[]): RunPoolRpcs;
}
