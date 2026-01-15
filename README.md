# Go Payment Ledger

## Action Items
* [ ] add more tests on the handler (e.g. Bad JSON request)
* [ ] Add idempotency key

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

### Database

clients
```
   Column   |           Type           | Collation | Nullable | Default 
------------+--------------------------+-----------+----------+---------
 client_id  | text                     |           | not null | 
 balance    | bigint                   |           | not null | 
 currency   | text                     |           | not null | 
 created_at | timestamp with time zone |           | not null | now()
```
ledger_entries
```
   Column   |           Type           | Collation | Nullable | Default 
------------+--------------------------+-----------+----------+---------
 entry_id   | uuid                     |           | not null | 
 client_id  | text                     |           | not null | 
 amount     | bigint                   |           | not null | 
 created_at | timestamp with time zone |           | not null | now()
```
### HTTP Request

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


