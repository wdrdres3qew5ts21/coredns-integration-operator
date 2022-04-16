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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DNSSpec defines the desired state of DNS
type DNSSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Select DNS Config CRD which define Zone and Record to using for DNS Server Query
	DNSConfigs []string `json:"dnsConfigs,omitempty"`
	// Domain Zone
	DomainZone DomainZone `json:"domainZone,omitempty"`
	Status     DNSStatus  `json:"status,omitempty"`
}

// DomainZone defines DNS Zone
type DomainZone struct {
	// internal zone of Domain like mycompany.local
	Name string `json:"name"`
	// adding  DNS Record for IPv4 or CNAME
	DNSRecords []DNSRecord `json:"dnsRecord,omitempty"`
}

// DNSRecord defines Record for IPv4 or CNAME
type DNSRecord struct {
	Name       string     `json:"name"`
	RecordType RecordType `json:"type"`
	Target     string     `json:"target"`
}

// +kubebuilder:validation:Enum=A;CNAME
type RecordType string

// DNSStatus defines the observed state of DNS
type DNSStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Nodes []string `json:"nodes"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DNS Core Service for Internal DNS Server
type DNS struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DNSSpec   `json:"spec,omitempty"`
	Status DNSStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DNSList contains a list of DNS
type DNSList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DNS `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DNS{}, &DNSList{})
}
