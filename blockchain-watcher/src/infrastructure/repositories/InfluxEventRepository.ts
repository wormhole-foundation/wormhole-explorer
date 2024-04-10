import { LogFoundEvent } from "../../domain/entities";
import winston from "../log";
import { InfluxDB, Point, convertTimeToNanos } from "@influxdata/influxdb-client";

export class InfluxEventRepository {
  private client: InfluxDB;
  private cfg: InfluxConfig;
  private logger: winston.Logger;

  constructor(client: InfluxDB, cfg: InfluxConfig) {
    this.client = client;
    this.cfg = cfg;
    this.logger = winston.child({ module: "InfluxEventRepository" });
    this.logger.info(`Created for bucket ${cfg.bucket}`);
  }

  async publish(events: LogFoundEvent<any>[]): Promise<InfluxPublishResult> {
    if (!events.length) {
      this.logger.debug("[publish] No events to publish, continuing...");
      return {
        status: "success",
      };
    }

    const timestamps: Record<string, boolean> = {};
    const inputs = events.map(InfluxPoint.fromLogFoundEvent).map((influxPoint) => {
      if (timestamps[influxPoint.timestamp]) {
        // see https://docs.influxdata.com/influxdb/v2/write-data/best-practices/duplicate-points/
        while (timestamps[influxPoint.timestamp]) {
          influxPoint.timestamp = `${BigInt(influxPoint.timestamp) + BigInt(1)}`;
        }
      }
      timestamps[influxPoint.timestamp] = true;

      const point = new Point(influxPoint.measurement).timestamp(influxPoint.timestamp);

      for (const [k, v] of influxPoint.getTags()) {
        point.tag(k, v);
      }

      for (const [k, v] of influxPoint.getFields()) {
        if (typeof v === "number") {
          point.intField(k, v);
        } else if (typeof v === "boolean") {
          point.booleanField(k, v);
        } else {
          point.stringField(k, v);
        }
      }

      return point;
    });

    try {
      const writeApi = this.client.getWriteApi(this.cfg.org, this.cfg.bucket, "ns");
      writeApi.writePoints(inputs);
      await writeApi.close();
    } catch (error: unknown) {
      this.logger.error(`[publish] ${error}`);

      return {
        status: "error",
      };
    }

    this.logger.info(`[publish] Published ${events.length} points to Influx`);
    return {
      status: "success",
    };
  }

  async asTarget(): Promise<(events: LogFoundEvent<any>[]) => Promise<void>> {
    return async (events: LogFoundEvent<any>[]) => {
      const result = await this.publish(events);
      if (result.status === "error") {
        this.logger.error(
          `[asTarget] Error publishing events to Influx: ${result.reason ?? result.reasons}`
        );
        throw new Error(`Error publishing events to Influx: ${result.reason}`);
      }
    };
  }
}

class InfluxPoint {
  constructor(
    public measurement: string,
    public source: string,
    public timestamp: string, // in nanoseconds
    public version: string,
    public fields: Record<string, any>,
    public tags: Record<string, string>
  ) {}

  static fromLogFoundEvent<T extends InfluxPointData>(
    logFoundEvent: LogFoundEvent<T>
  ): InfluxPoint {
    const ts = convertTimeToNanos(new Date(logFoundEvent.blockTime * 1000));
    if (!ts) {
      throw new Error(`Invalid timestamp ${logFoundEvent.blockTime}`);
    }

    return new InfluxPoint(
      logFoundEvent.name,
      "blockchain-watcher",
      ts,
      "1",
      logFoundEvent.attributes.fields,
      logFoundEvent.attributes.tags
    );
  }

  getTags() {
    return Object.entries(this.tags);
  }

  getFields() {
    return Object.entries(this.fields);
  }
}

export type InfluxPointData = {
  tags: Record<string, string>;
  fields: Record<string, any>;
};

export type InfluxConfig = {
  bucket: string;
  org: string;
  token: string;
  url: string;
};

export type InfluxPublishResult = {
  status: "success" | "error";
  reason?: string;
  reasons?: string[];
};
