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
- `teams`: Send information to `TEAMS_WEBHOOK`.
- `log`: Print information to `stderr`.

You can select which notifier to send an alert by configuration.
If you not select notifiers, the controller automatically selects `log`.

### Configurations

You can set following configurations by environment variables.

| ENV                | Required | Default          | Example                                                                 | Description                                                                                                                                                               |
|--------------------|----------|------------------|-------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `LOG_LEVEL`        | false    | `INFO`           | `DEBUG`, `error`                                                        | Configuration of log level for controller's logger.                                                                                                                       |
| `KUBE_CONFIG_PATH` | false    | `~/.kube/config` | `~/.kube/config`                                                        | Kubernetes cluster config (If not configured, controller reads local cluster config).                                                                                     |
| `INTERVAL`         | false    | `12h`            | `1m`, `24h`,                                                            | Controller verifies expiration of certificate in Ingress at this interval of time. This value must be between `1m` and `24h`.                                             |
| `THRESHOLD`        | false    | `336h` (2 weeks) | `24h`, `100h`, `336h`                                                   | When verifing expiration, controller compares expiration of certificate and `time.Now() - THRESHOLD` to detect issue.  This value must be greater than or equal to `24h`. |
| `NOTIFIERS`        | false    | `log`            | `slack,teams,log`                                                       | List of alert notifiers.                                                                                                                                                  |
| `SLACK_TOKEN`      | false    | -                | -                                                                       | Slack API token.                                                                                                                                                          |
| `SLACK_CHANNEL`    | false    | -                | `random`                                                                | Slack channel to send expiration alert (without `#`).                                                                                                                     |
| `TEAMS_WEBHOOK`    | false    | -                | `https://outlook.office.com/webhook/{id}/IncomingWebhook/{incoming_id}` | Microsoft Teams channel to send expiration alert, to get the Webhook click on 'More' on your desired channel and create an Incoming Webhook                               |

## Future works

- Support PagerDuty, Datadog and other services as a notifier.
- Support non-default port number. Current implementation only supports `443`.
- Support configurable alert template.

## Committers

Takamasa SAICHI ([@Everysick](https://github.com/Everysick))

## Contribution
Please read the CLA below carefully before submitting your contribution.

https://www.mercari.com/cla/

## LICENSE
Copyright 2018 Mercari, Inc.

Licensed under the MIT License.
