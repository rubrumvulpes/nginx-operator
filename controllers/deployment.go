package controllers

import (
	"context"
	webserverv1 "nginx-operator/api/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (reconciler *NginxReconciler) reconcileDeployment(ctx context.Context, controller *webserverv1.Nginx) error {

	logger := log.FromContext(ctx)
	logger.Info("reconciling Deployment")

	deploymentManifest := reconciler.createDeploymentManifest(controller)
	deploymentInstance := &appsv1.Deployment{}

	err := reconciler.Get(ctx, types.NamespacedName{
		Name:      deploymentManifest.Name,
		Namespace: controller.Namespace,
	}, deploymentInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			err := reconciler.Create(ctx, deploymentManifest)
			if err != nil {
				logger.Error(err, "no Deployment found, creation failed")
				return err
			}
			logger.Info("no Deployment found, created")
			return nil
		} else {
			logger.Error(err, "error while looking up Deployment")
			return err
		}
	}

	if controller.Spec.Image != deploymentInstance.Spec.Template.Spec.Containers[0].Image {
		deploymentInstance.Spec.Template.Spec.Containers[0].Image = controller.Spec.Image
	}
	if *controller.Spec.Replicas != *deploymentInstance.Spec.Replicas {
		*deploymentInstance.Spec.Replicas = *controller.Spec.Replicas
	}

	err = reconciler.Update(ctx, deploymentInstance)
	return err
}

func (reconciler *NginxReconciler) createDeploymentManifest(controller *webserverv1.Nginx) *appsv1.Deployment {

	labels := createSelectorLabels(controller)
	deploymentManifest := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-operator-deployment",
			Namespace: controller.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: controller.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           controller.Spec.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Name:            "nginx",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "nginx-web-port",
						}},
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(controller, deploymentManifest, reconciler.Scheme)
	return deploymentManifest
}
