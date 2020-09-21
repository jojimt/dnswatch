package v1alpha

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClientView describes a mapping between clients and their target names
type ClientView struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status ClientViewStatus `json:"status"`
}

// ClientViewStatus contains the status data of a ClientView
type ClientViewStatus struct {
	ClientMeta ClientViewMeta `json:"client-meta,omitempty"`
	DNSReqList []string       `json:"dns-req-list,omitempty"`
}

// ClientViewMeta contains the metadata identifying the client
type ClientViewMeta struct {
	Kind      string `json:"kind,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"namespace,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClientViewList is a list of ClientView objects
type ClientViewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []ClientView `json:"items"`
}
