# thriftkit

## About

This is an exercise project to learn about the Thrift RPC framework, the
[go-kit](https://github.com/go-kit/kit/) toolkit and golang performance tuning.

It's TOTALLY NOT ready to be used in any production environment.

## TODO

- [ ] Support oneway and void method (kit)
  - [x] kit code generation
  - [ ] check if service satisfy requirement

- [x] Refactor the messy client/invoker interface

- [ ] Refactor error handling
  - [ ] protocol and transport errors
  - [ ] kit framework errors
  - [ ] standardize error codes
  - [ ] errors in server/client codes

- [x] Remove the abandoned package lib/thrift

- [x] Rename the package lib/thrift2 as lib/thrift

- [ ] Implement command line option to generate go-kit codes optionally

- [ ] Implement nocopy reader

- [ ] Support service inheritance

- [ ] Refactor structure of the generated files
  - [ ] split constants, ttypes files
  - [ ] separate service file for each service
  - [ ] separate kitclient & kitserver files for each service
  - [ ] merge encoder & decoder methods into ttypes and service files

- [ ] Restructure the template files

- [ ] Generate main.go file optionally by command line option

- [x] Examples to use this library and framework

- [x] Register service for KitServer (service discovery)

- [ ] Rename the `Server` implementation as `SimpleServer`

- [ ] Rename project to thriftkit
