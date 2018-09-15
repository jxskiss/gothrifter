# thriftkit

## TODO

- Support oneway and void method (kit)
  - [v] kit code generation
  - [ ] check if service satisfy requirement

- [v] Refactor the messy client/invoker interface

- Refactor error handling
  - protocol and transport errors
  - kit framework errors
  - standard error codes
  - errors in server/client codes

- [v] Remove the abandoned package lib/thrift

- [v] Rename the package lib/thrift2 as lib/thrift

- Implement command line option to generate go-kit codes optionally

- Implement nocopy reader

- Support service inheritance

- Refactor structure of the generated files
  - split constants, ttypes files
  - separate service file for each service
  - separate kitclient & kitserver files for each service
  - merge encoder & decoder methods into ttypes and service files

- Restructure the template files

- Generate main.go file optionally by command line option

- Examples to use this library and framework

- [v] Register service for KitServer (service discovery)

- Rename the `Server` implementation as `SimpleServer`

- Rename project to thriftkit

- Generate more user-friendly client library
  - reference: google calendar api sdk
