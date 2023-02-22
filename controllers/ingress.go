package controllers

import (
	"context"
	webserverv1 "nginx-operator/api/v1"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (reconciler *NginxReconciler) reconcileIngress(ctx context.Context, controller *webserverv1.Nginx) error {

	logger := log.FromContext(ctx)
	logger.Info("reconciling Ingress")

	ingressManifest := reconciler.createIngressManifest(controller)
	ingressInstance := &networkingv1.Ingress{}

	err := reconciler.Get(ctx, types.NamespacedName{
		Name:      ingressManifest.Name,
		Namespace: controller.Namespace,
	}, ingressInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			err := reconciler.Create(ctx, ingressManifest)
			if err != nil {
				logger.Error(err, "no Ingress found, creation failed")
				return err
			}
			logger.Info("no Ingress found, created")
			return nil
		} else {
			logger.Error(err, "error while looking up Ingress")
			return err
		}
	}

	if controller.Spec.Host != ingressInstance.Spec.TLS[0].Hosts[0] {
		ingressInstance.Spec.TLS[0].Hosts[0] = controller.Spec.Host
	}
	if controller.Spec.Host != ingressInstance.Spec.Rules[0].Host {
		ingressInstance.Spec.Rules[0].Host = controller.Spec.Host
	}

	err = reconciler.Update(ctx, ingressInstance)
	return err
}

func (reconciler *NginxReconciler) createIngressManifest(controller *webserverv1.Nginx) *networkingv1.Ingress {

	ingressClass := "nginx"
	prefixPathType := networkingv1.PathTypePrefix
	ingressManifest := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "nginx-operator-ingress",
			Namespace:   controller.Namespace,
			Annotations: map[string]string{"cert-manager.io/cluster-issuer": "letsencrypt-staging"},
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClass,
			TLS: []networkingv1.IngressTLS{
				networkingv1.IngressTLS{
					SecretName: "letsencrypt-staging",
					Hosts: []string{
						controller.Spec.Host,
					},
				},
			},
			Rules: []networkingv1.IngressRule{
				networkingv1.IngressRule{
					Host: controller.Spec.Host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								networkingv1.HTTPIngressPath{
									Path:     "/",
									PathType: &prefixPathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "nginx-operator-service",
											Port: networkingv1.ServiceBackendPort{
												Number: int32(80),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	controllerutil.SetControllerReference(controller, ingressManifest, reconciler.Scheme)
	return ingressManifest
}
