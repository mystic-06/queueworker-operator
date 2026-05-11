/*
Copyright 2026.

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

package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/mystic-06/queueworker-operator/api/v1alpha1"
	appsv1alpha1 "github.com/mystic-06/queueworker-operator/api/v1alpha1"
	"github.com/rs/zerolog/log"
)

// QueueWorkerReconciler reconciles a QueueWorker object
type QueueWorkerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=apps.mystic-06.github.io,resources=queueworkers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.mystic-06.github.io,resources=queueworkers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.mystic-06.github.io,resources=queueworkers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the QueueWorker object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *QueueWorkerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	qworker := &v1alpha1.QueueWorker{}

	//Fetch queueworker instance
	if err := r.Get(ctx, req.NamespacedName, qworker); err != nil {
		//Check if resource is not found
		if apierrors.IsNotFound(err) {
			//The resource was deleted
			log.Info().Msg("QueueWorker resource not found.")
			return ctrl.Result{}, nil
		}
		//If it's a different error, return the error
		log.Error().Msg("Unable to fetch resource QueueWorker")
		return ctrl.Result{}, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      qworker.Name + "-deployment",
			Namespace: qworker.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment,
		func() error {
			//TO DO: Add logic here
			return nil
		},
	)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QueueWorkerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.QueueWorker{}).
		Named("queueworker").
		Complete(r)
}
