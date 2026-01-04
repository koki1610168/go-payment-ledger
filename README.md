# Go Payment Ledger

This repository includes a simple payment ledger written in Go.
The README describes the problems I encounter and the solutions I've implemented  
```! This repository is built for learning.```

## Problems
Building a payment ledger is hard because:  
* Double spending could occur
* Make sure no balance is negative
* Failures can happen at any point of transaction
* Correctness matters more than availability

## Design
Initial Design
* Single process (No need for sync among multitple nodes)
* No idempotency
* Simple HTTP request
---
```
* GET /clients/{clientId}/balance
* POST /payments
```
```
* POST /payments
{
    clientId: ...,
    amount: 1400,
    currency: "JPY"
}
```

