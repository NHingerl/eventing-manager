package test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/avast/retry-go/v3"
	"github.com/go-logr/zapr"
	apigatewayv1beta1 "github.com/kyma-project/api-gateway/apis/gateway/v1beta1"
	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	"github.com/stretchr/testify/require"
	kcorev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	klabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	kctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	kctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"

	eventingv1alpha2 "github.com/kyma-project/eventing-manager/api/eventing/v1alpha2"
	subscriptioncontrollereventmesh "github.com/kyma-project/eventing-manager/internal/controller/eventing/subscription/eventmesh"
	"github.com/kyma-project/eventing-manager/internal/controller/eventing/subscription/validator"
	"github.com/kyma-project/eventing-manager/pkg/backend/cleaner"
	backendeventmesh "github.com/kyma-project/eventing-manager/pkg/backend/eventmesh"
	"github.com/kyma-project/eventing-manager/pkg/backend/metrics"
	backendutils "github.com/kyma-project/eventing-manager/pkg/backend/utils"
	"github.com/kyma-project/eventing-manager/pkg/constants"
	emstypes "github.com/kyma-project/eventing-manager/pkg/ems/api/events/types"
	"github.com/kyma-project/eventing-manager/pkg/env"
	"github.com/kyma-project/eventing-manager/pkg/featureflags"
	"github.com/kyma-project/eventing-manager/pkg/logger"
	"github.com/kyma-project/eventing-manager/pkg/utils"
	testutils "github.com/kyma-project/eventing-manager/test/utils"
	eventingtesting "github.com/kyma-project/eventing-manager/testing"
)

type eventMeshTestEnsemble struct {
	k8sClient     client.Client
	testEnv       *envtest.Environment
	eventMeshMock *eventingtesting.EventMeshMock
	nameMapper    backendutils.NameMapper
	envConfig     env.Config
}

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
	testEnvStartDelay        = time.Minute
	testEnvStartAttempts     = 10
	twoMinTimeOut            = 120 * time.Second
	bigPollingInterval       = 3 * time.Second
	bigTimeOut               = 40 * time.Second
	smallTimeOut             = 5 * time.Second
	smallPollingInterval     = 1 * time.Second
	namespacePrefixLength    = 5
	syncPeriodSeconds        = 2
	maxReconnects            = 10
	eventMeshMockKeyPrefix   = "/messaging/events/subscriptions"
	certsURL                 = "https://domain.com/oauth2/certs"
)

//nolint:gochecknoglobals // only used in tests
var (
	emTestEnsemble    *eventMeshTestEnsemble
	k8sCancelFn       context.CancelFunc
	acceptableMethods = []string{http.MethodPost, http.MethodOptions}
)

func setupSuite() error {
	featureflags.SetEventingWebhookAuthEnabled(true)
	emTestEnsemble = &eventMeshTestEnsemble{}

	// define logger
	defaultLogger, err := logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	if err != nil {
		return err
	}
	kctrllog.SetLogger(zapr.NewLogger(defaultLogger.WithContext().Desugar()))

	// setup test Env
	cfg, err := startTestEnv()
	if err != nil || cfg == nil {
		return err
	}

	// start event mesh mock
	emTestEnsemble.eventMeshMock = startNewEventMeshMock()

	// add schemes
	if err = eventingv1alpha2.AddToScheme(scheme.Scheme); err != nil {
		return err
	}

	if err = apigatewayv1beta1.AddToScheme(scheme.Scheme); err != nil {
		return err
	}
	// +kubebuilder:scaffold:scheme

	// setup eventMesh manager instance
	k8sManager, err := setupManager(cfg)
	if err != nil {
		return err
	}

	// setup nameMapper for EventMesh
	emTestEnsemble.nameMapper = backendutils.NewBEBSubscriptionNameMapper(testutils.Domain,
		backendeventmesh.MaxSubscriptionNameLength)

	// setup eventMesh reconciler
	recorder := k8sManager.GetEventRecorderFor("eventing-controller")
	emTestEnsemble.envConfig = getEnvConfig()

	eventMesh, credentials := setupEventMesh(defaultLogger)

	// Init the Subscription validator.
	subscriptionValidator := validator.NewSubscriptionValidator(k8sManager.GetClient())

	col := metrics.NewCollector()
	testReconciler := subscriptioncontrollereventmesh.NewReconciler(
		k8sManager.GetClient(),
		defaultLogger,
		recorder,
		emTestEnsemble.envConfig,
		cleaner.NewEventMeshCleaner(defaultLogger),
		eventMesh,
		credentials,
		emTestEnsemble.nameMapper,
		subscriptionValidator,
		col,
		testutils.Domain,
	)

	if err = testReconciler.SetupUnmanaged(context.Background(), k8sManager); err != nil {
		return err
	}

	// start k8s client
	go func() {
		var ctx context.Context
		ctx, k8sCancelFn = context.WithCancel(kctrl.SetupSignalHandler())
		err = k8sManager.Start(ctx)
		if err != nil {
			panic(err)
		}
	}()

	emTestEnsemble.k8sClient = k8sManager.GetClient()

	return nil
}

