# Developer documentation

## Example services

Currently the project includes two example services.
- The [Example](./example.md) service is meant to be a simple example that demonstrates how to write a service using the framework.  The core functionality doesn't perform any useful actions, but it should serve as a good starting point.
- the [Lab project](./labproject.md) service is a more useful, real world service.  It's the service that the framework was originally built to serve and so informed a lot of the design decisions of the early stages of the project.


## Writing a service

Each service must be initialised with at least three arguments:
- A form instance.
- A PreAction function or a PostAction function (or both).
- A configuration instance.

### Forms

[![PkgGoDev](https://pkg.go.dev/badge/github.com/g-node/tonic)](https://pkg.go.dev/github.com/G-Node/tonic/tonic/form#Form)

A Form consists of one or more Pages.
A Page consists of one or more Elements.
Each Element defines an HTML form input of a given type.

Tonic uses this Form definition to create a web form with the given elements.  The form is displayed on a single page, but the different pages define separate sections of the form.

### PreAction and PostAction functions

The Action functions serve to process information on behalf of the user.
The [PreAction](#preaction) is meant to process information before the Form is displayed (e.g., to specify allowed values or valid elements).
The [PostAction](#postaction) is meant to process the data submitted by the user through the Form.

#### PreAction

The PreAction function takes three arguments:
1. A copy of the Form instance defined for the service.
2. A Client representing the service itself (bot client).
3. A Client representing the user that's logged in (user client).

The PreAction should use the bot and user clients to determine if any modifications to the form are necessary and return a new instance of the Form with any modifications made.  The following is a non-exhaustive list of examples of some modifications that the PreAction could make to the Form:
- Set a list of values for a Select type element based on the options relevant to the user.
- Disable an element that the specific user shouldn't have access to.
- Fill in the value of an element and mark it _read only_ if it only has one valid value.

In the case of services that don't define a PostAction, the PreAction can be used to build information-only services, that is, services that simply display information to a user but don't take any input.

#### PostAction

The PostAction function takes three arguments:
1. A map of values (`string->string`) that represent the values the user entered into the form.
2. A Client representing the service itself (bot client).
3. A client representing the user that's logged in (user client).

The PostAction should use the bot client to perform actions for the using the bot client and return a list of messages.  The messages should serve as an action log for users to review or to troubleshoot any problems.  The following is a non-exhaustive list of examples of some actions that the PostAction could perform for a user:
- Create a repository in an organisation in which the user is not an owner (see the [lab project](./labproject.md) microservice).
- Invite another user to a team in an organisation in which the user it not an owner.
- Rename or otherwise modify a repository of which they're not an admin.

The purpose of services like these is to give users the ability to perform specific administrative-level actions without giving them full administrative rights.
