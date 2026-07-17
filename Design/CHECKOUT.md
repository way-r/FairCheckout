# Checkout

`handler` parses the payload json, passes the request to `service`, and interprets the returned `EventID`.  
`service` is repsonsible for locking the attempt, checking if the address has already been used, sending the `paymentintent` to Stripe, managing the inventory, and asynchronously writing the results to persistent storage.  