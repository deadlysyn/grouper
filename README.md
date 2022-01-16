# Contents

- [Overview](#overview)
  - [Architecture](#architecture)
  - [Prerequisites](#prerequisites)
- [Workflow](#workflow)
- [How Auto-Detection Works](#how-auto-detection-works)
- [Implementation Detail](https://github.com/deadlysyn/grouper/blob/main/docs/IMPLEMENTATION.md)
- [Local Development](https://github.com/deadlysyn/grouper/blob/main/docs/DEV.md)
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

TODO: link to blog

### Architecture

Preferred deployment scenario, following the best practice of sandboxing
services in dedicated accounts and using cross-account role assumption.

![grouper architecture](https://github.com/deadlysyn/grouper/blob/main/assets/grouper-arch.png?raw=true)

### Prerequisites

While all are not specific to grouper or strictly required, keep the following
in mind when deploying grouper:

TODO: describe why these are needed...

#### Security controls

VPN access (grouper is only available internally)

#### AWS CLI

https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html)

#### aws-vault

https://github.com/99designs/aws-vault#installing

#### `curl` and `jq` utilities installed

used by groupadd

## Workflow

TODO: add groups endpoints examples

Some times it's not obvious what groups you belong to... you can get a list
of non-sensitive user information, including a list of all assigned groups:

```console
❯ http https://grouper/users/some.user
HTTP/1.1 200 OK
Connection: keep-alive
Content-Length: 581
Content-Type: application/json; charset=utf-8
Date: Sat, 30 Oct 2021 20:56:40 GMT

{
    "message": {
        "groups": [
            {
                "Arn": "arn:aws:iam::012345678901:group/user",
                "CreateDate": "2016-09-12T19:05:59Z",
                "GroupId": "AGPAI5QY6GU5PAEXAMPLE",
                "GroupName": "user",
                "Path": "/"
            },
            {
                "Arn": "arn:aws:iam::012345678901:group/admin",
                "CreateDate": "2016-08-22T13:02:49Z",
                "GroupId": "AGPAJGUTBU62VZEXAMPLE",
                "GroupName": "admin",
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

**NOTE:** You can only manage groups you are a member of (members of the
privileged `admin` group can manage all groups). Open an ITHELP ticket if you
need new groups created (or new policies attached), added to groups team
members are not currently part of, or group members removed.

To add new team members to one of your groups, use the `groupadd` helper:

```console
❯ ./groupadd

USAGE: groupadd -g <IAM_GROUP> -m <IAM_USERNAME> [-c <CALLER_ID> -k <KEY_ID>]

  -g  Friendly name of IAM group to update
  -m  IAM username of group member to add
  -c  User ARN of requester (only needed if auto-detection fails)
  -k  AWS_ACCESS_KEY_ID of requestor (only needed if auto-detection fails)
```

You only need `-g` and `-m`, by default the other arguments are
auto-detected (requires properly configured AWS CLI and `aws-vault`):

```console
❯ ./groupadd -g testgroup -m some.user
{
  "message": {
    "groups": [
      {
        "Arn": "arn:aws:iam::012345678901:group/user",
        "CreateDate": "2016-09-12T19:05:59Z",
        "GroupId": "AGPAI5QY6GU5PAEXAMPLE",
        "GroupName": "user",
        "Path": "/"
      },
      {
        "Arn": "arn:aws:iam::012345678901:group/admin",
        "CreateDate": "2016-08-22T13:02:49Z",
        "GroupId": "AGPAJGUTBU62VZEXAMPLE",
        "GroupName": "admin",
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

If auto-detection fails, see [How Auto-Detection Works](#how-auto-detection-works).
The additional options are there if you need to override. Look at
`groupadd`, note it is just curling a REST API. Endpoints are documented in
[Implementation Detail](https://github.com/deadlysyn/grouper/blob/main/docs/IMPLEMENTATION.md)

# How Auto-Detection Works

The `groupadd` helper script attempts to auto-detect AWS caller ID and access
key ID to construct the API payload. This requires properly configured
AWS CLI and `aws-vault`.

For AWS CLI, you must have `~/.aws/config` and `~/.aws/credentials` with
entries for your main IAM account (see [architecture diagram](#architecture)]):

```console
❯ cat ~/.aws/config
[profile main]
aws_account_id=your-iam-account
output=json
region=your-default-region
mfa_serial=arn:aws:iam::012345678901:mfa/first.last

❯ cat ~/.aws/credentials
[main]
aws_access_key_id = ...
aws_secret_access_key = ...
```

`aws-vault` should have these credentials added:

```console
❯ aws-vault list
Profile                  Credentials              Sessions
=======                  ===========              ========
main                     main                      sts.GetSessionToken:5h6m32s
```

Then `groupadd` can use these for auto-detection:

```console
❯ grep PROFILE groupadd
PROFILE=${AWS_PROFILE:-main}
VAULT="aws-vault exec $PROFILE --"
```

If you have a name other than `main` for your IAM account, simply
override `AWS_PROFILE` when running `groupadd`:

```console
❯ AWS_PROFILE=foobarbaz ./groupadd -g testgroup -m some.user
```

## FAQ

`Q:` `credentials missing` errors running `groupadd`

`A:` Check `aws-vault list` to see if the credentials for AWS_PROFILE (`main` by
default) are loaded. If not, try `aws-vault add <profile_name>`. If you change
profile names or credential configuration, you need to `aws-vault remove/add`.

`Q:` `Unable to parse config file: ~/.aws/credentials

`A:` Make sure you setup `~/.aws/credentials` using `aws configure`

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

[profile whatever]
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
- Plain-text credentials can be removed once you `aws-vault add ...`

## TODO

- Add tests
- Fully encapsulate all build steps (more Makefile steps inside Dockerfile)
- Update Terraform to include all IAM bits
