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
	"math"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/mystic-06/queueworker-operator/api/v1alpha1"
	appsv1alpha1 "github.com/mystic-06/queueworker-operator/api/v1alpha1"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// QueueWorkerReconciler reconciles a QueueWorker object
type QueueWorkerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func calculateReplicas(queueDepth int, tasksPerPod int32, minReplicas int32, maxReplicas int32) int32 {
	replicas := math.Ceil(float64(queueDepth) / float64(tasksPerPod))

	if replicas < float64(minReplicas) {
		return minReplicas
	} else if replicas > float64(maxReplicas) {
		return maxReplicas
	}

	return int32(replicas)
}

// +kubebuilder:rbac:groups=apps.mystic-06.github.io,resources=queueworkers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.mystic-06.github.io,resources=queueworkers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.mystic-06.github.io,resources=queueworkers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
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
		log.Error().Err(err).Msg("Unable to fetch QueueWorker")
		return ctrl.Result{}, err
	}

	queueDepth := 101

	desiredReplicas := calculateReplicas(
		queueDepth,
		qworker.Spec.TasksPerPod,
		qworker.Spec.MinReplicas,
		qworker.Spec.MaxReplicas,
	)

	log.Info().
		Int32("desiredReplicas", desiredReplicas).
		Int32("qworker.Spec.TasksPerPod", qworker.Spec.TasksPerPod).
		Msg("Calculated desired replicas")

	//Create or Update the deployment
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      qworker.Name + "-deployment",
			Namespace: qworker.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment,
		func() error {
			// var replicas int32 = qworker.Spec.MinReplicas

			deployment.Spec.Replicas = &desiredReplicas

			labels := map[string]string{
				"app": qworker.Name,
			}

			deployment.Spec.Selector = &metav1.LabelSelector{
				MatchLabels: labels,
			}

			//Update the Pod deployment
			deployment.Spec.Template = corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "worker",
							Image: qworker.Spec.Image,
						},
					},
				},
			}

			return controllerutil.SetControllerReference(
				qworker,
				deployment,
				r.Scheme,
			)
		},
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to reconcile Deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *QueueWorkerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.QueueWorker{}).
		Named("queueworker").
		Complete(r)
}
