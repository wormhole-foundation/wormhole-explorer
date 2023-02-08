# Pipeline

## Config SNS FIFO in localstack

aws --profile localstack --endpoint-url=http://localhost:4566 sns --name vaas-pipeline.fifo  --attributes FifoTopic=true,ContentBasedDeduplication=false
