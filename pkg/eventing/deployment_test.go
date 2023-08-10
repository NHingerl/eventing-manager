package eventing

import (
	"fmt"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	"github.com/kyma-project/eventing-manager/pkg/env"
	"github.com/kyma-project/eventing-manager/test"
)

const (
	natsURL         = "eventing-nats.kyma-system.svc.cluster.local"
	eventTypePrefix = "test.prefix"
)

func TestNewDeployment(t *testing.T) {
	publisherConfig := env.PublisherConfig{
		RequestsCPU:     "32m",
		RequestsMemory:  "64Mi",
		LimitsCPU:       "100m",
		LimitsMemory:    "128Mi",
		Image:           "testImage",
		ImagePullPolicy: "Always",
		AppLogLevel:     "info",
		AppLogFormat:    "json",
	}
	testCases := []struct {
		name                  string
		givenPublisherName    string
		givenBackend          string
		wantBackendAssertions func(t *testing.T, publisherName string, deployment appsv1.Deployment)
	}{
		{
			name:                  "NATS should be set properly after calling the constructor",
			givenPublisherName:    "test-name",
			givenBackend:          "NATS",
			wantBackendAssertions: natsBackendAssertions,
		},
		{
			name:                  "EventMesh should be set properly after calling the constructor",
			givenPublisherName:    "test-name",
			givenBackend:          "EventMesh",
			wantBackendAssertions: eventMeshBackendAssertions,
		},
	}

	publisherName := "test-name"
	publisherNamespace := "test-namespace"
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var deployment *appsv1.Deployment
			var natsConfig env.NATSConfig

			switch tc.givenBackend {
			case "NATS":
				natsConfig = env.NATSConfig{
					JSStreamName:    "kyma",
					URL:             natsURL,
					EventTypePrefix: eventTypePrefix,
				}
				deployment = newNATSPublisherDeployment(publisherName, publisherNamespace, natsConfig, publisherConfig)
			case "EventMesh":
				deployment = newEventMeshPublisherDeployment(publisherName, publisherNamespace, publisherConfig)
			default:
				t.Errorf("Invalid backend!")
			}

			// the tight backenType should be set
			assert.Equal(t, deployment.ObjectMeta.Labels[BackendLabelKey], tc.givenBackend)
			assert.Equal(t, deployment.ObjectMeta.Labels[AppLabelKey], publisherName)

			// check the container properties were set properly
			container := findPublisherContainer(publisherName, *deployment)
			assert.NotNil(t, container)

			assert.Equal(t, fmt.Sprint(container.Name), publisherName)
			assert.Equal(t, fmt.Sprint(container.Image), publisherConfig.Image)
			assert.Equal(t, fmt.Sprint(container.ImagePullPolicy), publisherConfig.ImagePullPolicy)

			tc.wantBackendAssertions(t, publisherName, *deployment)
		})
	}
}

func TestNewDeploymentSecurityContext(t *testing.T) {
	// given
	config := env.GetBackendConfig()
	deployment := newDeployment("test-name", "test-namespace", config.PublisherConfig, WithContainers("test-name", config.PublisherConfig))

	// when
	podSecurityContext := deployment.Spec.Template.Spec.SecurityContext
	containerSecurityContext := deployment.Spec.Template.Spec.Containers[0].SecurityContext

	// then
	assert.Equal(t, getPodSecurityContext(), podSecurityContext)
	assert.Equal(t, getContainerSecurityContext(), containerSecurityContext)
}

