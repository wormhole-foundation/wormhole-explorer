import { describe, expect, it, jest } from "@jest/globals";
import { SnsEventRepository, SnsConfig } from "../../../src/infrastructure/repositories";
import { SNSClient } from "@aws-sdk/client-sns";

let snsEventRepository: SnsEventRepository;
let snsClient: SNSClient;
let snsConfig: SnsConfig;

describe("SnsEventRepository", () => {
  it("should not call sns client when no events given", async () => {
    givenSnsEventRepository();

    const result = await snsEventRepository.publish([]);

    expect(result).toEqual({ status: "success" });
    expect(snsClient.send).not.toHaveBeenCalled();
  });

  it("should publish", async () => {
    givenSnsEventRepository();

    const result = await snsEventRepository.publish([
      {
        chainId: 1,
        address: "0x123456",
        txHash: "0x123",
        blockHeight: 123n,
        blockTime: 0,
        name: "LogMessagePublished",
        attributes: {},
      },
    ]);

    expect(result).toEqual({ status: "success" });
    expect(snsClient.send).toHaveBeenCalledTimes(1);
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
