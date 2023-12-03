/*
Copyright 2023 David Lynch.

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

// ExecutionSpec defines the desired state of Execution
type ExecutionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Reference to the pipeline which will be executed
	Pipeline string `json:"pipeline"`
	// Git repo hosting the code to be tested against pipeline
	Repo string `json:"repo"`
	// Git branch
	Branch   string `json:"branch"`
	CloneDir string `json:"cloneDir"`
}

// ExecutionStatus defines the observed state of Execution
type ExecutionStatus struct {
	// Shows if the PV for this execution has been provisioned
	VolumeProvisioned bool `json:"volumeProvisioned,omitempty" default:"false"`
	// Tells the controller if the repo is cloned
	RepoCloned bool `json:"repoCloned,omitempty" default:"false"`

	// Describes which tasks are currently running
	Executing []string `json:"executing,omitempty"`
	// States if the pipeline is completed
	Completed bool `json:"completed,omitempty" default:"false"`
	// Tasks which have already completed
	CompletedTasks []string `json:"completedTasks,omitempty"`
	// States if the pipeline has failed at any point
	Error bool `json:"error" default:"false"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Execution is the Schema for the executions API
type Execution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExecutionSpec   `json:"spec,omitempty"`
	Status ExecutionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExecutionList contains a list of Execution
type ExecutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Execution `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Execution{}, &ExecutionList{})
}