# influx backfiller


Based on a CSV with VAAs generates a `line protocol` file for bulk loading into influx bucket



## Usage

Run the programa and generate an `output.csv` file

```bash
./influx-backfiller vaas.csv
```

Then load the output to influx

```bash
influx write --bucket wormhole --file output.txt
```

## Line protocol and metrics

The generated file contains a metric with origin and dest chains and addresses.
The origin address is actaully the address of the token ECR-20 or similar.
Also we try to resolve to token symbol based on the address used on the payload. 
If is not possible then we put `none` to the token. 

Example a of generated entry: 

```
vaa,origin_chain=ethereum,target_chain=solana,token=1SOL,origin_address=ethereum,target_address=dadf14a0c31e3ade45b20e5b8e8ac7070235ea00742340a7cc4232a78dac2a3b amount=433719.00000000,notional=1.2341 1680468539000000000
```

The amount is formated using the decimal definition for the token.
The notional value represents the value of the token for that time period. 


## Historic Prices

Prices file is generated with `cmd/symbol_historic` and uses the coingeck api to fetch daily prices for
supported symbols. 

There is a compressed version of the file already generated and pushed called `prices.csv.gz`




