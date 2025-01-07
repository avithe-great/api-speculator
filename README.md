# api-speculator

```shell
$ ./speculator -h
speculator helps you to secure your APIs by identifying shadow and zombie APIs.

By analyzing API traffic in conjunction with your API specifications (e.g., OpenAPI, Swagger), speculator can detect:
  * Shadow APIs: Endpoints that are implemented and functional but not documented in your API specification.
  * Zombie APIs: Endpoints that are deprecated or abandoned in your API specification but they are still in use.

Usage:
  speculator [flags]

Flags:
      --config string   config file path
      --debug           run in debug mode
  -h, --help            help for speculator
```