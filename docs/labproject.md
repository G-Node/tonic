# Lab project service

The [lab project](/utonics/labproject/main.go) service takes advantage of almost all the features of Tonic.
The purpose of the service is to assist members of a Research Lab (represented by an Organisation on GIN) with using the GIN services in a way that promotes reproducible research practices.

> Note: The following list contains planned features that are not yet implemented.

The service defines a set of forms that can perform administrative actions on behalf of users such as:
- [x] Creating a repository structure with submodules based on a pre-defined research project template.
- [x] Creating teams to group users and repositories and to control access permissions.
- [ ] Modifying existing repositories and teams during the project lifetime.

## Setup and configuration

The configuration specifies the GIN server to connect to (currently gin.dev.g-node.org for testing purposes).  The username and password are for the bot user that the service uses to perform actions.  These credentials should be provided through a json file called `testbot`.

## Internals and components

Internally the service performs actions as an administrator of the organisation and sometimes the whole GIN instance (site admin).
This is represented by a `worker.Client`, which is responsible for making requests to the GIN API and performing git operations for the creation of repositories.

The service defines the following components:

### Form

The form consists of three elements:
1. Lab organisation: The name of the organisation in which the repository is going to be created.
2. Project name: The name of the repository or repositories to be created.
3. Optionally a team name where the repository will be added. If none is indicated, a team with the same name as the repository will be created.
4. Description: A long description for the project.

### PreAction

The PreAction determines whether there are any organisations on GIN in which the user is allowed to create repositories.  The requirements for an organisation to be eligible are:
- The service bot must be an owner or administrator of the organisation.
- The user must be a member of the organisation.

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
- Adds the logged in user to the **team**.
- Adds the **repository** and its **submodules** to the **team**.
