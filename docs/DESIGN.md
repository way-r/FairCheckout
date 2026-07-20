# Design

## Problem Statement
Anti-botting solutions for high-demand physical item releases are currently incomplete. Existing approaches focus heavily on network-level bans (IP blocking) and payment filtering. Both are easily bypassed by resellers utilizing residential proxy networks and generated Virtual Credit Cards (VCCs).

FairCheckout approaches this problem from a completely different angle. By using the physical delivery address as the strict, normalized unique identifier for orders, we create an un-forgeable bottleneck for resellers, rendering mass automated purchases logistically impossible.

### Goals
FairCheckout must provide a frictionless checkout experience for genuine fans while surviving extreme, sudden bursts of traffic from automated programs.
- **Throughput & Latency**: The system can sustain an average of 5,000 checkout requests per second for 10-minute drop windows, maintaining sub-100ms p99 latency.
- **Inventory Guarantees**: The system never oversells under any circumstances. Inventory decrements must be strictly atomic.
- **Jigging Resistance**: The system enforces a "one-per-household" rule using address normalization to detect address manipulation (jigging) in real-time.

### Non-Goals
FairCheckout is strictly a transaction coordinator with inventory reservation capabilities. It has unavoidable and intended functional limitations. 
- **Traffic Scrubbing**: The system cannot actively detect or block bot traffic at the network level. It relies on the merchant to provide standard edge protection (e.g., Cloudflare Turnstile).
- **Absolute Jigging Prevention**: The system does not guarantee a 0% false-positive and 0% false-negative rate for address matching. We use an aggressive algorithm for address normalization to maintain low latency. This introduces tradeoffs such as occasionally missing sophisticated jigs, and intentionally blocking genuine buyers.
- **Payment Gateway Implementation**: The system does not handle payment processing. We are integrating *Stripe* to handle PCI compliance and card authorization.
