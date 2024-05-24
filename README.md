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
