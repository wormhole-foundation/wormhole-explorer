package parser

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	sdk "github.com/wormhole-foundation/wormhole/sdk/vaa"
)

type vaaProcessor func(vaa *sdk.VAA) error

type VaaCsvParser struct {
	processor vaaProcessor
	filename  string
}

func NewVaaCsvParser(processor vaaProcessor, filename string) *VaaCsvParser {
	return &VaaCsvParser{
		processor: processor,
		filename:  filename,
	}
}

func (p *VaaCsvParser) Start(_ context.Context) {
	c := 0
	i := 0

	f, err := os.Open(p.filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	// read file line by line and send to workpool
	for lineNumber := uint(0); ; lineNumber++ {
		line, _, err := r.ReadLine() //loading chunk into buffer
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("a real error happened in line [%d]. %v\n", lineNumber, err)
		}
		err = p.processLine(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "processing line number [%d] failed: %v\n", lineNumber, err)
		} else {
			if c == 10000 {
				fmt.Printf(".")
				c = 0
				i := i + 1
				if i == 10 {
					fmt.Printf("\n")
					i = 0
				}
			}
		}
	}
}

func (p *VaaCsvParser) processLine(line []byte) error {
	tt := strings.Split(string(line), ",")

	if len(tt) != 2 {
		return fmt.Errorf("invalid line: %s", line)
	}

	data, err := hex.DecodeString(tt[1])
	if err != nil {
		return fmt.Errorf("error decoding: %v", err)
	}

	vaa, err := sdk.Unmarshal(data)
	if err != nil {
		return fmt.Errorf("error unmarshaling vaa: %v", err)
	}

	return p.processor(vaa)
}
