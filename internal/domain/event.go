package domain

type EventId int

const (
	PurchaseCompleted EventId = 1000
	OrderProcessing   EventId = 1001
	DuplicatedAddress EventId = 1003
	PaymentDecline    EventId = 2001
	InternalError     EventId = 3001
)

func (e EventId) String() string {
	switch e {
	case PurchaseCompleted:
		return "Purchase completed"
	case OrderProcessing:
		return "Purchase failed due to another order with duplicated address being processed"
	case DuplicatedAddress:
		return "Purchase failed due to another order completed with duplicated address"
	case PaymentDecline:
		return "Purchase failed due to payment processor declining the payment"
	case InternalError:
		return "Purchase failed due to an internal error"
	default:
		return "Unknown status"
	}
}
