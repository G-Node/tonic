# Example service

The [example](/utonics/example/main.go) service implements a very simple service that serves no useful purpose, but demonstrates how a service can be built.
It defines the following service components:

## Form

The Form consists of two pages.  The first page contains 3 elements:
1. Name: a simple text field.
2. Descritpion: a long form text area.
3. Duration: the duration in seconds that the action will run for.

A second page includes one of each type of input supported for demonstration purposes.

## PostAction

The PostAction of the service simply waits for the defined duration specified by the form and then returns a message log of the values it received.

## Config

The configuration does not specify a GIN server to connect to (or credentials) since it doesn't perform any online tasks.
The path to the database, a port number for the service, and a cookie name for the user's session are all required.
