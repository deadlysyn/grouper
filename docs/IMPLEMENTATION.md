# Implementation Detail

Endpoints:

| Method | URI                      | Notes                                              |
|--------|--------------------------|----------------------------------------------------|
| GET    | /healthz                 | simple healthcheck                                 |
| GET    | /groups                  | return list of IAM groups                          |
| GET    | /groups/:groupname       | return list of users for specified IAM group       |
| GET    | /users/:username         | return user/group detail for username              |
| POST   | /users/:username/groups  | modify groups for username                         |

`/users/:username/groups` payload:

```json
{
  "caller_id": "$AWS_CALLER_ID",
  "group": "$IAM_FRIENDLY_GROUP_NAME",
  "key_id": "$AWS_ACCESS_KEY_ID"
}
```
