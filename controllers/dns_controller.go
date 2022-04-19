/*
Copyright 2021.

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
	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cachev1alpha1 "github.com/wdrdres3qew5ts21/coredns-integration-operator/api/v1alpha1"
)

const (
	appName          = "private-dns-"
	dnsZoneConfigMap = "dns-config"
)

// DNSReconciler reconciles a DNS object
type DNSReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=cache.quay.io,resources=dns,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cache.quay.io,resources=dns/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cache.quay.io,resources=dns/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=DaemonSets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DNS object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *DNSReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// 1. Fetch the DNS instance
	instance := &cachev1alpha1.DNS{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("1. Fetch the DNS instance. DNS resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "1. Fetch the DNS instance. Failed to get DNS")
		return ctrl.Result{}, err
	}
	log.Info("1. Fetch the DNS instance. DNS resource found", "DNS.Name", instance.Name, "DNS.Namespace", instance.Namespace)

	fullAppInstanceName := appName + instance.Name

	foundConfigMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: fullAppInstanceName, Namespace: instance.Namespace}, foundConfigMap)

	// 2.0 Check if the ConfigMap Endpoint already exists, and create one if not exists.
	if err != nil && errors.IsNotFound(err) {
		// Define a new ConfigMap
		configMap := r.configMapForDNS(instance, fullAppInstanceName)
		log.Info("2.0.1  Check if the ConfigMap already exists, if not create a new one. Creating a new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
		err = r.Create(ctx, configMap)
		if err != nil {
			log.Error(err, "2.0.1  Check if the ConfigMap Endpoint already exists, if not create a new one. Failed to create new ConfigMap", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)
			return ctrl.Result{}, err
		}
		// ConfigMap created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "2.0.1  Check if the ConfigMap Endpoint already exists, if not create a new one. Failed to get ConfigMap")

		return ctrl.Result{}, err
	}
	// Check if CRD had been changed so it will redeploy ConfigMap or Update
	configMap := r.configMapForDNS(instance, fullAppInstanceName)
	if !equality.Semantic.DeepDerivative(configMap.Data, foundConfigMap.Data) {
		foundConfigMap = configMap
		log.Info("2.0.2 Updating ConfigMap", "ConfigMap.Namespace", foundConfigMap.Namespace, "ConfigMap.Name", foundConfigMap.Name)
		err := r.Update(ctx, foundConfigMap)
		if err != nil {
			log.Error(err, "Failed to update ConfigMap", "ConfigMap.Namespace", foundConfigMap.Namespace, "ConfigMap.Name", foundConfigMap.Name)
			return ctrl.Result{}, err
		}
		// rollout DaemonSet for take DNSRecord change from ConfigMap
		foundDaemonSet := &appsv1.DaemonSet{}
		err = r.Get(ctx, types.NamespacedName{Name: fullAppInstanceName, Namespace: instance.Namespace}, foundDaemonSet)
		if err == nil {
			// patch ConfigMap
			patch := client.MergeFrom(foundDaemonSet.DeepCopy())
			foundDaemonSet.Spec.Template.Annotations = map[string]string{
				"kubectl.kubernetes.io/restartedAt": time.Now().Format(time.RFC3339),
			}
			err := r.Patch(ctx, foundDaemonSet, patch)
			log.Info("2.0.3  Rollout DaemonSet", "DaemonSet.Namespace", foundDaemonSet.Namespace, "DaemonSet.Name", foundDaemonSet.Name)
			if err != nil {
				log.Error(err, "2.0.3  Check if Rollout DaemonSet fail", "DaemonSet.Namespace", foundDaemonSet.Namespace, "DaemonSet.Name", foundDaemonSet.Name)
				return ctrl.Result{}, err
			}
			// DaemonSet created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "2.1  Check if the DaemonSet already exists, if not create a new one. Failed to get DaemonSet")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 2.1 Check if the DaemonSet already exists, and create one if not exists.
	foundDaemonSet := &appsv1.DaemonSet{}
	err = r.Get(ctx, types.NamespacedName{Name: fullAppInstanceName, Namespace: instance.Namespace}, foundDaemonSet)
	if err != nil && errors.IsNotFound(err) {
		// Define a new DaemonSet
		daemonSet := r.DaemonSetForDNS(instance, fullAppInstanceName)
		log.Info("2.1  Check if the DaemonSet already exists, if not create a new one. Creating a new DaemonSet", "DaemonSet.Namespace", daemonSet.Namespace, "DaemonSet.Name", daemonSet.Name)
		err = r.Create(ctx, daemonSet)
		if err != nil {
			log.Error(err, "2.1  Check if the DaemonSet already exists, if not create a new one. Failed to create new DaemonSet", "DaemonSet.Namespace", daemonSet.Namespace, "DaemonSet.Name", daemonSet.Name)
			return ctrl.Result{}, err
		}
		// DaemonSet created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "2.1  Check if the DaemonSet already exists, if not create a new one. Failed to get DaemonSet")
		return ctrl.Result{}, err
	}

	// 2.2 Check if the Service Endpoint already exists, and create one if not exists.
	foundService := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{Name: fullAppInstanceName, Namespace: instance.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		service := r.serviceForDNS(instance, fullAppInstanceName)
		log.Info("2.2  Check if the Service Endpoint already exists, if not create a new one. Creating a new Service Endpoint", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
		err = r.Create(ctx, service)
		if err != nil {
			log.Error(err, "2.2  Check if the Service Endpoint already exists, if not create a new one. Failed to create new Service Endpoint", "Service.Namespace", service.Namespace, "Service.Name", service.Name)
			return ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "2.2  Check if the Service Endpoint already exists, if not create a new one. Failed to get Service Endpoint")
		return ctrl.Result{}, err
	}

	// 4. Update the DNS status with the pod names
	// List the pods for this DNS's DaemonSet
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels(labelsForDNS(fullAppInstanceName)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "4. Update the DNS status with the pod names. Failed to list pods", "DNS.Namespace", instance.Namespace, "DNS.Name", instance.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)
	log.Info("4. Update the DNS status with the pod names. Pod list", "podNames", podNames)
	// Update status.Nodes if needed
	if !reflect.DeepEqual(podNames, instance.Status.Nodes) {
		instance.Status.Nodes = podNames
		err := r.Status().Update(ctx, instance)
		if err != nil {
			log.Error(err, "4. Update the DNS status with the pod names. Failed to update DNS status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DNSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cachev1alpha1.DNS{}).
		Owns(&appsv1.DaemonSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}

// serviceForDNS returns a DNS Service object
func (r *DNSReconciler) configMapForDNS(instance *cachev1alpha1.DNS, fullAppInstanceName string) *corev1.ConfigMap {
	// Create Kubernetes Service for resolve Private DNS Server

	corefileResult := ""
	dnsRecordZoneMap := make(map[string]string)
	// Corefile Result List
	for i, zone := range instance.Spec.DomainZones {
		corefileResult += zone.Name + `:8053 {
			reload 3s
			erratic
			errors
			log stdout
			file /etc/coredns/` + zone.Name + `
		}` + "\n"

		// Zone Record Generated
		dnsRecords := instance.Spec.DomainZones[i].DNSRecords
		var dnsRecordResult string = "\n"
		for _, record := range dnsRecords {
			dnsRecordResult += record.Name + " IN " + string(record.RecordType) + " " + record.Target + "\n"
		}
		dnsRecordZoneMap[zone.Name] = "$TTL    1800\n" +
			"$ORIGIN " + zone.Name + ".\n" +
			`@ IN SOA dns domains (
		2020031101   ; serial
		300          ; refresh
		1800         ; retry
		14400        ; expire
		300 )        ; minimum` +
			"\n;PRIVATE_DNS_RECORD" +
			dnsRecordResult +
			";END_PRIVATE_DNS_RECORD\n"
	}
	dnsRecordZoneMap["Corefile"] = corefileResult

	// Final Result Corefile for CoreDNS configuration
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
		Name:      fullAppInstanceName,
		Namespace: instance.Namespace,
		Labels:    labelsForDNS(fullAppInstanceName),
	}, Data: dnsRecordZoneMap}
	ctrl.SetControllerReference(instance, configMap, r.Scheme)
	return configMap
}

// serviceForDNS returns a DNS Service object
func (r *DNSReconciler) serviceForDNS(instance *cachev1alpha1.DNS, fullAppInstanceName string) *corev1.Service {
	// Create Kubernetes Service for resolve Private DNS Server
	service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
		Name:      fullAppInstanceName,
		Namespace: instance.Namespace,
		Labels:    labelsForDNS(fullAppInstanceName),
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
		Selector: labelsForDNS(fullAppInstanceName),
		// ClusterIP:  "172.21.103.99",
		// ClusterIPs: []string{"172.21.103.99"},
		Type: corev1.ServiceTypeClusterIP,
	}}
	ctrl.SetControllerReference(instance, service, r.Scheme)
	return service
}

// DaemonSetForDNS returns a DNS DaemonSet object
func (r *DNSReconciler) DaemonSetForDNS(instance *cachev1alpha1.DNS, fullAppInstanceName string) *appsv1.DaemonSet {
	configMapMode := int32(420)
	// dynamic mount volume
	// volumeList := []corev1.Volume{corev1.Volume{}, corev1.Volume{}}
	// volumeMount := []corev1.VolumeMount{
	// 	{
	// 		Name:      fullAppInstanceName,
	// 		MountPath: "/etc/coredns",
	// 	},
	// }
	// for i, zone := range instance.Spec.DomainZones {
	// 	log.Log.Info("Inside Loop of hard")
	// 	volumeList[i] = corev1.Volume{
	// 		Name: strings.ReplaceAll(zone.Name, ".", "-"),
	// 		VolumeSource: corev1.VolumeSource{
	// 			ConfigMap: &corev1.ConfigMapVolumeSource{
	// 				LocalObjectReference: corev1.LocalObjectReference{Name: zone.Name},
	// 				DefaultMode:          &configMapMode,
	// 			}},
	// 	}
	// }

	// Generate DaemonSet
	dep := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullAppInstanceName,
			Namespace: instance.Namespace,
			Labels:    labelsForDNS(fullAppInstanceName),
		},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForDNS(fullAppInstanceName),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelsForDNS(fullAppInstanceName),
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: fullAppInstanceName,
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: fullAppInstanceName},
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
								Name:      fullAppInstanceName,
								MountPath: "/etc/coredns",
							},
						},
						Env: []corev1.EnvVar{
							{
								Name:  "TEST_VARIABLE",
								Value: "thesis-chula-demo",
							},
						}}},
				}},
		}}
	// Set DNS instance as the owner and controller
	ctrl.SetControllerReference(instance, dep, r.Scheme)
	return dep
}

// labelsForDNS returns the labels for selecting the resources
// belonging to the given DNS CR name.
func labelsForDNS(fullAppInstanceName string) map[string]string {
	return map[string]string{
		"app": fullAppInstanceName,
	}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
