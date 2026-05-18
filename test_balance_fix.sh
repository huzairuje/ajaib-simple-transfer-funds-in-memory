#!/bin/bash

echo "Testing balance calculation fix..."
echo ""

# Create a transfer with initial balances
echo "1. Creating transfer with from_balance=100000, to_balance=50000, amount=1500"
curl -s -X POST http://localhost:3400/v1/transfers \
  -H 'Content-Type: application/json' \
  -d '{
    "from": 1,
    "to": 2,
    "amount": 1500,
    "currency": "IDR",
    "from_balance": 100000,
    "to_balance": 50000
  }' | jq '.'

echo ""
echo "2. Getting transfer list to verify stored balances"
curl -s -X GET http://localhost:3400/v1/transfers | jq '.'

echo ""
echo "Expected from_balance: 98500 (100000 - 1500)"
echo "Expected to_balance: 51500 (50000 + 1500)"