func setupManager(cfg *rest.Config) (manager.Manager, error) {
	syncPeriod := syncPeriodSeconds * time.Second
	opts := kctrl.Options{
		Cache:                  cache.Options{SyncPeriod: &syncPeriod},
		HealthProbeBindAddress: "0", // disable
		Scheme:                 scheme.Scheme,
		Metrics:                server.Options{BindAddress: "0"}, // disable
	}
	k8sManager, err := kctrl.NewManager(cfg, opts)
	if err != nil {
		return nil, err
	}
	return k8sManager, nil
}

func setupEventMesh(defaultLogger *logger.Logger) (*backendeventmesh.EventMesh, *backendeventmesh.OAuth2ClientCredentials) {
	credentials := &backendeventmesh.OAuth2ClientCredentials{
		ClientID:     "foo-client-id",
		ClientSecret: "foo-client-secret",
		TokenURL:     "foo-token-url",
		CertsURL:     certsURL,
	}
	eventMesh := backendeventmesh.NewEventMesh(credentials, emTestEnsemble.nameMapper, defaultLogger)
	return eventMesh, credentials
}

func startTestEnv() (*rest.Config, error) {
	useExistingCluster := useExistingCluster
	emTestEnsemble.testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../../../../../", "config", "crd", "bases"),
			filepath.Join("../../../../../../", "config", "crd", "for-tests"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var cfg *rest.Config
	err := retry.Do(func() error {
		defer func() {
			if r := recover(); r != nil {
				log.Println("panic recovered:", r)
			}
		}()

		cfgLocal, startErr := emTestEnsemble.testEnv.Start()
		cfg = cfgLocal
		return startErr
	},
		retry.Delay(testEnvStartDelay),
		retry.DelayType(retry.FixedDelay),
		retry.Attempts(testEnvStartAttempts),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("[%v] try failed to start testenv: %s", n, err)
			if stopErr := emTestEnsemble.testEnv.Stop(); stopErr != nil {
				log.Printf("failed to stop testenv: %s", stopErr)
			}
		}),
	)
	return cfg, err
}

func getEnvConfig() env.Config {
	return env.Config{
		BEBAPIURL:                emTestEnsemble.eventMeshMock.MessagingURL,
		ClientID:                 "foo-id",
		ClientSecret:             "foo-secret",
		TokenEndpoint:            emTestEnsemble.eventMeshMock.TokenURL,
		WebhookActivationTimeout: 0,
		EventTypePrefix:          eventingtesting.EventMeshPrefix,
		BEBNamespace:             eventingtesting.EventMeshNamespaceNS,
		Qos:                      string(emstypes.QosAtLeastOnce),
	}
}

func tearDownSuite() error {
	if k8sCancelFn != nil {
		k8sCancelFn()
	}
	err := emTestEnsemble.testEnv.Stop()
	emTestEnsemble.eventMeshMock.Stop()
	return err
}

