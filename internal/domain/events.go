package domain

import (
	"net/http"
)

type EventID int

const (
	// purchase sucessful, no violation
	PurchaseCompleted = 1000
	// an order with the same address is being processed
	DuplicatedProcessing EventID = 1001
	// an order with the same address has been made
	DuplicatedOrder EventID = 1003
	// payment rejected by processor, no violation
	PaymentDecline EventID = 2001
	// service went wrong
	InternalError EventID = 3001
)

func (e EventID) String() string {
	switch e {
	case PurchaseCompleted:
		return "Purchase completed"
	case DuplicatedProcessing:
		return "Purchase failed due to another order with duplicated address being processed"
	case DuplicatedOrder:
		return "Purchase failed due to another order completed with duplicated address"
	case PaymentDecline:
		return "Purchase failed due to payment processor declining the payment"
	case InternalError:
		return "Purchase failed due to an internal error"
	default:
		return "Unknown status"
	}
}

func (e EventID) StatusCode() int {
	switch e {
	case PurchaseCompleted:
		return http.StatusAccepted
	case DuplicatedProcessing:
		return http.StatusTooManyRequests
	case DuplicatedOrder:
		return http.StatusConflict
	case PaymentDecline:
		return http.StatusPaymentRequired
	case InternalError:
		return http.StatusInternalServerError
	default:
		return 0
	}
}
