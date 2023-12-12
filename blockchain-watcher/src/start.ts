import { configuration } from "./infrastructure/config";
import { RepositoriesBuilder } from "./infrastructure/repositories/RepositoriesBuilder";
import log from "./infrastructure/log";
import { WebServer } from "./infrastructure/rpc/http/Server";
import { HealthController } from "./infrastructure/rpc/http/HealthController";
import { StartJobs, RunCronTask } from "./domain/actions";

let repos: RepositoriesBuilder;
let server: WebServer;

async function run(): Promise<void> {
  log.info(`Starting: dryRunEnabled -> ${configuration.dryRun}`);

  repos = new RepositoriesBuilder(configuration);
  await repos.init();

  await startServer(repos);

  const startJobs = new StartJobs(repos.getJobsRepository(), repos.getJobExecutionRepository(), {
    maxConcurrentJobs: configuration.jobs.maxConcurrency,
  });

  if (configuration.jobs.pollJobsCron) {
    const cronTask = new RunCronTask("pollJobs", configuration.jobs.pollJobsCron, () =>
      startJobs.run()
    );
    await cronTask.run();
  } else {
    await startJobs.run();
  }

  log.info("Started");

  // Handle shutdown
  process.on("SIGINT", handleShutdown);
  process.on("SIGTERM", handleShutdown);
}

const startServer = async (repos: RepositoriesBuilder) => {
  server = new WebServer(configuration.port, new HealthController(repos.getStatsRepository()));
};

const handleShutdown = async () => {
  try {
    await Promise.allSettled([repos.close(), server.stop()]);

    process.exit();
  } catch (error: unknown) {
    process.exit(1);
  }
};

run().catch((e) => {
  log.error("Fatal error caused process to exit", e);
});
