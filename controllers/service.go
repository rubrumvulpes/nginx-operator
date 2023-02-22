package controllers

import (
	"context"
	webserverv1 "nginx-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (reconciler *NginxReconciler) reconcileService(ctx context.Context, controller *webserverv1.Nginx) error {

	logger := log.FromContext(ctx)
	logger.Info("reconciling Service")

	serviceInstance := &corev1.Service{}
	serviceManifest := reconciler.createServiceManifest(controller)

	err := reconciler.Get(ctx, types.NamespacedName{
		Name:      serviceManifest.Name,
		Namespace: controller.Namespace,
	}, serviceInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			err := reconciler.Create(ctx, serviceManifest)
			if err != nil {
				logger.Error(err, "no Service found, creation failed")
				return err
			}
			logger.Info("no service found, created")
			return nil
		} else {
			logger.Error(err, "error while looking up Service")
			return err
		}
	}
	logger.Info("Service already exists")
	return nil
}

func (reconciler *NginxReconciler) createServiceManifest(controller *webserverv1.Nginx) *corev1.Service {

	labels := createSelectorLabels(controller)
	serviceManifest := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-operator-service",
			Namespace: controller.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Name:       "nginx-web-serviceport",
				Port:       80,
				TargetPort: intstr.FromInt(80),
			}},
		},
	}

	controllerutil.SetControllerReference(controller, serviceManifest, reconciler.Scheme)
	return serviceManifest
}
