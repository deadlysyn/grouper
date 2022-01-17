# Local Development

Grouper uses cross-account role assumption (allowing a service in one account
to act as an IAM Administrator in another). The role used for that is
specified in the `ASSUME_ROLE_ARN` environment variable.

When testing locally, "assuming" you are someone with admin access to the IAM
account (likely true if you are working on this service), cross-account
assumption is not required. Simply use your normal aws-vault profile and the
IAM account admin ARN:

```console
❯ ASSUME_ROLE_ARN="arn:aws:iam::012345678901:role/admin" aws-vault exec ops -- go run .
...
```
