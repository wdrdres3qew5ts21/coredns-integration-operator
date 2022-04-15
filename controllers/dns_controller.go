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
	"k8s.io/apimachinery/pkg/util/intstr"
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
	setupLog.Info("DNS Controller: Vanila Log by Supakorn Working")

	instance := &cachev1alpha1.DNS{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {
			setupLog.Info("DNS Controller: Delete DaemonSet ;)")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	setupLog.Info("DNS Controller: Escape Error from first error success")
	found := &appsv1.Deployment{}
	findMe := types.NamespacedName{
		Name:      "myDeployment",
		Namespace: instance.Namespace,
	}
	err = r.Client.Get(context.TODO(), findMe, found)

	setupLog.Info("DNS Controller: Before Created DaemonSet")
	if err != nil && errors.IsNotFound(err) {
		// initialize variable template
		replicasSize := int32(1)
		configMapMode := int32(420)
		dnsZoneConfigMap := "dns-config"
		appName := "private-dns-"
		fullAppInstanceName := appName + instance.Name
		// Creation logic
		labels := map[string]string{
			"app": fullAppInstanceName,
		}

		// Create DaemonSet for Private DNS Server

		dep := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fullAppInstanceName,
				Namespace: instance.Namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicasSize,
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: labels,
					},
					Spec: corev1.PodSpec{
						Volumes: []corev1.Volume{{
							Name: dnsZoneConfigMap,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{Name: dnsZoneConfigMap},
									DefaultMode:          &configMapMode,
								}},
						}},
						Containers: []corev1.Container{{
							Image: "quay.io/openshift/origin-coredns:4.9",
							Name:  "dns",
							Ports: []corev1.ContainerPort{{
								ContainerPort: 8053,
								Name:          "dns",
							}},
							Command: []string{"/usr/bin/coredns"},
							Args:    []string{"-dns.port", "8053", "-conf", "/etc/coredns/Corefile"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      dnsZoneConfigMap,
									MountPath: "/etc/coredns",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "TEST_VARIABLE",
									Value: "thesis-chula-demo",
								},
								// {
								// 	Name:      "TEST_REFERENCE",
								// 	ValueFrom: userSecret,
								// }, {
								// 	ValueFrom: passwordSecret,
								// },
							}}},
					}},
			}}
		// Create Kubernetes Service for resolve Private DNS Server
		service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
			Name:      "hardcode-service",
			Namespace: instance.Namespace,
		}, Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:       "8053-tcp",
				Protocol:   corev1.ProtocolTCP,
				Port:       8053,
				TargetPort: intstr.IntOrString{IntVal: 8053},
			}, {
				Name:       "8053-udp",
				Protocol:   corev1.ProtocolUDP,
				Port:       8053,
				TargetPort: intstr.IntOrString{IntVal: 8053},
			}},
			Selector: map[string]string{
				"app": fullAppInstanceName,
			},
			ClusterIP:  "172.21.103.99",
			ClusterIPs: []string{"172.21.103.99"},
			Type:       corev1.ServiceTypeClusterIP,
		}}

		// Wired Every Resource together to create child resource for create or delete whole dependency
		// controllerutil.SetControllerReference(instance, dep, r.Scheme)
		controllerutil.SetControllerReference(instance, service, r.Scheme)

		setupLog.Info("DNS Controller: Try to create DaemonSet !")

		err = r.Create(context.TODO(), service)
		if err != nil {
			setupLog.Error(err, "DNS Controller: Create Service Endpoint Error :(")
			return reconcile.Result{}, err
		} else {
			setupLog.Info("DNS Controller: Create Service Endpoint Successs :)")
		}

		err = r.Create(context.TODO(), dep)
		if err != nil {
			setupLog.Error(err, "DNS Controller: Create DaemonSet Error :(")
			return reconcile.Result{}, err
		} else {
			setupLog.Info("DNS Controller: Create DaemonSet Successs :)")
		}

	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DNSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.DNS{}).
		Complete(r)
}
