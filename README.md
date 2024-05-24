# aws-usage-alerts


## Input
- Payer tag name (value is email address)
- Admin emails for reports/alerts at 100% usage
- Some email credentials
- Usage threshold / month


## Store
- Schema version
- Usage to date
- When last sent an email for each recipient


## Output
- Alters at increments of threshold (50%, 70%, 80%, 90%, 95%, 100%)
- Admin emails when threshold is reached

https://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/using-the-aws-price-list-bulk-api-fetching-price-list-files.html
```
aws pricing list-price-lists --service-code AmazonEFS --currency-code USD --effective-date "2024-04-03 00:00" --region us-east-1
# get arn
aws pricing get-price-list-file-url --price-list-arn arn:aws:pricing:::price-list/aws/AmazonEFS/USD/20240216202306/eu-west-2 --file-format json  --region "us-east-1
# GET URL
```
