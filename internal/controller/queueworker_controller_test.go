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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	appsv1alpha1 "github.com/mystic-06/queueworker-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("QueueWorker Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		queueworker := &appsv1alpha1.QueueWorker{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind QueueWorker")
			err := k8sClient.Get(ctx, typeNamespacedName, queueworker)
			if err != nil && errors.IsNotFound(err) {
				resource := &appsv1alpha1.QueueWorker{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: appsv1alpha1.QueueWorkerSpec{
						MinReplicas: 1,
						MaxReplicas: 10,
						TasksPerPod: 35,
						Image:       "nginx",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &appsv1alpha1.QueueWorker{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance QueueWorker")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})

		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &QueueWorkerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})

			deployment := &appsv1.Deployment{}

			depErr := k8sClient.Get(
				ctx,
				client.ObjectKey{
					Name:      resourceName + "-deployment",
					Namespace: "default",
				},
				deployment,
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(depErr).NotTo(HaveOccurred())
			Expect(deployment.Spec.Replicas).NotTo(BeNil())
			Expect(*deployment.Spec.Replicas).To(Equal(int32(3)))
		})

		It("should update Deployment when QueueWorker spec changes", func() {
			ctx := context.Background()

			qw := &appsv1alpha1.QueueWorker{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "update-test",
					Namespace: "default",
				},
				Spec: appsv1alpha1.QueueWorkerSpec{
					MinReplicas: 1,
					MaxReplicas: 10,
					TasksPerPod: 50,
					Image:       "nginx",
				},
			}

			Expect(k8sClient.Create(ctx, qw)).To(Succeed())

			reconciler := &QueueWorkerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			//During first reconcile
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "update-test",
					Namespace: "default",
				},
			})

			Expect(err).NotTo(HaveOccurred())

			deployment := &appsv1.Deployment{}

			Expect(k8sClient.Get(
				ctx,
				client.ObjectKey{
					Name:      "update-test-deployment",
					Namespace: "default",
				},
				deployment,
			)).To(Succeed())

			//queueDepth is hardcoded right now

			Expect(*deployment.Spec.Replicas).To(Equal(int32(3)))

			Expect(k8sClient.Get(
				ctx,
				types.NamespacedName{
					Name:      "update-test",
					Namespace: "default",
				},
				qw,
			)).To(Succeed())

			qw.Spec.TasksPerPod = 25

			Expect(k8sClient.Update(ctx, qw)).To(Succeed())

			// Reconcile again
			_, err = reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      "update-test",
					Namespace: "default",
				},
			})

			Expect(err).NotTo(HaveOccurred())

			// Fetch deployment again
			Expect(k8sClient.Get(
				ctx,
				client.ObjectKey{
					Name:      "update-test-deployment",
					Namespace: "default",
				},
				deployment,
			)).To(Succeed())

			Expect(*deployment.Spec.Replicas).To(Equal(int32(5)))
		})

		It("should handle QueueWorker deletion gracefully", func() {
			resource := &appsv1alpha1.QueueWorker{}

			Expect(k8sClient.Get(
				ctx,
				typeNamespacedName,
				resource,
			)).To(Succeed())

			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())

			reconciler := &QueueWorkerReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      queueworker.Name,
					Namespace: queueworker.Namespace,
				},
			})

			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func TestCalculateReplicas(t *testing.T) {
	tests := []struct {
		name string

		queueDepth  int
		tasksPerPod int32
		minReplicas int32
		maxReplicas int32

		expectedVal int32
	}{
		{
			name:        "normal scaling",
			queueDepth:  101,
			tasksPerPod: 50,
			minReplicas: 1,
			maxReplicas: 10,
			expectedVal: 3,
		},
		{
			name:        "minimum replicas enforced",
			queueDepth:  0,
			tasksPerPod: 50,
			minReplicas: 1,
			maxReplicas: 10,
			expectedVal: 1,
		},
		{
			name:        "maximum replicas enforced",
			queueDepth:  999,
			tasksPerPod: 50,
			minReplicas: 1,
			maxReplicas: 10,
			expectedVal: 10,
		},
		{
			name:        "exact division",
			queueDepth:  100,
			tasksPerPod: 50,
			minReplicas: 1,
			maxReplicas: 10,
			expectedVal: 2,
		},
		{
			name:        "ceiling behavior",
			queueDepth:  51,
			tasksPerPod: 50,
			minReplicas: 1,
			maxReplicas: 10,
			expectedVal: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateReplicas(tt.queueDepth, tt.tasksPerPod, tt.minReplicas, tt.maxReplicas)

			if got != tt.expectedVal {
				t.Errorf("got %d wanted %d", got, tt.expectedVal)
			}
		})
	}
}
