/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package eventing

import (
	"context"
	"fmt"

	eventingv1alpha1 "github.com/kyma-project/eventing-manager/api/v1alpha1"
	"github.com/kyma-project/eventing-manager/pkg/k8s"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"go.uber.org/zap"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/kyma-project/eventing-manager/pkg/eventing"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/util/retry"
)

const (
	FinalizerName             = "eventing.operator.kyma-project.io/finalizer"
	ControllerName            = "eventing-manager-controller"
	ManagedByLabelKey         = "app.kubernetes.io/managed-by"
	ManagedByLabelValue       = ControllerName
	NatsServerNotAvailableMsg = "NATS server is not available in namespace %s"
)

// Reconciler reconciles a Eventing object
type Reconciler struct {
	client.Client
	logger          *logger.Logger
	ctrlManager     ctrl.Manager
	eventingManager eventing.Manager
	kubeClient      k8s.Client
}

func NewReconciler(
	ctx context.Context,
	natsConfig env.NATSConfig,
	client client.Client,
	scheme *runtime.Scheme,
	logger *logger.Logger,
	recorder record.EventRecorder,
) *Reconciler {
	return &Reconciler{
		Client:          client,
		logger:          logger,
		eventingManager: eventing.NewEventingManager(ctx, client, natsConfig, logger, recorder),
		kubeClient:      k8s.NewKubeClient(client),
	}
}

//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=eventings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=eventings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=eventings/finalizers,verbs=update
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;delete;patch
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=nats,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.namedLogger().Info("Reconciliation triggered")
	// fetch latest subscription object
	currentEventing := &eventingv1alpha1.Eventing{}
	if err := r.Get(ctx, req.NamespacedName, currentEventing); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// copy the object, so we don't modify the source object
	eventing := currentEventing.DeepCopy()

	// logger with eventing details
	log := r.loggerWithEventing(eventing)

	// check if eventing is in deletion state
	if !eventing.DeletionTimestamp.IsZero() {
		return r.handleEventingDeletion(ctx, eventing, log)
	}

	// handle reconciliation
	return r.handleEventingReconcile(ctx, eventing, log)
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.ctrlManager = mgr
	return ctrl.NewControllerManagedBy(mgr).
		For(&eventingv1alpha1.Eventing{}).
		Owns(&v1.Deployment{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Complete(r)
}

// loggerWithEventing returns a logger with the given Eventing CR details.
func (r *Reconciler) loggerWithEventing(eventing *eventingv1alpha1.Eventing) *zap.SugaredLogger {
	return r.namedLogger().With(
		"kind", eventing.GetObjectKind().GroupVersionKind().Kind,
		"resourceVersion", eventing.GetResourceVersion(),
		"generation", eventing.GetGeneration(),
		"namespace", eventing.GetNamespace(),
		"name", eventing.GetName(),
	)
}

func (r *Reconciler) handleEventingDeletion(_ context.Context, _ *eventingv1alpha1.Eventing,
	log *zap.SugaredLogger) (ctrl.Result, error) {
	log.Info("handling Eventing deletion...")
	// TODO: Implement me.
	return ctrl.Result{}, nil
}

func (r *Reconciler) handleEventingReconcile(ctx context.Context,
	eventing *eventingv1alpha1.Eventing, log *zap.SugaredLogger) (ctrl.Result, error) {
	log.Info("handling Eventing reconciliation...")

	// set state processing if not set yet
	r.InitStateProcessing(eventing)

	for _, backend := range eventing.Spec.Backends {
		switch backend.Type {
		case eventingv1alpha1.NatsBackendType:
			return r.reconcileNATSBackend(ctx, eventing, log)
		case eventingv1alpha1.EventMeshBackendType:
			return r.reconcileEventMeshBackend(ctx, eventing, log)
		default:
			return ctrl.Result{Requeue: false}, fmt.Errorf("not supported backend type %s", backend.Type)
		}
	}
	// this should never happen, but if happens do nothing
	return ctrl.Result{Requeue: false}, fmt.Errorf("no backend is provided in the spec")
}

func (r *Reconciler) reconcileNATSBackend(ctx context.Context, eventing *eventingv1alpha1.Eventing, log *zap.SugaredLogger) (ctrl.Result, error) {
	// check nats CR if it exists and is in natsAvailable state
	err := r.checkNATSAvailability(ctx, eventing)
	if err != nil {
		return ctrl.Result{}, r.syncStatusWithNATSErr(ctx, eventing, err, log)
	}

	// set NATSAvailable condition to true and update status
	eventing.Status.SetNATSAvailableConditionToTrue()
	if err := r.syncEventingStatus(ctx, eventing, log); err != nil {
		return ctrl.Result{}, err
	}

	deployment, err := r.handlePublisherProxy(ctx, eventing, log)
	if err != nil {
		return ctrl.Result{}, r.syncStatusWithPublisherProxyErr(ctx, eventing, err, log)
	}

	return r.handleEventingState(ctx, deployment, eventing, log)
}

func (r *Reconciler) checkNATSAvailability(ctx context.Context, eventing *eventingv1alpha1.Eventing) error {
	natsAvailable, err := r.eventingManager.IsNATSAvailable(ctx, eventing.Namespace)
	if err != nil {
		return err
	}
	if !natsAvailable {
		return fmt.Errorf(NatsServerNotAvailableMsg, eventing.Namespace)
	}
	return nil
}

func (r *Reconciler) handlePublisherProxy(ctx context.Context, eventing *eventingv1alpha1.Eventing,
	log *zap.SugaredLogger) (*v1.Deployment, error) {
	// CreateOrUpdate deployment for eventing publisher proxy deployment
	deployment, err := r.eventingManager.CreateOrUpdatePublisherProxy(ctx, eventing)
	if err != nil {
		return nil, err
	}

	// TODO: remove other owner references and set the following owner reference
	// Overwrite owner reference for publisher proxy deployment as the EC sets its deployment as owner
	// and we want the Eventing CR to be the owner.
	err = r.setDeploymentOwnerReference(ctx, deployment, eventing)
	if err != nil {
		return deployment, fmt.Errorf("failed to set owner reference for publisher proxy deployment: %s", err)
	}

	// CreateOrUpdate HPA for publisher proxy deployment
	err = r.eventingManager.CreateOrUpdateHPA(ctx, deployment, eventing, 60, 60)
	if err != nil {
		return deployment, fmt.Errorf("failed to create or update HPA for publisher proxy deployment: %s", err)
	}

	return deployment, nil
}

func (r *Reconciler) reconcileEventMeshBackend(ctx context.Context, eventing *eventingv1alpha1.Eventing, log *zap.SugaredLogger) (ctrl.Result, error) {
	// TODO: Implement me.
	return ctrl.Result{}, nil
}

func (r *Reconciler) setDeploymentOwnerReference(ctx context.Context, deployment *v1.Deployment, eventing *eventingv1alpha1.Eventing) error {
	// Update the deployment object
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, deployment)
		if err != nil {
			return err
		}
		// Set the controller reference to the parent object
		if err := controllerutil.SetControllerReference(eventing, deployment, r.Scheme()); err != nil {
			return fmt.Errorf("failed to set controller reference: %v", err)
		}
		return r.Update(ctx, deployment)
	})
	return retryErr
}

func (r *Reconciler) namedLogger() *zap.SugaredLogger {
	return r.logger.WithContext().Named(ControllerName)
}