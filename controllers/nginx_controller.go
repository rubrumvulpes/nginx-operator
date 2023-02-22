/*
Copyright 2023.

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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webserverv1 "nginx-operator/api/v1"
)

// NginxReconciler reconciles a Nginx object
type NginxReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webserver.cisco.davidkertesz.hu,resources=nginxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webserver.cisco.davidkertesz.hu,resources=nginxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webserver.cisco.davidkertesz.hu,resources=nginxes/finalizers,verbs=update
//+kubebuilder:rbac:groups=webserver.cisco.davidkertesz.hu,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webserver.cisco.davidkertesz.hu,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=webserver.cisco.davidkertesz.hu,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webserver.cisco.davidkertesz.hu,resources=services/status,verbs=get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Nginx object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (reconciler *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)

	logger.Info("retrieving Nginx operator")
	var nginx webserverv1.Nginx
	if err := reconciler.Get(ctx, req.NamespacedName, &nginx); err != nil {
		logger.Error(err, "unable to fetch Nginx instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err := reconciler.reconcileDeployment(ctx, &nginx)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = reconciler.reconcileService(ctx, &nginx)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = reconciler.reconcileIngress(ctx, &nginx)
	if err != nil {
		return ctrl.Result{}, err
	}

	logger.Info("reconciliation loop finished")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (reconciler *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webserverv1.Nginx{}).
		Complete(reconciler)
}

func createSelectorLabels(controller *webserverv1.Nginx) map[string]string {
	return map[string]string{
		"app":      "nginx",
		"operator": controller.GetName(),
	}
}
