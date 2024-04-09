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

package controllers

import (
	"context"
	"fmt"
	"path/filepath"

	pipelinesv1alpha1 "github.com/davidlynch-sd/bramble/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	sourceRoot = "/src/"
)

func initExecution(
	ctx context.Context,
	r *ExecutionReconciler,
	execution *pipelinesv1alpha1.Execution,
	pvc *corev1.PersistentVolumeClaim,
) error {
	logger := log.Log.WithName(fmt.Sprintf("Execution: %v", execution.ObjectMeta.Name))

	executionSourcePath := filepath.Join(sourceRoot, execution.ObjectMeta.Name)

	if !execution.Status.VolumeProvisioned {
		logger.Info("Provisioning PV")

		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:   fmt.Sprintf("%v-pv", execution.ObjectMeta.Name),
				Labels: map[string]string{"bramble-execution": execution.ObjectMeta.Name},
			},
			Spec: corev1.PersistentVolumeSpec{
				Capacity: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName: "standard",
				PersistentVolumeSource: corev1.PersistentVolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: executionSourcePath,
					},
				},
			},
		}

		err := r.Create(ctx, pv)
		if err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("PV %v created", execution.ObjectMeta.Name))

		logger.Info("Provisioning PVC")

		pvc = &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%v-pvc", pv.ObjectMeta.Name),
				Namespace: execution.ObjectMeta.Namespace,
				Labels:    map[string]string{"bramble-execution": execution.ObjectMeta.Name},
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceStorage: resource.MustParse("1Gi"),
					},
				},
			},
		}

		err = r.Create(ctx, pvc)
		if err != nil {
			return err
		}

		logger.Info("Provisioning Cloner Pod")
	}

	if !execution.Status.RepoCloned {
		cloneJob := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%v-cloner", execution.ObjectMeta.Name),
				Namespace: execution.ObjectMeta.Namespace,
				Labels:    map[string]string{"bramble-execution": execution.ObjectMeta.Name},
			}, Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						RestartPolicy: corev1.RestartPolicyOnFailure,
						Containers: []corev1.Container{
							{
								Name:  "cloner",
								Image: "alpine/git",
								Command: []string{
									"sh", "-c",
									fmt.Sprintf("rm -rf %v && git clone %v --branch=%v %v",
										execution.ObjectMeta.Name,
										execution.Spec.Repo,
										execution.Spec.Branch,
										execution.ObjectMeta.Name,
									),
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name:      "cloner-volume",
										MountPath: executionSourcePath,
									},
								},
								WorkingDir: executionSourcePath,
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "cloner-volume",
								VolumeSource: corev1.VolumeSource{
									PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
										ClaimName: pvc.ObjectMeta.Name,
									},
								},
							},
						},
					},
				},
			},
		}

		err := r.Create(ctx, cloneJob)
		if err != nil {
			return err
		}

		logger.Info("Cloner Pod provisioned")
	}

	return nil
}
