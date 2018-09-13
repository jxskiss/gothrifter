# thriftkit

## TODO

- Support oneway and void method (kit)
  - [v] kit code generation
  - [ ] check if service satisfy requirement

- Refactor the messy client/invoker interface

- Refactor error handling
  - protocol and transport errors
  - kit framework errors
  - standard error codes
  - errors in server/client codes

- Rename project to thriftkit

- Implement command line option to generate go-kit codes optionally

- Remove the abandoned package lib/thrift

- Rename the package lib/thrift2 as lib/thrift

- Rename the `Server` implementation as `SimpleServer`

- Implement nocopy reader

- Support service inheritance

- Refactor structure of the generated files
  - split constants, ttypes files
  - separate service file for each service
  - separate kitclient & kitserver files for each service
  - merge encoder & decoder methods into ttypes and service files

- Generate main.go file optionally by command line option

- Generate more user-friendly client library
  - reference: google calendar api sdk

- Examples to use this library and framework

- Register service for KitServer (service discovery)
