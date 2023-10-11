# DBoM Gateway

The DBoM gateway component for the Digital Bill of Materials

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [DBoM Gateway](#dbom-gateway)
  - [How to Use](#how-to-use)
    - [API](#api)
    - [Configuration](#configuration)
  - [\[WIP\] Helm Deployment](#wip-helm-deployment)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## How to Use

### API

Latest OpenAPI Specification for this API is available on the [API-SPECS repository](https://github.com/DBOMproject/api-specs/tree/2.0.0-alpha-1)

### Configuration

| Environment Variable         | Default                           | Description                                 |
| ---------------------------- | --------------------------------- | ------------------------------------------- |
| PORT                         | `3050`                            | The Client API port number                  |
| FED_PORT                     | `7205`                            | The federation API port number              |
| NATS_URI                     | (Example) `nats://localhost:4222` | The NATS URI                                |
| NODE_ID                      | (Example) `node1`                  | The node ID                                 |
| NODE_URI                     | (Example) `node1.test.com`         | The node URI                                |
| LOG_LEVEL                    | `info`                            | The verbosity of the logging                |
| JAEGER_ENABLED               | `false`                           | Is jaeger tracing enabled                   |
| JAEGER_HOST                  | ``                                | The jaeger host to send traces to           |
| JAEGER_SAMPLER_PARAM         | `1`                               | The parameter to pass to the jaeger sampler |
| JAEGER_SAMPLER_TYPE          | `const`                           | The jaeger sampler type to use              |
| JAEGER_SERVICE_NAME          | `Chainsource Gateway`             | The name of the service passed to jaeger    |
| JAEGER_AGENT_SIDECAR_ENABLED | `false`                           | Is jaeger agent sidecar injection enabled   |

## [WIP] Helm Deployment

[WIP] Once Completed - Instructions for deploying the database-agent using helm charts can be found [here](chainsource-gateway/README.md)