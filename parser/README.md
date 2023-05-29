# Parser

This component is in charge of parsing the VAA payload and persists it.

VAA parsing is delegated to external service so that users can add custom parsers in the external service without affecting this service.

## Usage

### Service

```bash
parser service
```

### Backfiller
```bash
parser backfiller [flags]
```

#### Command-line arguments
- **--end-time** *string*                  maximum VAA timestamp to process (default now)
- **--log-level** *string*                 log level (default "INFO")
- **--mongo-database** *string*            mongo database
- **--mongo-uri** *string*                 mongo connection
- **--page-size** *int*                    VAA payload parser timeout (default 100)
- **--start-time** *string*                minimum VAA timestamp to process (default "1970-01-01T00:00:00Z")
- **--vaa-payload-parser-timeout** *int*   maximum waiting time in call to VAA payload service in second (default 10)
- **--vaa-payload-parser-url** *string*    VAA payload parser service URL


## Running parser as service with localstack

Here are some aws commands to configure localstack with the necessary resources

### Config sns topic

```bash
aws --profile localstack --endpoint-url=http://localhost:4566 sns create-topic --name vaas-pipeline.fifo  --attributes FifoTopic=true,ContentBasedDeduplication=false
```
### Config SQS FIFO with dead letter queue localstack

```bash
aws --profile localstack --endpoint-url=http://localhost:4566 sqs create-queue --queue-name=wormhole-vaa-parser-dlq-queue.fifo --attributes "FifoQueue=true"
```

```bash
aws --profile localstack --endpoint-url=http://localhost:4566 sqs create-queue --queue-name=wormhole-vaa-parser-queue.fifo --attributes FifoQueue=true,MessageRetentionPeriod=3600,ReceiveMessageWaitTimeSeconds=5,VisibilityTimeout=20,RedrivePolicy="\"{\\\"deadLetterTargetArn\\\":\\\"arn:aws:sqs:us-east-1:000000000000:wormhole-vaa-parser-dlq-queue.fifo\\\",\\\"maxReceiveCount\\\":\\\"2\\\"}\""
```

### Subscribe SQS FIFO to vaas-pipeline.fifo topic

```bash
aws --profile localstack --endpoint-url=http://localhost:4566 sns subscribe --topic-arn arn:aws:sns:us-east-1:000000000000:vaas-pipeline.fifo --protocol sqs --notification-endpoint http://localhost:4566/000000000000/wormhole-vaa-parser-queue.fifo
```

### Check message in the dead letter queue localstack

```bash
aws --profile localstack --endpoint-url=http://localhost:4566 sqs receive-message --queue-url=http://localhost:4566/000000000000/wormhole-vaa-parser-dlq-queue.fifo
```
