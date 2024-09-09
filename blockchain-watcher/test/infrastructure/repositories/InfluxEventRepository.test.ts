import { afterAll, describe, expect, it, jest } from "@jest/globals";
import {
  InfluxConfig,
  InfluxEventRepository,
  InfluxPoint,
} from "../../../src/infrastructure/repositories/target/InfluxEventRepository";
import { InfluxDB, WriteApi } from "@influxdata/influxdb-client";
import { LogFoundEvent } from "../../../src/domain/entities";

let eventRepository: InfluxEventRepository;
let influxClient: InfluxDB;
let influxWriteApi: WriteApi;
let config: InfluxConfig;

describe("InfluxEventRepository", () => {
  afterAll(async () => {
    await influxWriteApi.close();
  });

  describe("InfluxEventRepository", () => {
    it("should not call influx client when no events given", async () => {
      givenInfluxEventRepository();

      const result = await eventRepository.publish([]);

      expect(result).toEqual({ status: "success" });
      expect(influxWriteApi.writePoints).not.toHaveBeenCalled();
    });

    it("should publish", async () => {
      givenInfluxEventRepository();

      const result = await eventRepository.publish([
        {
          chainId: 1,
          address: "0x123456",
          txHash: "0x123",
          blockHeight: 123n,
          blockTime: 0,
          name: "LogMessagePublished",
          attributes: {
            sequence: 1,
          },
          tags: {
            sender: "0x123456",
          },
        },
      ]);

      expect(result).toEqual({ status: "success" });
      expect(influxWriteApi.writePoints).toHaveBeenCalledTimes(1);
    });

    it("should fail to publish unsupported attributes", async () => {
      givenInfluxEventRepository();

      const result = await eventRepository.publish([
        {
          chainId: 1,
          address: "0x123456",
          txHash: "0x123",
          blockHeight: 123n,
          blockTime: 0,
          name: "LogMessagePublished",
          attributes: {
            sequences: { sequence: 1 },
          },
        },
      ]);

      expect(result).toEqual({
        status: "error",
        reason: "Unsupported field type for sequences: object",
      });
      expect(influxWriteApi.writePoints).toHaveBeenCalledTimes(0);
    });
  });

  describe("InfluxPoint", () => {
    it("should skip fields if they are present in the tags", async () => {
      const event: LogFoundEvent<any> = {
        name: "test-event",
        address: "0x123456",
        chainId: 1,
        txHash: "0x123",
        blockHeight: 123n,
        blockTime: 0,
        attributes: {
          attribute1: "value1",
          attribute2: "value2",
          attribute3: 12345,
        },
        tags: {
          attribute2: "tag2",
        },
      };

      const point = InfluxPoint.fromLogFoundEvent(event);

      const attributes = point.getFields();
      expect(attributes).toEqual([
        ["attribute1", "value1"],
        ["attribute3", 12345],
      ]);

      const tags = point.getTags();
      expect(tags).toEqual([["attribute2", "tag2"]]);
    });
  });
});

const givenInfluxEventRepository = () => {
  config = {
    url: "http://localhost:8086",
    token: "my-token",
    org: "my-org",
    bucket: "my-bucket",
  };
  influxWriteApi = {
    writePoints: jest.fn(() => {}),
    close: jest.fn(),
  } as unknown as WriteApi;
  influxClient = {
    getWriteApi: () => influxWriteApi,
  } as unknown as InfluxDB;
  eventRepository = new InfluxEventRepository(influxClient, config);
};
