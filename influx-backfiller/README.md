# influx backfiller

Takes CSV file with VAAs as input, and generates a `line protocol` file for bulk loading into InfluxDB.



## Usage

Run the program to generate InfluxDB dump files:

```bash
./influx-backfiller metrics vaa-count --input vaas-signed.csv --output vaa-count.csv
./influx-backfiller metrics vaa-volume --input vaas-signed.csv --output vaa-volume.csv
```

Then load the files into InfluxDB:

```bash
influx write --bucket wormhole-explorer --file vaa-count.csv
influx write --bucket wormhole-explorer --file vaa-volume.csv
```

## Historic Prices

The prices file is generated with `cmd/symbol_historic` and uses the coingecko api to fetch daily prices for
all supported symbols. 

There is a compressed version of the file already generated and pushed called `prices.csv.gz`




