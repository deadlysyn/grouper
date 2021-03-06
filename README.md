# Contents

- [Overview](#overview)
  - [Architecture](#architecture)
  - [Requirements](#requirements)
  - [Endpoints](#endpoints)
- [Workflow](#workflow)
  - [What groups exist?](#what-groups-exist)
  - [Who do I ask for access?](#who-do-i-ask-for-access)
  - [What groups am I in?](#what-groups-am-i-in)
  - [How do I add group members?](#how-do-i-add-group-members)
- [How Auto-Detection Works](#how-auto-detection-works)
- [Development](#development)
- [Deployment](#deployment)
- [FAQ](#faq)
- [TODO](#todo)

## Overview

![goliath grouper](https://github.com/deadlysyn/grouper/blob/main/assets/grouper.png?raw=true)

"It's kind of a big deal..."

Key tenets:

- You hire smart people you can trust
- Self-service enables autonomy
- Elimination of toil frees humans for more creative work

With that in mind, if you're a member of an autonomous agile team and have
AWS access required to do your job, one conclusion is that you should be able
to grant team members the same access without friction. Grouper lets you do that.

[Related blog post.](https://deadlysyn.com/blog/posts/lean-aws-iam)

### Architecture

Preferred deployment scenario, following the best practice of sandboxing
services in dedicated accounts and using cross-account role assumption.
See [IAM Roles](#iam-roles).

![grouper architecture](https://github.com/deadlysyn/grouper/blob/main/assets/grouper-arch.png?raw=true)

### Requirements

#### Security controls

Grouper is meant to be as lightweight as possible. One decision to accomplish
that is allowing RO endpoints (e.g. get a list of groups) to be accessed
without authentication. While RW endpoints (e.g. updating a group) do require
authentication and authorization (see "Access Keys" below), there is likely no
reason for this service to ever be exposed to the Internet.

Whether you lock it down using security groups, authenticated proxies,
internal ALBs or other means is up to you. Just place it in a trust zone
with adequate protections considering the role it plays.

I've deployed it on a private VPC, behind an internal ALB, with a security
group only allowing access from VPN. You don't have to be that paranoid,
but it doesn't hurt!

#### Access Keys

**IMPORTANT NOTE:** While access keys themselves are not transmitted over
the wire, the key IDs should be considered sensitive. I configure grouper's
ALB to only accept HTTPS traffic (no HTTP redirect) as an extra precaution
to ensure key IDs never transit the network in plaintext.

Calls which modify resources require authorization. To avoid the need for
an external identity source, AWS access keys are used. This means you need
at least one access key associated with your account and configured locally
to be able to modify resources.

The API payload for `PUT`, `POST` or `DELETE` operations contains the caller
ID (user ARN) and access key ID. Since the caller ID could be easily spoofed,
the key ID is used as a "shared secret". You assert who you are with the
user ARN, and grouper verifies that identity by confirming the provided key
ID is actually associated with the specified user ARN.

#### IAM Roles

Grouper acts as an IAM administrator proxy, so you need appropriate roles and
policies granting grouper IAM access. The exact configuration you choose will
depend on context, but here's an example using ECS and cross account role
assumption...

In the account running grouper, ECS has "task" and "exec" roles. Attach the
following policy to your "task" role, along with any other permissions your
tasks need such as reading secrets (`012345678901` is the fictional account
housing your managed IAM resources):

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "sts:AssumeRole",
            "Resource": [
                "arn:aws:iam::012345678901:role/grouper"
            ]
        }
    ]
}
```

This allows grouper tasks to assume the specified role in the IAM account.
Create the `grouper` role in the IAM account. Adjust as needed, but the
simplest approach would attach the `IAMFullAccess` AWS managed policy. Edit
trust relationships and add the grouper account and task role principals
(`09876543210` is the fictional account running your grouper tasks):

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::09876543210:root"
      },
      "Action": "sts:AssumeRole"
    },
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::09876543210:role/your-grouper-task-role"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

#### AWS CLI

[You need to have the AWS CLI installed and configured](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html).
The `groupadd` script wraps `aws-vault` vs using the AWS CLI directly, but you
still need a functional CLI to initially load credentials. AWS has excellent
standalone installers and most OS distributions have packaged options
([AUR](https://aur.archlinux.org/packages/aws-cli-v2-bin), [homebrew](https://formulae.brew.sh/formula/awscli),
etc).

#### `aws-vault`

If you routinely juggle a lot of different accounts and don't like having
plaintext credentials lying around your disk (although I'm sure you have disk
encryption enabled, so this is just about layers of protection at this point!),
you need [aws-vault](https://github.com/99designs/aws-vault#installing).

Technically you don't need the AWS CLI or `aws-vault` installed. You can simply
curl the API! These are requirements of the `groupadd` convenience wrapper.
That said, if you're doing other things which routinely use the AWS CLI, you
should be using `aws-vault`!

#### `curl` and `jq`

While not a requirement for the API, `curl` and `jq` are used by the `groupadd`
script. Who doesn't have these installed anyway? :-)

### Endpoints

All endpoints beside `/healthz` prefixed by `/api/v1`.

Endpoints:

| Method | URI                                | Notes                                    |
|--------|------------------------------------|------------------------------------------|
| GET    | /healthz                           | simple healthcheck                       |
| GET    | /groups                            | return list of groups                    |
| GET    | /groups/:groupname                 | return list of users for specified group |
| GET    | /users/:username                   | return user/group detail for username    |
| PUT    | /groups/:groupname/users/:username | add user to group                        |
| DELETE | /groups/:groupname/users/:username | remove user from group. admins only.     |

`PUT` and `DELETE` payload:

```json
{
  "caller_id": "$AWS_CALLER_ID",
  "key_id": "$AWS_ACCESS_KEY_ID"
}
```

## Workflow

This is a lean microservice meant to help individuals with AWS access navigate
a group-based permissions scheme, and enable group members to self-serve
adding new members. The following scenarios are the most common...

### What groups exist?

I wish all groups neatly conformed to the naming convention (because naming
conventions solve everything!), but organic growth is a reality. This can be
solved with good documentation... and documentation can be generated from
good tooling (self-discovery might even avoid managing more documentation).

```console
??? http https://grouper/api/v1/groups
HTTP/1.1 200 OK
Connection: keep-alive
Content-Length: 581
Content-Type: application/json; charset=utf-8
Date: Sat, 30 Oct 2021 20:56:40 GMT

{
  "message": {
    "groups": [
      {
        "Arn": "arn:aws:iam::012345678901:group/foo",
        "CreateDate": "2019-08-20T18:48:41Z",
        "GroupId": "AGPAWSFUDZ6256EXAMPLE",
        "GroupName": "foo",
        "Path": "/"
      },
      {
        "Arn": "arn:aws:iam::012345678901:group/bar",
        "CreateDate": "2019-08-20T18:48:41Z",
        "GroupId": "AGPAWSFUDZ622BEXAMPLE",
        "GroupName": "bar",
        "Path": "/"
      },
...
```

### Who do I ask for access?

You've been pulled into a new team and even found a matching group, but aren't
completely sure who to ask for access. You could ping your manager, but they're
always in meetings. You could @team but you're really looking for @subteam.
Find group members yourself and start a slack thread with those who can
directly help you:

```console
??? http https://grouper/api/v1/groups/foo
HTTP/1.1 200 OK
Connection: keep-alive
Content-Type: application/json; charset=utf-8
Date: Sun, 16 Jan 2022 23:11:24 GMT
Transfer-Encoding: chunked

{
    "message": {
        "group": {
            "Arn": "arn:aws:iam::012345678901:group/foo",
            "CreateDate": "2016-08-22T13:02:49Z",
            "GroupId": "AGPAJGUTBU62VREXAMPLE",
            "GroupName": "foo",
            "Path": "/"
        },
        "members": [
            {
                "Arn": "arn:aws:iam::012345678901:user/some.user",
                "CreateDate": "2018-06-04T21:49:03Z",
                "PasswordLastUsed": "2022-01-14T16:40:57Z",
                "Path": "/",
                "PermissionsBoundary": null,
                "Tags": null,
                "UserId": "AIDAIDYYHCWNMGEXAMPLE",
                "UserName": "some.user"
            },
            {
                "Arn": "arn:aws:iam::012345678901:user/another.user",
                "CreateDate": "2016-11-15T15:49:50Z",
                "PasswordLastUsed": "2022-01-14T15:30:32Z",
                "Path": "/",
                "PermissionsBoundary": null,
                "Tags": null,
                "UserId": "AIDAIHXNI2F3E5EXAMPLE",
                "UserName": "another.user"
            },
...
```

### What groups am I in?

Some times it's not obvious what groups you belong to... you can get a list
of non-sensitive user information, including a list of all assigned groups:

```console
??? http https://grouper/api/v1/users/some.user
HTTP/1.1 200 OK
Connection: keep-alive
Content-Length: 581
Content-Type: application/json; charset=utf-8
Date: Sat, 30 Oct 2021 20:56:40 GMT

{
    "message": {
        "groups": [
            {
                "Arn": "arn:aws:iam::012345678901:group/foo",
                "CreateDate": "2016-09-12T19:05:59Z",
                "GroupId": "AGPAI5QY6GU5PAEXAMPLE",
                "GroupName": "foo",
                "Path": "/"
            },
            {
                "Arn": "arn:aws:iam::012345678901:group/bar",
                "CreateDate": "2016-08-22T13:02:49Z",
                "GroupId": "AGPAJGUTBU62VZEXAMPLE",
                "GroupName": "bar",
                "Path": "/"
            }
        ],
        "user": {
            "Arn": "arn:aws:iam::012345678901:user/some.user",
            "CreateDate": "2020-01-08T21:06:25Z",
            "PasswordLastUsed": "2021-10-30T20:28:31Z",
            "Path": "/",
            "PermissionsBoundary": null,
            "Tags": null,
            "UserId": "AIDAWSFUDZ626PEXAMPLE",
            "UserName": "some.user"
        }
    },
    "status": 200
}
```

### How do I add group members?

You can only manage groups you are a member of (members of the privileged
`ADMIN_GROUP` can manage all groups). If you need new groups created, custom
policies attached, added to groups team members are not currently part of,
etc. start with a conversation. If it's a common enough use case, submit a PR. :-)

To add new team members to one of your groups, use the `groupadd` helper.
This is only a convenience wrapper, you could just curl the API or build
other options.

```console
??? ./groupadd

USAGE: groupadd -g <IAM_GROUP> -m <IAM_USERNAME> [-c <CALLER_ID> -k <KEY_ID>]

  -g  Friendly name of IAM group to update
  -m  IAM username of group member to add
  -c  User ARN of requester (only needed if auto-detection fails)
  -k  AWS_ACCESS_KEY_ID of requestor (only needed if auto-detection fails)
```

You only need `-g` and `-m`, by default the other arguments are
[auto-detected](#how-auto-detection-works) (requires properly configured AWS
CLI and `aws-vault`):

```console
??? ./groupadd -g testgroup -m some.user
{
  "message": {
    "groups": [
      {
        "Arn": "arn:aws:iam::012345678901:group/foo",
        "CreateDate": "2016-09-12T19:05:59Z",
        "GroupId": "AGPAI5QY6GU5PAEXAMPLE",
        "GroupName": "foo",
        "Path": "/"
      },
      {
        "Arn": "arn:aws:iam::012345678901:group/bar",
        "CreateDate": "2016-08-22T13:02:49Z",
        "GroupId": "AGPAJGUTBU62VZEXAMPLE",
        "GroupName": "bar",
        "Path": "/"
      },
      {
        "Arn": "arn:aws:iam::012345678901:group/testgroup",
        "CreateDate": "2021-10-26T18:30:38Z",
        "GroupId": "AGPAWSFUDZ62ZSEXAMPLE",
        "GroupName": "testgroup",
        "Path": "/"
      }
    ],
    "user": {
      "Arn": "arn:aws:iam::012345678901:user/some.user",
      "CreateDate": "2020-01-08T21:06:25Z",
      "Path": "/",
      "UserId": "AIDAWSFUDZ626PEXAMPLE",
      "UserName": "some.user",
      "PasswordLastUsed": "2021-10-30T20:28:31Z",
      "PermissionsBoundary": null,
      "Tags": null
    }
  },
  "status": 201
}
```

## How Auto-Detection Works

The `groupadd` helper script attempts to auto-detect AWS caller ID and access
key ID to construct the API payload. This requires properly configured
AWS CLI and `aws-vault`.

For AWS CLI, you must have `~/.aws/config` and `~/.aws/credentials` with
entries for your main IAM account (see [architecture diagram](#architecture)):

```console
??? cat ~/.aws/config
[profile main]
aws_account_id=your-iam-account
output=json
region=your-default-region
mfa_serial=arn:aws:iam::012345678901:mfa/first.last

??? cat ~/.aws/credentials
[main]
aws_access_key_id = ...
aws_secret_access_key = ...
```

`aws-vault` should have these credentials added:

```console
??? aws-vault list
Profile                  Credentials              Sessions
=======                  ===========              ========
main                     main                     sts.GetSessionToken:5h6m32s
```

Then `groupadd` can use these for auto-detection:

```console
??? grep PROFILE groupadd
PROFILE=${AWS_PROFILE:-main}
VAULT="aws-vault exec $PROFILE --"
```

If you have a name other than `main` for your IAM account, simply
override `AWS_PROFILE` when running `groupadd`:

```console
??? AWS_PROFILE=foobarbaz ./groupadd -g testgroup -m some.user
```

## Development

Grouper uses cross-account role assumption. The role used for that is
specified in the `ASSUME_ROLE_ARN` environment variable.

When testing locally, "assuming" you are someone with admin access to the IAM
account (likely true if you are working on this service), simply use your
normal `aws-vault` profile and the IAM account admin ARN:

```console
??? ASSUME_ROLE_ARN="arn:aws:iam::012345678901:role/admin" aws-vault exec ops -- go run .
...
```

## Deployment

The provided [Dockerfile](https://github.com/deadlysyn/grouper/blob/main/Dockerfile)
encapsulates all build steps, providing an image which can be pushed to your
registry of choice (ECR, Docker Hub, Quay.io, etc) and ran atop your container
orchestrator of choice (ECS, Kubernetes, etc).

```console
# ECR/ECS example
??? docker build -t grouper:latest .
??? docker tag grouper:latest 012345678901.dkr.ecr.region.amazonaws.com:latest
??? aws ecr get-login-password | docker login --username AWS --password-stdin 012345678901.dkr.ecr.region.amazonaws.com
??? docker push 012345678901.dkr.ecr.region.amazonaws.com/grouper:latest
??? aws ecs update-service --force-new-deployment --cluster your-ecs-cluster --service your-ecs-service
```

Configuration options:

| Environment Variable   | Default Value | Notes                                              |
|------------------------|---------------|----------------------------------------------------|
| ADMIN_GROUP            | ""            | Name of group in IAM account containing admins. Admins can add members to any group and delete members. Required. |
| ASSUME_ROLE_ARN        | ""            | ARN of role used by grouper for cross-account assumption. Required. |
| PORT                   | 8080          | Gin/service listen port. Optional. |
| SLACK_WEBHOOK          | ""            | [Slack webhook](https://api.slack.com/messaging/webhooks) URL. Optional. |
| TRUSTED_PROXIES        | ""            | Space delimited list of v4/v6 IPs or CIDR ranges. See [this](https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies). Optional. |
| TRUSTED_REQUEST_HEADER | ""            | HTTP header name holding client IP. See [this](https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies). Optional. |

## FAQ

`Q:` `credentials missing` errors running `groupadd`

`A:` Check `aws-vault list` to see if the credentials for AWS_PROFILE (`main` by
default) are loaded. If not, try `aws-vault add <profile_name>`. If you change
profile names or credential configuration, you need to `aws-vault remove/add`.

`Q:` `Unable to parse config file: ~/.aws/credentials`

`A:` Make sure you setup `~/.aws/credentials` using `aws configure`.

`Q:` `InvalidClientTokenId` or `The security token included in the request is invalid`.
Manually running `aws iam list-access-keys` also fails.

`A:` Ensure the default policy attached to users allows listing access keys.
If that is true, this is typically a configuration issue.

You don't have to follow this exactly, but here are two example
configurations known to work. First, using `default` and sourcing that as needed:

```console
[default]
output=json
region=your-default-region
mfa_serial=arn:aws:iam::012345678901:mfa/first.last

[profile main]
role_arn=arn:aws:iam::012345678901:role/admin
source_profile=default
mfa_serial=arn:aws:iam::012345678901:mfa/first.last
```

Defining the `main` profile directly (`main` is the default used by `groupadd`):

```console
[profile main]
aws_account_id=your-iam-account
output=json
region=your-default-region
mfa_serial=arn:aws:iam::012345678901:mfa/first.last
```

Important things to remember:

- You need to specify `mfa_serial` for each profile (not just the sourced profile).
- You need matching `~/.aws/credentials` entries

## TODO

- Add tests