func startNewEventMeshMock() *eventingtesting.EventMeshMock {
	emMock := eventingtesting.NewEventMeshMock()
	emMock.Start()
	return emMock
}

func getTestNamespace() string {
	return fmt.Sprintf("ns-%s", utils.GetRandString(namespacePrefixLength))
}

func ensureNamespaceCreated(ctx context.Context, t *testing.T, namespace string) {
	t.Helper()
	if namespace == "default" {
		return
	}
	// create namespace
	ns := fixtureNamespace(namespace)
	err := emTestEnsemble.k8sClient.Create(ctx, ns)
	if !kerrors.IsAlreadyExists(err) {
		require.NoError(t, err)
	}
}

func fixtureNamespace(name string) *kcorev1.Namespace {
	namespace := kcorev1.Namespace{
		TypeMeta: kmetav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: kmetav1.ObjectMeta{
			Name: name,
		},
	}
	return &namespace
}

func ensureK8sResourceCreated(ctx context.Context, t *testing.T, obj client.Object) {
	t.Helper()
	require.NoError(t, emTestEnsemble.k8sClient.Create(ctx, obj))
}

func ensureK8sResourceDeleted(ctx context.Context, t *testing.T, obj client.Object) {
	t.Helper()
	require.NoError(t, emTestEnsemble.k8sClient.Delete(ctx, obj))
}

func ensureK8sSubscriptionUpdated(ctx context.Context, t *testing.T, subscription *eventingv1alpha2.Subscription) {
	t.Helper()
	require.Eventually(t, func() bool {
		latestSubscription := &eventingv1alpha2.Subscription{}
		lookupKey := types.NamespacedName{
			Namespace: subscription.Namespace,
			Name:      subscription.Name,
		}
		require.NoError(t, emTestEnsemble.k8sClient.Get(ctx, lookupKey, latestSubscription))
		require.NotEmpty(t, latestSubscription.Name)
		latestSubscription.Spec = subscription.Spec
		latestSubscription.Labels = subscription.Labels
		require.NoError(t, emTestEnsemble.k8sClient.Update(ctx, latestSubscription))
		return true
	}, bigTimeOut, bigPollingInterval)
}

// ensureAPIRuleStatusUpdatedWithStatusReady updates the status fof the APIRule (mocking APIGateway controller).
func ensureAPIRuleStatusUpdatedWithStatusReady(ctx context.Context, t *testing.T, apiRule *apigatewayv1beta1.APIRule) {
	t.Helper()
	require.Eventually(t, func() bool {
		fetchedAPIRule, err := getAPIRule(ctx, apiRule)
		if err != nil {
			return false
		}

		newAPIRule := fetchedAPIRule.DeepCopy()
		// mark the ApiRule status as ready
		eventingtesting.MarkReady(newAPIRule)

		// update ApiRule status on k8s
		err = emTestEnsemble.k8sClient.Status().Update(ctx, newAPIRule)
		return err == nil
	}, bigTimeOut, bigPollingInterval)
}

// ensureAPIRuleNotFound ensures that a APIRule does not exists (or deleted).
func ensureAPIRuleNotFound(ctx context.Context, t *testing.T, apiRule *apigatewayv1beta1.APIRule) {
	t.Helper()
	require.Eventually(t, func() bool {
		apiRuleKey := client.ObjectKey{
			Namespace: apiRule.Namespace,
			Name:      apiRule.Name,
		}

		apiRule2 := new(apigatewayv1beta1.APIRule)
		err := emTestEnsemble.k8sClient.Get(ctx, apiRuleKey, apiRule2)
		return kerrors.IsNotFound(err)
	}, bigTimeOut, bigPollingInterval)
}

