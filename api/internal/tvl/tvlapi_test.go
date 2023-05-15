package tvlapi

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/test-go/testify/assert"
	"github.com/tidwall/gjson"
)

func TestTvlAPI_GetNotionalUSD(t *testing.T) {

	// open and read json file
	jsonFile, err := os.Open("tvl_data.json")
	if err != nil {
		t.Fatal(err)
	}
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		t.Fatal(err)
	}

	rr := gjson.Get(string(byteValue), "AllTime.\\*.\\*.Notional")
	//	tvl := rr.Float()
	fmt.Println(rr.String())

	assert.Equal(t, "329194177.19779253", rr.String())

}
