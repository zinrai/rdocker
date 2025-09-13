# rdocker

`rdocker` is a command-line tool that allows you to run Docker and Docker Compose commands on a remote host while using your local project files. It simplifies the process of deploying and managing Docker containers on remote servers.

## Features

- Run Docker and Docker Compose commands on a remote host
- Automatically sync your local project directory to the remote host
- Validate the presence of necessary files (Dockerfile or docker-compose.yml) before executing commands

## Installation

Build the binary:

```
$ go build
```

## Remote Server Setup

Before using `rdocker`, you need to prepare your remote server. Follow these steps:

1. Install the required packages on the remote server:

   ```
   $ sudo apt install rsync docker.io docker-compose
   ```

2. Configure sudo to allow the user to run Docker commands without a password. Replace `nanashi` with the username you'll use to connect to the server:

   ```
   $ sudo tee /etc/sudoers.d/docker << EOF
   nanashi ALL=(ALL) NOPASSWD: /usr/bin/docker, /usr/bin/docker-compose
   EOF
   ```

   This configuration allows the specified user to run `docker` and `docker-compose` commands with sudo privileges without requiring a password.

## Usage

The basic syntax for using `rdocker` is:

```
$ rdocker [options] <remote_host> -- <docker/docker-compose command>
```

Examples:

1. Run `docker ps` on a remote host:
   ```
   $ rdocker -u nanashi remote.example.com -- docker ps
   ```

2. Run `docker-compose up -d` on a remote host:
   ```
   rdocker -u nanashi -k ~/.ssh/id_rsa remote.example.com -- docker-compose up -d
   ```

## How It Works

1. `rdocker` first checks if the necessary files (Dockerfile for Docker commands, docker-compose.yml or compose.yaml for Docker Compose commands) are present in your current directory.

2. It creates a temporary directory on the remote host.

3. It syncs your current directory to the temporary directory on the remote host (excluding the .git directory).

4. It executes the specified Docker or Docker Compose command on the remote host in the synced directory.

5. The output of the command is displayed in your local terminal.

## Important Notes

- The remote host must have Docker, Docker Compose, and rsync installed (see Remote Server Setup).
- The user specified with the `-u` option must have sudo privileges to run Docker commands on the remote host, configured as shown in the Remote Server Setup section.
- Ensure that your SSH key is properly set up for passwordless authentication to the remote host.
- The tool uses `rsync` for file synchronization, so make sure it's installed on both your local machine and the remote host.

## License

This project is licensed under the [MIT License](./LICENSE).
