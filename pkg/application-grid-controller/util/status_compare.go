package util

import appsv1 "k8s.io/api/apps/v1"

func FiltDgStatus(status appsv1.DeploymentStatus) appsv1.DeploymentStatus {
	tempstatus := appsv1.DeploymentStatus{}
	tempstatus.Replicas = status.Replicas
	tempstatus.ReadyReplicas = status.ReadyReplicas
	for _, v := range status.Conditions {
		if v.Type == appsv1.DeploymentAvailable {
			tempstatus.Conditions = []appsv1.DeploymentCondition{v}
			break
		}
	}
	return tempstatus
}
