import { describe, it, expect, beforeEach } from "@jest/globals";
import { StartJobs } from "../../../../src/domain/actions/jobs/StartJobs";
import { JobDefinition, JobExecution } from "../../../../src/domain/entities";
import { JobExecutionRepository, JobRepository } from "../../../../src/domain/repositories";

let startJobs: StartJobs | undefined;
let jobRepository: JobRepository;
let jobExecutionRepository: JobExecutionRepository;

describe("StartJobs", () => {
  beforeEach(() => {
    startJobs = undefined;
  });

  describe("run", () => {
    it("should run a single job and return the job execution", async () => {
      const job = createJobDefinitionExample();

      givenJobsPresent([job]);
      givenNoJobsExecutionPresent([job]);

      const jobExecutions = await whenStartJobsIsCalled();

      expect(jobExecutions).toBeDefined();
      expect(jobExecutions).toHaveLength(1);
    });

    it("should run a job iff no other execution is present", async () => {
      const job = createJobDefinitionExample();

      givenJobsPresent([job]);
      givenNoJobsExecutionPresent([job]);
      const jobExecutions = await whenStartJobsIsCalled();
      expect(jobExecutions).toBeDefined();
      expect(jobExecutions).toHaveLength(1);

      givenJobExecutionsPresent([job]);
      givenStartJobsAction();
      const nextJobExecutions = await whenStartJobsIsCalled();
      expect(nextJobExecutions).toBeDefined();
      expect(nextJobExecutions).toHaveLength(0);
    });

    it("should run jobs with no current execs and ignore the ones running or paused", async () => {
      const jobs = ["job-1", "job-2", "job-3", "job-4"].map(createJobDefinitionExample);
      const [firstJob, secondJob, thirdJob] = jobs;

      thirdJob.paused = true;
      givenJobsPresent(jobs);
      givenNoJobsExecutionPresent([firstJob, secondJob]);

      const jobExecutions = await whenStartJobsIsCalled();
      expect(jobExecutions).toHaveLength(2);
    });

    it("should stop paused jobs if running", async () => {
      const job = createJobDefinitionExample();
      givenJobsPresent([job]);
      givenNoJobsExecutionPresent([job]);
      givenStartJobsAction();

      let jobExecutions = await whenStartJobsIsCalled();
      expect(jobExecutions).toHaveLength(1);

      job.paused = true;
      givenJobsPresent([job]);
      givenJobExecutionsPresent([job]);

      jobExecutions = await whenStartJobsIsCalled();

      expect(jobExecutions).toHaveLength(0);
    });

    it("should stop failing jobs", async () => {
      const job = createJobDefinitionExample();
      givenJobsPresent([job], false);
      givenNoJobsExecutionPresent([job]);
      givenStartJobsAction();

      let jobExecutions = await whenStartJobsIsCalled();
      expect(jobExecutions).toHaveLength(1);

      jobExecutions = await whenStartJobsIsCalled();
      // Should be present again, as it has
      expect(jobExecutions).toHaveLength(1);
    });
  });
});

const givenJobsPresent = (jobs: JobDefinition[], runWorks: boolean = true) => {
  jobRepository = {
    getJobs: () => Promise.resolve(jobs),
    getRunnableJob: () => {
      const runnable = {
        run: () => (runWorks ? Promise.resolve() : Promise.reject(new Error("Error running job"))),
        stop: () => Promise.resolve(),
      };
      return runnable;
    },
    getHandlers: () => Promise.resolve([() => Promise.resolve()]),
  };
};

const createJobDefinitionExample = (id: string = "job-1") => {
  return {
    id,
    name: "Test Job" + id,
    chain: "ethereum",
    source: {
      action: "test",
      config: {},
    },
    handlers: [
      {
        action: "test",
        target: "dummy",
        mapper: "test",
        config: {},
      },
    ],
  } as JobDefinition;
};

const givenJobExecutionsPresent = (jobs: JobDefinition[]) => {
  jobExecutionRepository = {
    start: (job: JobDefinition) => {
      if (jobs.includes(job)) {
        return Promise.reject(new Error("Job already running"));
      }
      return Promise.resolve({ id: job.id, job, status: "running", startedAt: new Date() });
    },
    stop: (jobExec: JobExecution) => {
      if (jobs.includes(jobExec.job)) {
        return Promise.reject(new Error("Job not running"));
      }
      return Promise.resolve({
        ...jobExec,
        status: "stopped",
        startedAt: new Date(),
        finishedAt: new Date(),
      });
    },
  };
};

const givenNoJobsExecutionPresent = (jobs: JobDefinition[]) => {
  jobExecutionRepository = {
    start: (job: JobDefinition) => {
      if (jobs.includes(job)) {
        return Promise.resolve({ id: job.id, job, status: "running", startedAt: new Date() });
      }
      return Promise.reject(new Error("Job already running"));
    },
    stop: (jobExec: JobExecution, error?: Error) => {
      if (jobs.includes(jobExec.job)) {
        return Promise.resolve({
          ...jobExec,
          status: "stopped",
          startedAt: new Date(),
          finishedAt: new Date(),
        });
      }
      return Promise.reject(new Error("Job not running"));
    },
  };
};

const givenStartJobsAction = () => {
  startJobs = new StartJobs(jobRepository, jobExecutionRepository);
};

const whenStartJobsIsCalled = () => {
  if (!startJobs) {
    givenStartJobsAction();
  }
  return startJobs?.run();
};
