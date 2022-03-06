/*
Copyright 2022.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	cachev1alpha1 "github.com/wdrdres3qew5ts21/coredns-integration-operator/api/v1alpha1"
)

// DNSReconciler reconciles a DNS object
type DNSReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	setupLog = ctrl.Log.WithName("setup")
)

//+kubebuilder:rbac:groups=cache.quay.io,resources=dns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.quay.io,resources=dns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.quay.io,resources=dns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DNS object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DNSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	setupLog.Info("DNS Controller", "DNS Controller: Vanila Log by Supakorn Working")

	instance := &cachev1alpha1.DNSRecord{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			setupLog.Info("DNS Controller", "DNS Controller: Delete DaemonSet ;)")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	found := &appsv1.Deployment{}
	findMe := types.NamespacedName{
		Name:      "myDeployment",
		Namespace: instance.Namespace,
	}
	err = r.Client.Get(context.TODO(), findMe, found)

	if err != nil && errors.IsNotFound(err) {
		// Creation logic
		labels := map[string]string{
			"app": "visitors", "visitorssite_cr": instance.Name, "tier": "mysql",
		}
		size := int32(1)
		// userSecret := &corev1.EnvVarSource{
		// 	SecretKeyRef: &corev1.SecretKeySelector{
		// 		LocalObjectReference: corev1.LocalObjectReference{Name: mysqlAuthName()},
		// 		Key:                  "username",
		// 	},
		// }
		// passwordSecret := &corev1.EnvVarSource{
		// 	SecretKeyRef: &corev1.SecretKeySelector{
		// 		LocalObjectReference: corev1.LocalObjectReference{Name: mysqlAuthName()},
		// 		Key:                  "password",
		// 	},
		// }
		dep := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysql-backend-service",
				Namespace: instance.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &size,
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Image: "docker.io/mysql:5.7",
							Name:  "visitors-mysql",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 3306,
								Name:          "mysql",
							}},
							Env: []corev1.EnvVar{
								{
									Name:  "MYSQL_ROOT_PASSWORD",
									Value: "password",
								}, {
									Value: "visitors",
								},
								// {
								// 	Name:      "MYSQL_USER",
								// 	ValueFrom: userSecret,
								// }, {
								// 	ValueFrom: passwordSecret,
								// },
							}}},
					}},
			}}
		controllerutil.SetControllerReference(instance, dep, r.Scheme)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DNSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.DNS{}).
		Complete(r)
}
