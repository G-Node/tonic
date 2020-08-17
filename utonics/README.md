# µTonics

µTonics (micro-tonics) are micro-services built using Tonic.  The all have the following common components and flow:

- User login using [GIN](https://gin.g-node.org) credentials.
- Presentation of a [web form](#form).
- Form data is submitted to a worker pool to run a specific [job](#job).
- Queued and completed jobs can be viewed in the job log.

## Form

Each µTonic defines its own web form using a series of Element structs.

## Job

Each µTonic defines its own job.  A job is a function that has access to the form data.

## Examples

_WIP_