func Test_GetNATSEnvVars(t *testing.T) {
	testCases := []struct {
		name            string
		givenEnvs       map[string]string
		givenNATSConfig env.NATSConfig
		wantEnvs        map[string]string
	}{
		{
			name: "REQUEST_TIMEOUT should not be set and JS envs should stay empty",
			givenEnvs: map[string]string{
				"PUBLISHER_REQUESTS_CPU":    "64m",
				"PUBLISHER_REQUESTS_MEMORY": "128Mi",
				"PUBLISHER_REQUEST_TIMEOUT": "10s",
			},
			givenNATSConfig: env.NATSConfig{},
			wantEnvs: map[string]string{
				"REQUEST_TIMEOUT": "10s",
				"JS_STREAM_NAME":  "",
			},
		},
		{
			name: "Test the REQUEST_TIMEOUT and non-empty NatsConfig",
			givenEnvs: map[string]string{
				"PUBLISHER_REQUEST_TIMEOUT": "10s",
			},
			givenNATSConfig: env.NATSConfig{
				JSStreamName: "kyma",
			},
			wantEnvs: map[string]string{
				"REQUEST_TIMEOUT": "10s",
				"JS_STREAM_NAME":  "kyma",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.givenEnvs {
				t.Setenv(k, v)
			}
			backendConfig := env.GetBackendConfig()
			envVars := getNATSEnvVars(tc.givenNATSConfig, backendConfig.PublisherConfig)

			// ensure the right envs were set
			for index, val := range tc.wantEnvs {
				gotEnv := test.FindEnvVar(envVars, index)
				assert.NotNil(t, gotEnv)
				assert.Equal(t, val, gotEnv.Value)
			}
		})
	}
}
func Test_GetLogEnvVars(t *testing.T) {
	testCases := []struct {
		name      string
		givenEnvs map[string]string
		wantEnvs  map[string]string
	}{
		{
			name: "APP_LOG_FORMAT should be text and APP_LOG_LEVEL should become the default info value",
			givenEnvs: map[string]string{
				"APP_LOG_FORMAT": "text",
			},
			wantEnvs: map[string]string{
				"APP_LOG_FORMAT": "text",
				"APP_LOG_LEVEL":  "info",
			},
		},
		{
			name: "APP_LOG_FORMAT should become default json and APP_LOG_LEVEL should be warning",
			givenEnvs: map[string]string{
				"APP_LOG_LEVEL": "warning",
			},
			wantEnvs: map[string]string{
				"APP_LOG_FORMAT": "json",
				"APP_LOG_LEVEL":  "warning",
			},
		},
		{
			name:      "APP_LOG_FORMAT and APP_LOG_LEVEL should take the default values",
			givenEnvs: map[string]string{},
			wantEnvs: map[string]string{
				"APP_LOG_FORMAT": "json",
				"APP_LOG_LEVEL":  "info",
			},
		},
		{
			name: "APP_LOG_FORMAT should be testFormat and APP_LOG_LEVEL should be error",
			givenEnvs: map[string]string{
				"APP_LOG_FORMAT": "text",
				"APP_LOG_LEVEL":  "error",
			},
			wantEnvs: map[string]string{
				"APP_LOG_FORMAT": "text",
				"APP_LOG_LEVEL":  "error",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.givenEnvs {
				t.Setenv(k, v)
			}
			backendConfig := env.GetBackendConfig()
			envVars := getLogEnvVars(backendConfig.PublisherConfig)

			// ensure the right envs were set
			for index, val := range tc.wantEnvs {
				gotEnv := test.FindEnvVar(envVars, index)
				assert.NotNil(t, gotEnv)
				assert.Equal(t, val, gotEnv.Value)
			}
		})
	}
}

func Test_GetEventMeshEnvVars(t *testing.T) {
	testCases := []struct {
		name      string
		givenEnvs map[string]string
		wantEnvs  map[string]string
	}{
		{
			name: "REQUEST_TIMEOUT is not set, the default value should be taken",
			givenEnvs: map[string]string{
				"PUBLISHER_REQUESTS_CPU": "64m",
			},
			wantEnvs: map[string]string{
				"REQUEST_TIMEOUT": "5s", // default value
			},
		},
		{
			name: "REQUEST_TIMEOUT should be set",
			givenEnvs: map[string]string{
				"PUBLISHER_REQUEST_TIMEOUT": "10s",
			},
			wantEnvs: map[string]string{
				"REQUEST_TIMEOUT": "10s",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.givenEnvs {
				t.Setenv(k, v)
			}
			backendConfig := env.GetBackendConfig()
			envVars := getEventMeshEnvVars("test-name", backendConfig.PublisherConfig)

			// ensure the right envs were set
			for index, val := range tc.wantEnvs {
				gotEnv := test.FindEnvVar(envVars, index)
				assert.NotNil(t, gotEnv)
				assert.Equal(t, val, gotEnv.Value)
			}
		})
	}
}

// natsBackendAssertions checks that the NATS-specific data was set in the NewNATSPublisherDeployment.
func natsBackendAssertions(t *testing.T, publisherName string, deployment appsv1.Deployment) {
	container := findPublisherContainer(publisherName, deployment)
	assert.NotNil(t, container)

	streamName := test.FindEnvVar(container.Env, "JS_STREAM_NAME")
	assert.Equal(t, streamName.Value, "kyma")
	url := test.FindEnvVar(container.Env, "NATS_URL")
	assert.Equal(t, url.Value, natsURL)
	eventTypePrefixEnv := test.FindEnvVar(container.Env, "EVENT_TYPE_PREFIX")
	assert.Equal(t, eventTypePrefixEnv.Value, eventTypePrefix)

	// check the affinity was set
	affinityLabels := deployment.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution[0].PodAffinityTerm.LabelSelector.MatchLabels
	for _, val := range affinityLabels {
		assert.Equal(t, val, publisherName)
	}
}

// eventMeshBackendAssertions checks that the eventmesh-specific data was set in the NewEventMeshPublisherDeployment.
func eventMeshBackendAssertions(t *testing.T, publisherName string, deployment appsv1.Deployment) {
	container := findPublisherContainer(publisherName, deployment)
	assert.NotNil(t, container)

	// check eventmesh-specific env variables
	eventMeshNamespace := test.FindEnvVar(container.Env, "BEB_NAMESPACE")
	assert.Equal(t, eventMeshNamespace.Value, fmt.Sprintf("%s$(BEB_NAMESPACE_VALUE)", eventMeshNamespacePrefix))

	// check the affinity is empty
	assert.Empty(t, deployment.Spec.Template.Spec.Affinity)
}

// findPublisherContainer gets the publisher proxy container by its name.
func findPublisherContainer(publisherName string, deployment appsv1.Deployment) v1.Container {
	var container v1.Container
	for _, c := range deployment.Spec.Template.Spec.Containers {
		if strings.EqualFold(c.Name, publisherName) {
			container = c
		}
	}
	return container
}