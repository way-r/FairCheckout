package domain

import (
	"net/http"
)

type EventID int

const (
	PurchaseCompleted    EventID = 1000 // purchase sucessful, no violation
	DuplicatedProcessing EventID = 1001 // an order with the same address is being processed
	DuplicatedOrder      EventID = 1003 // an order with the same address has been made
	OutOfStock           EventID = 1005 // item requested is out of stock
	PaymentDecline       EventID = 2001 // payment rejected by processor, no violation
	InternalError        EventID = 3001 // service went wrong
)

func (e EventID) String() string {
	switch e {
	case PurchaseCompleted:
		return "Purchase completed"
	case DuplicatedProcessing:
		return "Purchase failed due to another order with duplicated address being processed"
	case DuplicatedOrder:
		return "Purchase failed due to another order completed with duplicated address"
	case OutOfStock:
		return "Purchase failed due to item being out of stock"
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
	case OutOfStock:
		return http.StatusConflict
	case PaymentDecline:
		return http.StatusPaymentRequired
	case InternalError:
		return http.StatusInternalServerError
	default:
		return 0
	}
}
