https://developer.paypal.com/docs/api/payments.payouts-batch/v1/#payouts_create
**MODIFY PRIOR IMPLEMENTATION**
This is used for Payouts API to make payments to multiple Paypal/Venmo recipients. The payouts API is a fast, convenient way to send commissions, rebates, rewards, and general disbursements. You can send up to 15,000 payments per call. If interegrated prior 01SEP2017, you receive transaction repots through Mass Payments Reporting. Otherwise view from PayPal Acct on desktop.

curl -v -X POST https://api.sandbox.paypal.com/v1/payments/payouts \
-H "Content-Type: application/json" \
-H "Authorization: Bearer Access-Token" \
-d '{
  "sender_batch_header": {
    "sender_batch_id": "Payouts_2018_100007",
    "email_subject": "You have a payout!",
    "email_message": "You have received a payout! Thanks for using our service!"
  },
  "items": [
    {
      "recipient_type": "EMAIL",
      "amount": {
        "value": "9.87",
        "currency": "USD"
      },
      "note": "Thanks for your patronage!",
      "sender_item_id": "201403140001",
      "receiver": "receiver@example.com",
      "alternate_notification_method": {
        "phone": {
          "country_code": "91",
          "national_number": "9999988888"
        }
      },
      "notification_language": "fr-FR"
    },
    {
      "recipient_type": "PHONE",
      "amount": {
        "value": "112.34",
        "currency": "USD"
      },
      "note": "Thanks for your support!",
      "sender_item_id": "201403140002",
      "receiver": "91-734-234-1234"
    },
    {
      "recipient_type": "PAYPAL_ID",
      "amount": {
        "value": "5.32",
        "currency": "USD"
      },
      "note": "Thanks for your patronage!",
      "sender_item_id": "201403140003",
      "receiver": "G83JXTJ5EHCQ2"
    }
  ]
}'