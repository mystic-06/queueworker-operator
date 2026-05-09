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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// QueueWorkerSpec defines the desired state of QueueWorker
type QueueWorkerSpec struct {
	//URL of queue to monitor
	QueueURL string `json:"queueURL"`

	// +kubebuilder:validation:Minimum=1
	// Minimum no. of worker pods
	MinReplicas int32 `json:"minReplicas"`

	// +kubebuilder:validation:Minimum=1
	// Maximum no. of worker pods
	MaxReplicas int32 `json:"maxReplicas"`

	// Defines how many tasks in the queue can be occupied by one worker pod.
	// +kubebuilder:validation:Minimum=1
	TasksPerPod int32 `json:"tasksPerPod"`

	// Image is the Docker image for the worker pods
	Image string `json:"image"`
}

// QueueWorkerStatus defines the observed state of QueueWorker.
type QueueWorkerStatus struct {
	// CurrentReplicas is the actual number of pods currently running
	CurrentReplicas int32 `json:"currentReplicas"`

	// LastScaleTime is the timestamp of the last scaling event
	LastScaleTime *metav1.Time `json:"lastScaleTime,omitempty"`

	// Conditions represent the latest available observations of the state
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// QueueWorker is the Schema for the queueworkers API
type QueueWorker struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of QueueWorker
	// +required
	Spec QueueWorkerSpec `json:"spec"`

	// status defines the observed state of QueueWorker
	// +optional
	Status QueueWorkerStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// QueueWorkerList contains a list of QueueWorker
type QueueWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []QueueWorker `json:"items"`
}

func init() {
	SchemeBuilder.Register(&QueueWorker{}, &QueueWorkerList{})
}
