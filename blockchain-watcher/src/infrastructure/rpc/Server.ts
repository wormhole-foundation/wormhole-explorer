import http from "http";
import url from "url";
import { HealthController } from "./HealthController";
import log from "../log";

export class WebServer {
  private server: http.Server;
  private port: number;

  constructor(port: number, healthController: HealthController) {
    this.port = port;
    this.server = http.createServer(async (req, res) => {
      const route = url.parse(req.url ?? "").pathname;

      if (route === "/metrics") {
        // Return all metrics the Prometheus exposition format
        res.setHeader("Content-Type", "text/plain");
        res.end(await healthController.metrics());
      }

      if (route === "/health") {
        res.end("OK");
      }

      res.statusCode = 404;
      res.end();
    });
    this.start();
  }

  start() {
    this.server.listen(this.port, () => {
      log.info(`Server started on port ${this.port}`);
    });
  }

  stop() {
    this.server.close();
  }
}
