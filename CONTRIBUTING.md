# Contributing to uts-proxy

## Development

### Prerequisites

- Go 1.22 or later

### Building

```bash
go build -o uts-proxy .
```

### Running tests

```bash
go test -race -v ./...
```

### Code style

This project uses standard Go formatting. Before submitting a PR:

```bash
go vet ./...
gofmt -l .
```

## Submitting a pull request

1. Fork this repository
2. Create a feature branch from `main`
3. Make your changes
4. Ensure tests pass: `go test -race ./...`
5. Submit a pull request

## Reporting issues

If you find a bug or have a feature request, please [open an issue](https://github.com/ably/uts-proxy/issues).

## License

By contributing to this project, you agree that your contributions will be licensed under the [Apache License 2.0](./LICENSE).
