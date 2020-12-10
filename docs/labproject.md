# Lab project service

The [lab project](/utonics/labproject/main.go) service takes advantage of almost all the features of Tonic.
The purpose of the service is to assist members of a Research Lab (represented by an Organisation on GIN) with using the GIN services in a way that promotes reproducible research practices.

> **NOTE:** The following list contains planned features that are not yet implemented.

The service defines a set of forms that can perform administrative actions on behalf of users such as:
- [x] Creating a repository structure with submodules based on a pre-defined research project template.
- [x] Creating teams to group users and repositories and to control access permissions.
- [ ] Modifying existing repositories and teams during the project lifetime.

## Setup and configuration

### Service user

The service provides access to administrative tasks on behalf of non-privileged users in an organisation.
This requires that the service have its own account (or access to one) that has administrative privileges (owner or admin) for the organisation it will support.
 This account will be called `bot user` here, in order to distinguish it from the non-privileged user who will use the service.
Before starting the service setup, **create a user on the GIN server** which will become the `bot user` the service will work with.
The credentials for the new `bot user` will be required for the service configuration.

**Add the `bot user` to the organisation(s)** which it will support, and
give it admin rights by adding it to the **Owners** group.

### Service configuration

Create a file called `labproject.json` (later called via `/path/to/labproject.json`, so note where you saved it) with the following content:
```json
{
  "gin": {
    "web": "<web address for GIN service: required>",
    "git": "<git address for GIN service: required>",
    "username": "<bot user username: required>",
    "password": "<bot user pasword: required>"
  },
  "templaterepo": "<template repository: required>",
  "cookiename": "<session cookie name: optional (default: utonic-labproject)>",
  "port": <port for service to listen on: optional (default: 3000)>,
  "dbpath": "<path to sqlite database file: optional (default: ./labproject.db)>"
}
```

- The `web` value must specify both the protocol scheme and the port, even if it's the standard one, e.g., `https://gin.g-node.org:443`.
- The `git` value must specify the user and the port, even if it's the standard one, e.g., `git@gin.g-node.org:22`.
- The `username` and `password` must match the credentials of the `bot user` that was created in the previous step.

If any of the above values is incorrect, the service will fail to start.

- The `templaterepo` should be of the form `user/repository` and will be used as the template for all new projects.
No check is made on startup to determine if the repository exists.

- The `cookiename` value can be any name or word.
It is used to name the session cookie stored in users' browsers.
- The `port` is the port the service will listen on.
Port numbers below 1024 require elevated privileges on the server (or inside the container).
Note that unlike the rest of the options, the port value is a number and should not be quoted.
- The `dbpath` value should point to an accessible path.
If the file does not exist on startup, an empty database will be created.

### Compile and run

Requires Go v1.15 or newer.

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

