# Event Id
| Id   | Event                | Meaning                                           | Http Response |
| ---- | -------------------- | ------------------------------------------------- | ------------- |
| 1000 | PurchaseCompleted    | purchase sucessful, no violation                  | 200           |
| 1001 | DuplicatedProcessing | an order with the same address is being processed | 429           |
| 1003 | DuplicatedOrder      | an order with the same address has been made      | 409           |
| 2001 | PaymentDecline       | payment rejected by processor, no violation       | 402           |
| 3001 | InternalError        | service went wrong                                | 500           |