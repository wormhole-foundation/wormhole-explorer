import { describe, expect, it, jest } from "@jest/globals";
import {
  SnsEventRepository,
  SnsConfig,
} from "../../../src/infrastructure/repositories/SnsEventRepository";
import { SNSClient } from "@aws-sdk/client-sns";

let snsEventRepository: SnsEventRepository;
let snsClient: SNSClient;
let snsConfig: SnsConfig;

describe("SnsEventRepository", () => {
  it("should publish events", async () => {
    givenSnsEventRepository();

    const result = await snsEventRepository.publish([]);

    expect(result).toEqual({ status: "success" });
  });
});

const givenSnsEventRepository = () => {
  snsConfig = {
    region: "us-east-1",
    topicArn: "arn:aws:sns:us-east-1:123456789012:MyTopic",
    groupId: "groupId",
    subject: "subject",
  };
  snsClient = {
    send: jest.fn().mockReturnThis(),
  } as unknown as SNSClient;
  snsEventRepository = new SnsEventRepository(snsClient, snsConfig);
};
