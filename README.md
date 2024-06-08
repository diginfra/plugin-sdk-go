# plugin-sdk-go

[![Diginfra Core Repository](https://github.com/diginfra/evolution/blob/main/repos/badges/diginfra-core-blue.svg)](https://github.com/diginfra/evolution/blob/main/REPOSITORIES.md#core-scope) [![Stable](https://img.shields.io/badge/status-stable-brightgreen?style=for-the-badge)](https://github.com/diginfra/evolution/blob/main/REPOSITORIES.md#stable) [![License](https://img.shields.io/github/license/diginfra/plugin-sdk-go?style=for-the-badge)](./LICENSE)

[![Go Reference](https://pkg.go.dev/badge/github.com/diginfra/plugin-sdk-go/pkg/sdk.svg)](https://pkg.go.dev/github.com/diginfra/plugin-sdk-go/pkg/sdk)
[![Release](https://img.shields.io/github/release/diginfra/plugin-sdk-go.svg?style=flat-square)](https://github.com/diginfra/plugin-sdk-go/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/diginfra/plugin-sdk-go?style=flat-square)](https://goreportcard.com/report/github.com/diginfra/plugin-sdk-go)

## Introduction

This SDK facilitates writing [plugins](https://diginfra.org/docs/plugins) for [Diginfra](https://github.com/diginfra/diginfra) or application using [Diginfrasecurity's libs](https://github.com/diginfra/libs).

## Quick start

Before using this SDK, review the [developer's guide](https://diginfra.org/docs/plugins/developers_guide/) which fully documents the API and provides best practices for writing plugins. The developer's guide includes a [walkthrough](https://diginfra.org/docs/plugins/go-sdk-walkthrough/#example-go-plugin-dummy) of a plugin written in Go that uses this package.

For a quick start, you can refer to the provided examples:
 - [plugin with field extraction](https://github.com/diginfra/plugin-sdk-go/tree/main/examples/extractor) 
 - [plugin with event sourcing](https://github.com/diginfra/plugin-sdk-go/tree/main/examples/source)
 - [plugin with both event sourcing and field extraction](https://github.com/diginfra/plugin-sdk-go/tree/main/examples/full)



## What's next

When ready to release your plugin, make sure to register the plugin with the Diginfrasecurity organization by creating a PR to the [diginfra/plugins](https://github.com/diginfra/plugins) respository with details on the new plugin. This ensures that a given ID is used by exactly one plugin with event sourcing capability, and allows authors of plugins with field extraction capability to coordinate about event source formats.

## Join the Community

To get involved with The Diginfra Project please visit [the community repository](https://github.com/diginfra/community) to find more.

How to reach out?

 - Join the [#diginfra](https://kubernetes.slack.com/messages/diginfra) channel on the [Kubernetes Slack](https://slack.k8s.io)
 - [Join the Diginfra mailing list](https://lists.cncf.io/g/cncf-diginfra-dev)


## Contributing

See the [CONTRIBUTING.md](https://github.com/diginfra/.github/blob/master/CONTRIBUTING.md).

## Security Audit

A third party security audit was performed by Cure53, you can see the full report [here](https://github.com/diginfra/diginfra/blob/master/audits/SECURITY_AUDIT_2019_07.pdf).

## Reporting security vulnerabilities

Please report security vulnerabilities following the community process documented [here](https://github.com/diginfra/.github/blob/master/SECURITY.md).

## License Terms

This project is licensed to you under the [Apache 2.0](./LICENSE) open source license.


