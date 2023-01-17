# Parser

## Config SQS FIFO with dead letter queue localstack

aws --profile localstack --endpoint-url=http://localhost:4566 sqs create-queue --queue-name=wormhole-vaa-parser-dlq-queue.fifo --attributes "FifoQueue=true"

aws --profile localstack --endpoint-url=http://localhost:4566 sqs create-queue --queue-name=wormhole-vaa-parser-queue.fifo --attributes FifoQueue=true,MessageRetentionPeriod=3600,ReceiveMessageWaitTimeSeconds=5,VisibilityTimeout=20,RedrivePolicy="\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:us-east-1:000000000000:wormhole-vaa-parser-dlq-queue.fifo\\\",\\\"maxReceiveCount\\\":\\\"2\\\"}\""

## Check message in the dead letter queue localstack

aws --profile localstack --endpoint-url=http://localhost:4566 sqs receive-message --queue-url=http://localhost:4566/000000000000/wormhole-vaa-parser-dlq-queue.fifo