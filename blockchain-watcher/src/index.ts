import {
  createHandlers,
  createWatchers,
  getEnvironment,
  initializeEnvironment,
} from "./infrastructure/environment";
import AbstractWatcher from "./infrastructure/watchers/AbstractWatcher";

async function run() {
  initializeEnvironment(process.env.WATCHER_CONFIG_PATH || "../config/local.json");
  const ENVIRONMENT = await getEnvironment();

  //TODO instantiate the persistence module(s)

  //TODO either hand the persistence module to the watcher, or pull necessary config from the persistence module here

  //TODO the event watchers currently instantiate themselves, which isn't ideal. Refactor for next version
  const handlers = createHandlers(ENVIRONMENT);
  const watchers = createWatchers(ENVIRONMENT, handlers);

  await runAllProcesses(watchers);
}

async function runAllProcesses(allWatchers: AbstractWatcher[]) {
  //These are all the raw processes that will run, wrapped to contain their process ID and a top level error handler
  let allProcesses = new Map<number, () => Promise<number>>();
  let processIdCounter = 0;

  //These are all the processes, keyed by their process ID, that we know are not currently running.
  const unstartedProcesses = new Set<number>();

  //Go through all the watchers, wrap their processes, and add them to the unstarted processes set
  for (const watcher of allWatchers) {
    allProcesses.set(
      processIdCounter,
      wrapProcessWithTracker(processIdCounter, watcher.startWebsocketProcessor)
    );
    unstartedProcesses.add(processIdCounter);
    processIdCounter++;

    allProcesses.set(
      processIdCounter,
      wrapProcessWithTracker(processIdCounter, watcher.startQueryProcessor)
    );
    unstartedProcesses.add(processIdCounter);
    processIdCounter++;

    allProcesses.set(
      processIdCounter,
      wrapProcessWithTracker(processIdCounter, watcher.startGapProcessor)
    );
    unstartedProcesses.add(processIdCounter);
    processIdCounter++;
  }

  //If a process ends, reenqueue it into the unstarted processes set
  const reenqueueCallback = (processId: number) => {
    unstartedProcesses.add(processId);
  };

  //Every 5 seconds, try to start any unstarted processes
  while (true) {
    for (const processId of unstartedProcesses) {
      const process = allProcesses.get(processId);
      if (process) {
        //TODO the process ID is a good key but is difficult to track to meaningful information
        console.log(`Starting process ${processId}`);
        unstartedProcesses.delete(processId);
        process()
          .then((processId) => {
            reenqueueCallback(processId);
          })
          .catch((e) => {
            reenqueueCallback(processId);
          });
      } else {
        //should never happen
        console.error(`Process ${processId} not found`);
      }
    }

    await new Promise((resolve) => setTimeout(resolve, 5000));
  }
}

function wrapProcessWithTracker(
  processId: number,
  process: () => Promise<void>
): () => Promise<number> {
  return () => {
    return process()
      .then(() => {
        console.log(`Process ${processId} exited via promise resolution`);
        return processId;
      })
      .catch((e) => {
        console.error(`Process ${processId} exited via promise rejection`);
        console.error(e);
        return processId;
      });
  };
}

//run should never stop, unless an unexpected fatal error occurs
run()
  .then(() => {
    console.log("run() finished");
  })
  .catch((e) => {
    console.error(e);
    console.error("Fatal error caused process to exit");
  });
