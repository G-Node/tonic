# Lab project service

The [lab project](/utonics/labproject/main.go) service takes advantage of almost all the features of Tonic.  Its purpose is to create repositories inside an organisation on GIN where the user is not an administrator or owner.
It defines the following service components:

## Form

The Form consists of a two pages.

The first page consists of three elements:
1. Lab organisation: The name of the organisation in which the repository is going to be created.
2. Project name: The name of the repository or repositories to be created.
3. Description: A long description for the project.

The second page includes toggles for submodules that the user may or may not require for the project.

## PreAction

The PreAction determines whether there are any organisations on GIN in which the user is allowed to create repositories.  The requirements for an organisation to be eligible are:
- The service bot must be an owner or administrator of the organisation.
- The user must be a member of the organisation.

Since multiple organisations may fit the requirements, the PreAction determines their names and sets the available values for the _Lab organisation_ select field.

## PostAction

The PostAction receives the form values and creates the repository on behalf of the user.  Before doing so, it checks if the organisation is in fact eligible using the same criteria defined in the [PreAction](#preaction).

If the input is valid, the service performs the following actions:
- Create a **repository** with the _Project name_ as defined by the user.
- Create a **team** with the _Project name_ as defined by the user.
- Adds the service user to the **team**.
- Adds the **repository** to the **team**.

*Note: More actions are planned for this service and will be added soon.*

## Config

The configuration specifies the GIN server to connect to (currently gin.dev.g-node.org for testing purposes).  The username and password are for the bot user that the service uses to perform actions.  These credentials should be provided through a json file called `testbot`.
