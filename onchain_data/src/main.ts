import getCustodyData from "./getCustodyData";
import { open } from "fs/promises";
let retryTimeout = 5 * 60 * 1000;
if (process.env.RETRY_TIMEOUT) {
  try {
    retryTimeout = parseInt(process.env.RETRY_TIMEOUT);
  } catch (e) {
    console.log(
      `could not parseInt ${process.env.RETRY_TIMEOUT}. Using default timeout=${retryTimeout}`
    );
  }
}
const filename = "/app/ready";
let firstRun = true;

async function main() {
  while (true) {
    console.log(`${new Date().toISOString()} - fetching custody data`);
    await getCustodyData();
    if (firstRun) {
      let fh = await open(filename, "a");
      await fh.close();
      firstRun = false;
    }
    console.log(`${new Date().toISOString()} - sleeping for ${retryTimeout}`);
    await new Promise((resolve) => setTimeout(resolve, Number(retryTimeout)));
  }
}

main();
