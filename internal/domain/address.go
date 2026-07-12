package domain

import (
	"fmt"
	"strings"
)

type ShippingAddress struct {
	Line1   string `json:"line_1" binding:"required"`
	Line2   string `json:"line_2,omitempty"`
	City    string `json:"city" binding:"required"`
	ZipCode string `json:"zip_code" binding:"required"`
}

func (a *ShippingAddress) BaseKey() string {
	rawAddress := fmt.Sprintf("%s %s", a.Line1, a.Line2)
	address := strings.ToLower(rawAddress)
	words := strings.Fields(address)
	cleanAddress := strings.Join(words, "_")
	return fmt.Sprintf("{%s}_%s", a.ZipCode, cleanAddress)
}
