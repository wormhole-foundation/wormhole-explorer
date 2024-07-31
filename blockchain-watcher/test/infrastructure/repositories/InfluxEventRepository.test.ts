import { afterAll, describe, expect, it, jest } from "@jest/globals";
import {
  InfluxConfig,
  InfluxEventRepository,
} from "../../../src/infrastructure/repositories/target/InfluxEventRepository";
import { InfluxDB, WriteApi } from "@influxdata/influxdb-client";

let eventRepository: InfluxEventRepository;
let influxClient: InfluxDB;
let influxWriteApi: WriteApi;
let config: InfluxConfig;

describe("InfluxEventRepository", () => {
  afterAll(async () => {
    await influxWriteApi.close();
  });

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