By default the service runs on port 3000, so you can access the example form at http://localhost:3000.
If you specified a different port in the [Service configuration](#service-configuration) section, use that one instead.

If the credentials are incorrect (see the [Service configuration](#service-configuration) section above), the startup will fail.

Press `ctrl+c` to stop the service.

There is no built-in way to run the service in a daemon/background mode.
For that, you must write a service file for your platform (e.g., a systemd service file).
Alternatively, read below for setting up the service using  Docker.

### Docker

> *NOTE:* Please refer to Docker documentation to install Docker and learn how to use it, the commands that follows can be pasted in a bash/shell window.

Clone the repository and build the image:
```
git clone https://github.com/G-Node/tonic
cd tonic
docker build --build-arg service=labproject -t local/tonic:labproject .
```

The `--build-arg service=labproject` option specifies which service to build.
If omitted, it will build the [Example](#example) service.

> *NOTE:* Here the image is named `local/tonic` and tagged as `labproject`, but this could be named anything.

For the first run, an empty file must be created for the database that will be mapped into the container:
```
touch /path/to/labproject.db
```
> *NOTE:* `/path/to/` should be modified.
It should be the relative path from your current terminal's working directory or an absolute path.
You will have to modify it to the same path in the following commands.

To start the service run:
```
docker run -it --rm --publish 3000:3000 --volume /path/to/labproject.db:/tonic/labproject.db --volume /path/to/labproject.json:/tonic/labproject.json --name labproject local/tonic:labproject
```

> *NOTE:* The `--rm` flag will delete the container once it exits.
The mapped files (`labproject.db` and `labproject.json`) will remain unaffected on the host.

> *NOTE:* The `-i` and `-t` flags (combined `-it`) attach the running container to the foreground in the terminal.
This is useful for testing and troubleshooting, but for regular usage, you may want to omit them to make the container run silently in the background.

The `--volume /path/to/labproject.json:/tonic/labproject.json` option places the configuration file (see [Configuration](#configuration) above) into the running container for the service to read.
The path `/path/to/labproject.json` must be changed to a file on disk with the configuration values.
The `--volume /path/to/labproject.db:/tonic/labproject.db` option places the database file (see [Configuration](#configuration) above) into the running container for the service to read.
It is important that the file already exists outside the container, otherwise it will be created as a directory on service startup and the service will fail with an error.

The `--publish 3000:3000` option publishes port 3000 from inside the container to the host system's network (even externally).
It makes the running container accessible at `http://localhost:3000` or at `http(s)://server-address:3000`.
If omitted, the container can be accessed from the container's internal IP address, which can be determined using `docker inspect labproject`.
For production environments, it is considered best practice to configure a reverse proxy web service to forward the server's external web port (80 or 443) to the docker internal IP address, instead of publishing the container port directly.

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

Type `ctrl+c` to stop the service if it is attached (using `-it`).
Otherwise, you can stop it with `docker stop labproject`, where `labproject` is the value given to the `--name` argument in the `docker run` call.
You may also use the docker desktop app.

## Internals and components

Internally the service performs actions as an administrator of the organisation and sometimes the whole GIN instance (site admin).
This is represented by a `worker.Client`, which is responsible for making requests to the GIN API and performing git operations for the creation of repositories.

The service defines the following components:

### Form

The form consists of three elements:
1. Lab organisation: The name of the organisation in which the repository is going to be created.
2. Project name: The name of the repository or repositories to be created.
3. Team: A team name to be created where the repository and the user who requested the creation will be added.
If left empty, a team with the same name as the repository (i.e., the project name) will be created instead.
4. Title: A title for the project.
The project title is added to the repository's Description/Title field on GIN.

### PreAction

The PreAction determines whether there are any organisations on GIN in which the user is allowed to create repositories.  The requirements for an organisation to be eligible are:
- The `bot user` must be an owner or administrator of the organisation.
- The logged-in user must be a member of the organisation.

Since multiple organisations may fit the requirements, the PreAction determines their names and sets the available values for the _Lab organisation_ select field.

### PostAction

The PostAction receives the form values and creates the repository on behalf of the user.  Before doing so, it checks if the organisation is in fact eligible using the same criteria defined in the [PreAction](#preaction).

If the input is valid, the service performs the following actions:
- Clone the **template repository** (specified in the main service configuration) and all its submodules.
- Create a **repository** with the _Project name_ as defined by the user in the _Organisation_ on the server.
  - Create one repository for each **submodule** found in the template with the name _Project name_._submodule name_ in the _Organisation_ on the server.
- Configure the **repository** with a new remote pointing to the newly created repository on the server.
  - Configure each **submodule** with a new remote pointing to the newly created submodule repositories on the server.
- Push the **repository contents** to the server.
  - Push each **submodule's contents** to the server.
- Create a **team** with the _Team name_ provided by the user (or _Project name_ if unspecified).
The team gets write permissions on its repositories.
- Adds the logged in user to the **team**.
- Adds the **repository** and its **submodules** to the **team**.
