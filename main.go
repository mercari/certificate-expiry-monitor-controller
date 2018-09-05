package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mercari/certificate-expiry-monitor-controller/config"
	"github.com/mercari/certificate-expiry-monitor-controller/controller"
	logging "github.com/mercari/certificate-expiry-monitor-controller/log"
	"github.com/mercari/certificate-expiry-monitor-controller/notifier"
	"github.com/mercari/certificate-expiry-monitor-controller/notifier/datadog"
	"github.com/mercari/certificate-expiry-monitor-controller/notifier/log"
	"github.com/mercari/certificate-expiry-monitor-controller/notifier/slack"

	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	os.Exit(runMain())
}

func runMain() int {
	// Parse configurations from environment variables
	var env config.Env
	if err := env.ParseEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to parse environement variables: %s\n", err.Error())
		return 1
	}

	// Setup clientSet from configuration.
	clientSet, err := newClientSet(env.KubeconfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to create clientSet: %s\n", err.Error())
		return 1
	}

	// Setup notifiers from configuration.
	// If user specify unsupported notifier name, program returns exit code `1`.
	notifiers := make([]notifier.Notifier, len(env.Notifiers))
	for i, name := range env.Notifiers {
		switch name {
		case slack.String():
			sl, err := slack.NewNotifier(env.SlackToken, env.SlackChannel)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Failed to create slack notifier: %s\n", err.Error())
				return 1
			}

			notifiers[i] = sl
		case log.String():
			logger, err := logging.NewLogger(log.AlertLogLevel())
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Failed to create log notifier: %s\n", err.Error())
				return 1
			}

			notifiers[i] = log.NewNotifier(logger)
		case datadog.String():
			dd, err := datadog.NewNotifier(env.DatadogToken, env.DatadogAddress, env.DatadogTags, env.DatadogGaugeKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Failed to create datadog notifier: %s\n", err.Error())
				return 1
			}

			notifiers[i] = dd
		default:
			fmt.Fprintf(os.Stderr, "[ERROR] Unexpected notifier name: %s\n", name)
			return 1
		}
	}

	// Setup logger that wrapped zap.Logger to use common settings.
	logger, err := logging.NewLogger(env.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to create logger: %s\n", err.Error())
		return 1
	}

	// Create new contoller instance.
	controller, err := controller.NewController(logger, clientSet, env.VerifyInterval, env.AlertThreshold, notifiers)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Failed to create controller: %s\n", err.Error())
		return 1
	}

	// When controller receives SIGINT or SIGTERM,
	// handleSignal goroutine triggers stopCh to terminate controller.
	stopCh := make(chan struct{}, 1)
	go handleSignal(stopCh)

	controller.Run(stopCh)

	return 0
}

// Create new Kubernetes's clientSet.
// When configured env.KubeconfigPath, read config from env.KubeconfigPath.
// When not configured env.KubeconfigPath, read internal cluster config.
func newClientSet(kubeconfigPath string) (*kubernetes.Clientset, error) {
	var err error
	var config *rest.Config

	if kubeconfigPath == "" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientSet, nil
}

// Handling syscall.SIGTERM and syscall.SIGINT
// When trap those, function send message to stopCh
func handleSignal(stopCh chan struct{}) {
	signalCh := make(chan os.Signal)

	signal.Notify(signalCh, syscall.SIGTERM)
	signal.Notify(signalCh, syscall.SIGINT)

	<-signalCh
	close(stopCh)
}
