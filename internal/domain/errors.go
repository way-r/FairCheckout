package domain

import (
	"errors"
)

var ErrLockBusy = errors.New("Lock is currently held by another transaction")
var ErrDupOrder = errors.New("An order has been made to the address")
