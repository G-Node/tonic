# Quick start

## Example service

The example service is simply for demonstration purposes. See the [Example service](./example.md) document for a detailed description.

### Compile and run

Requires Go v1.15.

Clone this repository, build the included services, and run:
```
git clone https://github.com/G-Node/tonic
cd tonic
make
./build/example
```

After calling the last command, the output should be similar to the following:
```
tonic: 2020/10/02 13:31:06 Initialising database
tonic: 2020/10/02 13:31:06 Initialising worker
tonic: 2020/10/02 13:31:06 Initialising web service
tonic: 2020/10/02 13:31:06 Setting up router
tonic: 2020/10/02 13:31:06 Starting worker
tonic: 2020/10/02 13:31:06 Worker started
tonic: 2020/10/02 13:31:06 Starting web service
tonic: 2020/10/02 13:31:06 Web server started
tonic: 2020/10/02 13:31:06 No server configured - skipping login and disabling login requirements
tonic: 2020/10/02 13:31:06 WARNING: Authentication is open!
```

By default the service runs on port 3000, so you can access the example form at http://localhost:3000

Type `ctrl+c` to stop the service.

### Docker

Clone the repository and build the image:
```
git clone https://github.com/G-Node/tonic
cd tonic
docker build -t local/tonic:example .
```

*NOTE:* Here the image is named `local/tonic` and tagged as `example`, but this could be named anything.

For the example, no options or external files are required to run:
```
docker run --rm -p 3000:3000 local/tonic:example
```
*NOTE:* The `--rm` flag will delete the container once it exits.

The `-p 3000:3000` option publishes port 3000 from inside the container to the host system's network (even externally). It makes the running container accessible at http://localhost:3000. If omitted, the container can be accessed from the container's internal IP address, which can be determined using `docker inspect`.

The output should be similar to the following:
```
tonic: 2020/10/02 13:31:06 Initialising database
tonic: 2020/10/02 13:31:06 Initialising worker
tonic: 2020/10/02 13:31:06 Initialising web service
tonic: 2020/10/02 13:31:06 Setting up router
tonic: 2020/10/02 13:31:06 Starting worker
tonic: 2020/10/02 13:31:06 Worker started
tonic: 2020/10/02 13:31:06 Starting web service
tonic: 2020/10/02 13:31:06 Web server started
tonic: 2020/10/02 13:31:06 No server configured - skipping login and disabling login requirements
tonic: 2020/10/02 13:31:06 WARNING: Authentication is open!
```

Type `ctrl+c` to stop the service.


## Lab project service

The Lab project service requires a configuration and credentials to make calls against a GIN API and clone repositories. See the [Lab project service](./labproject.md) document for a detailed description.

### Configuration

The configuration is read from a file called `labproject.json` in the working directory where the service is launched.
The following configuration keys are supported:
```json
{
  "gin": {
    "web": "<web address for GIN service: required>",
      "git": "<git address for GIN service: required>",
      "username": "<service username: required>",
      "password": "<service pasword: required>"
  },
  "templaterepo": "<template repository: required>",
  "cookiename": "<session cookie name: optional (default: utonic-labproject)>",
  "port": "<port for service to listen on: optional (default: 3000)>",
  "dbpath": "<path to sqlite database file: optional (default: ./labproject.db)>"
}
```

If any configuration values marked `required` are not specified, the service will fail to start.
Omitting optional values will set and print the default value on startup.

### Compile and run

Requires Go v1.15.

Clone this repository, build the included services, and run:
```
git clone https://github.com/G-Node/tonic
cd tonic
make
./build/labproject
```

After calling the last command, the output should be similar to the following:
```
tonic: 2020/10/02 13:56:13 Initialising database
tonic: 2020/10/02 13:56:13 Initialising worker
tonic: 2020/10/02 13:56:13 Initialising web service
tonic: 2020/10/02 13:56:13 Setting up router
tonic: 2020/10/02 13:56:13 Starting worker
tonic: 2020/10/02 13:56:13 Worker started
tonic: 2020/10/02 13:56:13 Starting web service
tonic: 2020/10/02 13:56:13 Web server started
tonic: 2020/10/02 13:56:13 Logging in to gin
tonic: 2020/10/02 13:56:13 Logged in and ready
```

By default the service runs on port 3000, so you can access the example form at http://localhost:3000

If the credentials are incorrect (see [Configuration](#configuration) above), the startup will fail.

Type `ctrl+c` to stop the service.

### Docker

Clone the repository and build the image:
```
git clone https://github.com/G-Node/tonic
cd tonic
docker build --build-arg service=labproject -t local/tonic:labproject .
```

The `--build-arg service=labproject` option specifies which service to build. If omitted, it will build the [Example](#example) service.

*NOTE:* Here the image is named `local/tonic` and tagged as `labproject`, but this could be named anything.

For the first run, an empty file must be created for the database that will be mapped into the container:
```
touch /path/to/labproject.db
```

To start the service run:
```
docker run -it --rm --volume /path/to/labproject.db:/tonic/labproject.db --volume /path/to/labproject.json:/tonic/labproject.json --name labproject local/tonic:labproject
```

*NOTE:* The `--rm` flag will delete the container once it exits.

The `--volume /path/to/labproject.json:/tonic/labproject.json` option places the configuration file (see [Configuration](#configuration) above) into the running container for the service to read. The path must be changed to a file on disk with the configuration values.
The `--volume /path/to/labproject.db:/tonic/labproject.db` option places the database file (see [Configuration](#configuration) above) into the running container for the service to read. It is important that the file already exists outside the container, otherwise it will be created as a directory on service startup and the service will fail with an error.

The `-p 3000:3000` option publishes port 3000 from inside the container to the host system's network (even externally). It makes the running container accessible at http://localhost:3000. If omitted, the container can be accessed from the container's internal IP address, which can be determined using `docker inspect`.

The output should be similar to the following:
```
tonic: 2020/11/23 12:43:21 Initialising database
tonic: 2020/11/23 12:43:21 Initialising worker
tonic: 2020/11/23 12:43:21 Initialising web service
tonic: 2020/11/23 12:43:21 Setting up router
tonic: 2020/11/23 12:43:21 Starting worker
tonic: 2020/11/23 12:43:21 Worker started
tonic: 2020/11/23 12:43:21 Starting web service
tonic: 2020/11/23 12:43:21 Web server started
tonic: 2020/11/23 12:43:21 Logging in to gin (<gin web address>)
tonic: 2020/11/23 12:43:21 Logged in and ready
```

Type `ctrl+c` to stop the service.
