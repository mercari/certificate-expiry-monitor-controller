# Certificate Expiry Monitor Controller
[![CircleCI](https://circleci.com/gh/mercari/certificate-expiry-monitor-controller.svg?style=svg)](https://circleci.com/gh/mercari/certificate-expiry-monitor-controller)
[![codecov](https://codecov.io/gh/mercari/certificate-expiry-monitor-controller/branch/master/graph/badge.svg)](https://codecov.io/gh/mercari/certificate-expiry-monitor-controller)

Certificate Expiry Monitor Controller monitors the expiration of TLS certificates used in Ingress.

## Installation

You can apply to your cluster using the following example.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: certificate-expiry-monitor-controller
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: certificate-expiry-monitor-controller
  template:
    metadata:
      labels:
        app: certificate-expiry-monitor-controller
    spec:
      containers:
        - name: certificate-expiry-monitor-controller
          image: mercari/certificate-expiry-monitor-controller:<VERSION>
```

Once you apply above, controller will start running inside the cluster and print monitoring results to pod `stderr`.

## Usage

You can set `INTERVAL` and `THRESHOLD` as configuration. Then, the controller monitors the expiration of certificate for each set interval.
If the expiration is expired or the expiration reaches the threshold, the controller sends the alert using the configured notifier.

### Notifiers

In latest version, the contoller supports following notifiers.

- `slack`: Send information to `SLACK_CHANNEL` in your workspace using `SLACK_TOKEN`.
- `log`: Print information to `stderr`.

You can select which notifier to send an alert by configuration.
If you not select notifiers, the controller automatically selects `log`.

### Configurations

You can set following configurations by environment variables.

| ENV                | Required | Default          | Example               | Description                                                                                                                                                               |
|--------------------|----------|------------------|-----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `LOG_LEVEL`        | false    | `INFO`           | `DEBUG`, `error`      | Configuration of log level for controller's logger.                                                                                                                       |
| `KUBE_CONFIG_PATH` | false    | `~/.kube/config` | `~/.kube/config`      | Kubernetes cluster config (If not configured, controller reads local cluster config).                                                                                     |
| `INTERVAL`         | false    | `12h`            | `1m`, `24h`,          | Controller verifies expiration of certificate in Ingress at this interval of time. This value must be between `1m` and `24h`.                                             |
| `THRESHOLD`        | false    | `336h` (2 weeks) | `24h`, `100h`, `336h` | When verifing expiration, controller compares expiration of certificate and `time.Now() - THRESHOLD` to detect issue.  This value must be greater than or equal to `24h`. |
| `NOTIFIERS`        | false    | `log`            | `slack,log`           | List of alert notifiers.                                                                                                                                                  |
| `SLACK_TOKEN`      | false    | -                | -                     | Slack API token.                                                                                                                                                          |
| `SLACK_CHANNEL`    | false    | -                | `random`              | Slack channel to send expiration alert (without `#`).                                                                                                                     |

## Synthetics test management

You can use certificate-expiry-monitor-controller to generate and manage synthetics tests.
It is useful if you want to leverage an external provider synthetics to extend the controller's monitoring capabilities.
**Currently, only Datadog is supported.**

This functionality is disabled by default and can be toggled on by using the `SYNTHETICS_ENABLED` environment variable.

Supported features:

- Adding synthetics tests in Datadog
  - Using Ingress endpoint list fetched from Kubernetes API
  - Using a predefined environment variable with a list of endpoints to manage
- Deleting synthetics tests in Datadog when not matching existing endpoints

Synthetics tests have many parts configurable by environment variables:

- Alert message body
- Check frequency
- Tags
- Default tag

**Notice: To avoid unwanted destructive behavior with existing synthetics tests, a default tag is used as a safeguard. Only synthetics tests having this default tag will be handled by the controller.**

### Configuration

You can set following configurations for the synthetics test manager by using environment variables.

| ENV                | Required | Default          | Example               | Description                                                                                                                                                               |
|--------------------|----------|------------------|-----------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `SYNTHETICS_ENABLED`    | false    | false                | `false`, `true`              | Feature-flag to enable synthetics tests management. Disabled by default.
| `DATADOG_API_KEY` | false    | -           | -    | Datadog API key to manage synthetics tests                                                                       |
| `DATADOG_APPLICATION_KEY` | false    | -           | -    | Datadog application key to manage synthetics tests                                                                       |
| `SYNTHETICS_ALERT_MESSAGE` | false    | "" | `"{{#is_alert}}\n\nCertificate alert, either the expiration data is under XX days or a self-signed certificate.\n\n{{/is_alert}}\n\n @slack-jp-ms-platform-alert"`      | Alert message for synthetics tests with failing assertion                                                                  |
| `SYNTHETICS_CHECK_INTERVAL`         | false    | `900`            | `60`, `300`, `900`, `1800`, `3600`, `21600`, `43200`, `86400`, `604800`         | The interval in seconds at which the synthetics test checks will run. Lowest value is 60 seconds (1min) and highest value is 604800 seconds (1 week).                                             |
| `SYNTHETICS_TAGS`        | false    | "" | `foo:bar`, `"foo:bar, bar:foo"`  | List of tags to attribute to synthetics tests, as key:value format string separated by comma. |
| `SYNTHETICS_DEFAULT_TAG`        | false    | `managed-by-cert-expiry-mon`            | `my-control-tag`  | Default tag used to control synthetics tests managed by certificate-expiry-monitor-controller.                                                                                                                                                  |
| `SYNTHETICS_DEFAULT_LOCATIONS`        | false    | `"aws:ap-northeast-1"`            | `"aws:ap-northeast-1, aws:ap-east-1"`  | List of default locations to run synthetic tests from. [Available locations are retrievable here](https://docs.datadoghq.com/api/?lang=bash#get-available-locations)                                                                                                                                          |
| `SYNTHETICS_ADDITIONAL_ENDPOINTS`      | false    | ""                | "example.com", "example.com:8443", "example.com, example2.com:8443", "example.com:8443, example2.com:8443" | List of endpoints to add to the synthetics test controller. Useful to monitor services not served by an Ingress. Uses the format `endpoint:port`, port is optional, 443 is implied if not set.|


## Future works

- Support PagerDuty, Datadog and other services as a notifier.
- Support non-default port number. Current implementation only supports `443`.
- Support configurable alert template.

## Committers

Takamasa SAICHI ([@Everysick](https://github.com/Everysick))
Raphael FRAYSSE ([@lainra](https://github.com/lainra))

## Contribution

Please read the CLA below carefully before submitting your contribution.

https://www.mercari.com/cla/

## LICENSE

Copyright 2018 Mercari, Inc.

Licensed under the MIT License.
