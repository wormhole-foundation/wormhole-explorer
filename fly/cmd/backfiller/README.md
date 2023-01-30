# generic backfiller 

reads a bulk csv dump of stuff (VAAs, TxHash, etc)  and upsert it into mongodb

## compile 

```bash
go build
```

## run 

```bash
./backfiller -strategy vaa -file test.csv
```


When you run the backfiller must pass a valid strategy
Current supported strategies are:
  - `vaa`  for backfilling VAAs
  - `txhash` for backfilling of txHash
  


## config

The mongodb uri is set via env using the variable `MONGODB_URI`
If is not set it will use the default `mongodb://localhost:27017/`