func getAPIRulesList(ctx context.Context, svc *kcorev1.Service) (*apigatewayv1beta1.APIRuleList, error) {
	labels := map[string]string{
		constants.ControllerServiceLabelKey:  svc.Name,
		constants.ControllerIdentityLabelKey: constants.ControllerIdentityLabelValue,
	}
	apiRules := &apigatewayv1beta1.APIRuleList{}
	err := emTestEnsemble.k8sClient.List(ctx, apiRules, &client.ListOptions{
		LabelSelector: klabels.SelectorFromSet(labels),
		Namespace:     svc.Namespace,
	})
	return apiRules, err
}

func getAPIRule(ctx context.Context, apiRule *apigatewayv1beta1.APIRule) (*apigatewayv1beta1.APIRule, error) {
	lookUpKey := types.NamespacedName{
		Namespace: apiRule.Namespace,
		Name:      apiRule.Name,
	}
	err := emTestEnsemble.k8sClient.Get(ctx, lookUpKey, apiRule)
	return apiRule, err
}

func filterAPIRulesForASvc(apiRules *apigatewayv1beta1.APIRuleList, svc *kcorev1.Service) apigatewayv1beta1.APIRule {
	if len(apiRules.Items) == 1 && *apiRules.Items[0].Spec.Service.Name == svc.Name {
		return apiRules.Items[0]
	}
	return apigatewayv1beta1.APIRule{}
}

// countEventMeshRequests returns how many requests for a given subscription are sent for each HTTP method
//

func countEventMeshRequests(subscriptionName, eventType string) (int, int, int) {
	countGet, countPost, countDelete := 0, 0, 0
	emTestEnsemble.eventMeshMock.Requests.ReadEach(
		func(request *http.Request, payload interface{}) {
			switch method := request.Method; method {
			case http.MethodGet:
				if strings.Contains(request.URL.Path, subscriptionName) {
					countGet++
				}
			case http.MethodPost:
				if sub, ok := payload.(emstypes.Subscription); ok {
					if len(sub.Events) > 0 {
						for _, event := range sub.Events {
							if event.Type == eventType && sub.Name == subscriptionName {
								countPost++
							}
						}
					}
				}
			case http.MethodDelete:
				if strings.Contains(request.URL.Path, subscriptionName) {
					countDelete++
				}
			}
		})
	return countGet, countPost, countDelete
}

func getEventMeshSubFromMock(subscriptionName, subscriptionNamespace string) *emstypes.Subscription {
	key := getEventMeshSubKeyForMock(subscriptionName, subscriptionNamespace)
	return emTestEnsemble.eventMeshMock.Subscriptions.GetSubscription(key)
}

func getEventMeshSubKeyForMock(subscriptionName, subscriptionNamespace string) string {
	nm1 := emTestEnsemble.nameMapper.MapSubscriptionName(subscriptionName, subscriptionNamespace)
	return fmt.Sprintf("%s/%s", eventMeshMockKeyPrefix, nm1)
}

func getEventMeshKeyForMock(name string) string {
	return fmt.Sprintf("%s/%s", eventMeshMockKeyPrefix, name)
}

// ensureK8sEventReceived checks if a certain event have triggered for the given namespace.
func ensureK8sEventReceived(t *testing.T, event kcorev1.Event, namespace string) {
	t.Helper()
	ctx := context.TODO()
	require.Eventually(t, func() bool {
		// get all events from k8s for namespace
		eventList := &kcorev1.EventList{}
		err := emTestEnsemble.k8sClient.List(ctx, eventList, client.InNamespace(namespace))
		require.NoError(t, err)

		// find the desired event
		var receivedEvent *kcorev1.Event
		for i, e := range eventList.Items {
			if e.Reason == event.Reason {
				receivedEvent = &eventList.Items[i]
				break
			}
		}

		// check the received event
		require.NotNil(t, receivedEvent)
		require.Equal(t, receivedEvent.Reason, event.Reason)
		require.Equal(t, receivedEvent.Message, event.Message)
		require.Equal(t, receivedEvent.Type, event.Type)
		return true
	}, bigTimeOut, bigPollingInterval)
}
