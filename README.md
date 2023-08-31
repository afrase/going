# Going

Small CLI tool for working with AWS. This tool is mostly used to log in to ECS containers.

# Usage

All commands will prompt for an AWS profile defined in the shared AWS config file `$HOME/.aws/config`.
You can use the `-p, --profile` flag to specify a profile and not be prompted.

All prompts have fuzzy searching.

## shell command

You can connect to an ECS container using the `shell` command.
This command will make sure the cached SSO credentials are valid and if not will attempt to refresh them.

```shell
going shell -p staging
```

This will prompt you to select a cluster, service, task (if more than 1 is running), and a container.

If the ExecuteCommand agent isn't running in the container you can use the `--ssm` flag to use SSM directly.

```
Open a shell to a container in ECS

Usage:
  going shell [flags]

Flags:
  -c, --cluster string     The cluster name
  -r, --container string   The container name
  -h, --help               help for shell
  -s, --service string     The service name
      --ssm                Use SSM directly to get a shell

Global Flags:
  -p, --profile string   The AWS profile to use
```

## sso command

The `sso` command by itself will print the AWS `access_key`, `secret_key`, and `token` credentials.
Adding the `-e, --env` flag will output the credentials as environment variables that can be pasted into a `.env` file.

This command has multiple sub-commands for login/logout and an automatic way of replacing AWS environment variables.

```shell
going sso --env
```

### replace command

The `replace` command will try to replace the environment variables for the AWS credentials in `.env` files.
The command takes either a list of files as positional arguments or if no flags exist it will try `$PWD/.env`.

```shell
going sso replace /project1/.env /project2/.env
```

The way this is done currently is pretty simple, it will only update keys that exist in the file and not add missing ones.
The keys it updates are `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_SESSION_TOKEN`.

### login command

This command will perform a full SSO login, ignoring the cached SSO credentials.

```shell
going sso login
```

### logout command

This command will perform a logout of the current SSO session by telling AWS to invalidate the session and deleting the cached credential file.

```shell
going sso logout
```
