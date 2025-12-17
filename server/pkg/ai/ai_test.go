package ai

import (
	"context"
	"fmt"
	"testing"
)

func TestNameAiFIBACountryV2(t *testing.T) {
	contry, err := CountryList(context.Background())
	fmt.Println(contry, err)
}
