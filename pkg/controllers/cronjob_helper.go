/*
Copyright 2020 Backup Operator Authors

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
	"path/filepath"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	WorkerContainerName    = "worker"
	WorkerConfigVolumeName = "config"
	WorkerConfigMountPath  = "/etc/worker"
)

var (
	WorkerConfigFilePath = filepath.Join(WorkerConfigMountPath, "plan.json")
)

func UpdateCronJobSpec(cronJob *batchv1.CronJob, secretRef *corev1.ObjectReference, schedule string, activeDeadlineSeconds int64, image string, env []corev1.EnvVar, subcmd string,
	volumes []corev1.Volume,
	volumeMounts []corev1.VolumeMount) error {
	cronJob.Spec.Schedule = schedule
	jobSpec := &cronJob.Spec.JobTemplate.Spec
	jobSpec.ActiveDeadlineSeconds = &activeDeadlineSeconds
	podSpec := &jobSpec.Template.Spec

	podSpec.Volumes = append(volumes, corev1.Volume{
		Name: WorkerConfigVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secretRef.Name,
			},
		},
	},
	)

	podSpec.Containers = []corev1.Container{
		{
			Name:            WorkerContainerName,
			Image:           image,
			ImagePullPolicy: corev1.PullIfNotPresent, // NOTE: Currently required for tests!
			Env:             env,
			Command:         []string{"/worker"},
			Args:            []string{subcmd, WorkerConfigFilePath},
			VolumeMounts: append(volumeMounts,
				corev1.VolumeMount{
					Name:      WorkerConfigVolumeName,
					MountPath: WorkerConfigMountPath,
					ReadOnly:  true,
				}),
		},
	}
	podSpec.RestartPolicy = corev1.RestartPolicyOnFailure
	return nil
}
