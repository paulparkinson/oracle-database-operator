/*
** Copyright (c) 2021 Oracle and/or its affiliates.
**
** The Universal Permissive License (UPL), Version 1.0
**
** Subject to the condition set forth below, permission is hereby granted to any
** person obtaining a copy of this software, associated documentation and/or data
** (collectively the "Software"), free of charge and under any and all copyright
** rights in the Software, and any and all patent rights owned or freely
** licensable by each licensor hereunder covering either (i) the unmodified
** Software as contributed to or provided by such licensor, or (ii) the Larger
** Works (as defined below), to deal in both
**
** (a) the Software, and
** (b) any piece of software and/or hardware listed in the lrgrwrks.txt file if
** one is included with the Software (each a "Larger Work" to which the Software
** is contributed by such licensors),
**
** without restriction, including without limitation the rights to copy, create
** derivative works of, display, perform, and distribute the Software and make,
** use, sell, offer for sale, import, export, have made, and have sold the
** Software and the Larger Work(s), and to sublicense the foregoing rights on
** either these or other terms.
**
** This license is subject to the following condition:
** The above copyright notice and either this complete permission notice or at
** a minimum a reference to the UPL must be included in all copies or
** substantial portions of the Software.
**
** THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
** IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
** FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
** AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
** LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
** OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
** SOFTWARE.
 */

package v1alpha1

import (
	"encoding/json"
	// "strconv"
	//
	// "github.com/oracle/oci-go-sdk/v45/database"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	//
	"github.com/oracle/oracle-database-operator/commons/annotations"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AutonomousDatabaseSpec defines the desired state of AutonomousDatabase
// Important: Run "make" to regenerate code after modifying this file
type DatabaseMetricsSpec struct {
	Details DatabaseMetricsDetails `json:"details"`
	// OCIConfig OCIConfigSpec             `json:"ociConfig,omitempty"`
	HardLink *bool `json:"hardLink,omitempty"`
}

//
// type OCIConfigSpec struct {
// 	ConfigMapName *string `json:"configMapName,omitempty"`
// 	SecretName    *string `json:"secretName,omitempty"`
// }
//
// // AutonomousDatabaseDetails defines the detail information of AutonomousDatabase, corresponding to oci-go-sdk/database/AutonomousDatabase
type DatabaseMetricsDetails struct {
	// AutonomousDatabaseOCID *string `json:"autonomousDatabaseOCID,omitempty"`
	// CompartmentOCID        *string `json:"compartmentOCID,omitempty"`
	// DisplayName            *string `json:"displayName,omitempty"`
	DbName *string `json:"dbName,omitempty"`
	// +kubebuilder:validation:Enum:=OLTP;DW;AJD;APEX
	// DbWorkload           database.AutonomousDatabaseDbWorkloadEnum     `json:"dbWorkload,omitempty"`
	// IsDedicated          *bool                                         `json:"isDedicated,omitempty"`
	// DbVersion            *string                                       `json:"dbVersion,omitempty"`
	// DataStorageSizeInTBs *int                                          `json:"dataStorageSizeInTBs,omitempty"`
	// CPUCoreCount         *int                                          `json:"cpuCoreCount,omitempty"`
	// AdminPassword        PasswordSpec                                  `json:"adminPassword,omitempty"`
	// IsAutoScalingEnabled *bool                                         `json:"isAutoScalingEnabled,omitempty"`
	// LifecycleState       database.AutonomousDatabaseLifecycleStateEnum `json:"lifecycleState,omitempty"`
	// SubnetOCID           *string                                       `json:"subnetOCID,omitempty"`
	// NsgOCIDs             []string                                      `json:"nsgOCIDs,omitempty"`
	// PrivateEndpoint      *string                                       `json:"privateEndpoint,omitempty"`
	// PrivateEndpointLabel *string                                       `json:"privateEndpointLabel,omitempty"`
	// PrivateEndpointIP    *string                                       `json:"privateEndpointIP,omitempty"`
	// FreeformTags         map[string]string                             `json:"freeformTags,omitempty"`
	// Wallet               WalletSpec                                    `json:"wallet,omitempty"`
}

// AutonomousDatabaseStatus defines the observed state of AutonomousDatabase
type DatabaseMetricsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DisplayName string `json:"displayName,omitempty"`
	// LifecycleState       database.AutonomousDatabaseLifecycleStateEnum `json:"lifecycleState,omitempty"`
	// IsDedicated          string                                        `json:"isDedicated,omitempty"`
	// CPUCoreCount         int                                           `json:"cpuCoreCount,omitempty"`
	// DataStorageSizeInTBs int                                           `json:"dataStorageSizeInTBs,omitempty"`
	// DbWorkload           database.AutonomousDatabaseDbWorkloadEnum     `json:"dbWorkload,omitempty"`
	// TimeCreated          string                                        `json:"timeCreated,omitempty"`
}

// AutonomousDatabase is the Schema for the autonomousdatabases API
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="dbmetrics";"databasemetrics"
// +kubebuilder:subresource:status
type DatabaseMetrics struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseMetricsSpec   `json:"spec,omitempty"`
	Status DatabaseMetricsStatus `json:"status,omitempty"`
}

// LastSuccessfulSpec is an annotation key which maps to the value of last successful spec
// const LastSuccessfulSpec string = "lastSuccessfulSpec"

// GetLastSuccessfulSpec returns spec from the lass successful reconciliation.
// Returns nil, nil if there is no lastSuccessfulSpec.
func (adb *DatabaseMetrics) GetLastSuccessfulSpec() (*DatabaseMetricsSpec, error) {
	val, ok := adb.GetAnnotations()[LastSuccessfulSpec]
	if !ok {
		return nil, nil
	}

	specBytes := []byte(val)
	sucSpec := DatabaseMetricsSpec{}

	err := json.Unmarshal(specBytes, &sucSpec)
	if err != nil {
		return nil, err
	}

	return &sucSpec, nil
}

// UpdateLastSuccessfulSpec updates lastSuccessfulSpec with the current spec.
func (adb *DatabaseMetrics) UpdateLastSuccessfulSpec(kubeClient client.Client) error {
	specBytes, err := json.Marshal(adb.Spec)
	if err != nil {
		return err
	}

	anns := map[string]string{
		LastSuccessfulSpec: string(specBytes),
	}

	return annotations.SetAnnotations(kubeClient, adb, anns)
}

// UpdateAttrFromOCIAutonomousDatabase updates the attributes from database.AutonomousDatabase object and returns the resource
func (dbmetrics *DatabaseMetrics) UpdateAttrFromOCIAutonomousDatabase(ociObj DatabaseMetrics) *DatabaseMetrics {
	// adb.Spec.Details.AutonomousDatabaseOCID = ociObj.Id
	// adb.Spec.Details.CompartmentOCID = ociObj.CompartmentId
	// adb.Spec.Details.DisplayName = ociObj.DisplayName
	// adb.Spec.Details.DbName = ociObj.DbName
	// adb.Spec.Details.DbWorkload = ociObj.DbWorkload
	// adb.Spec.Details.IsDedicated = ociObj.IsDedicated
	// adb.Spec.Details.DbVersion = ociObj.DbVersion
	// adb.Spec.Details.DataStorageSizeInTBs = ociObj.DataStorageSizeInTBs
	// adb.Spec.Details.CPUCoreCount = ociObj.CpuCoreCount
	// adb.Spec.Details.IsAutoScalingEnabled = ociObj.IsAutoScalingEnabled
	// adb.Spec.Details.LifecycleState = ociObj.LifecycleState
	// adb.Spec.Details.FreeformTags = ociObj.FreeformTags
	//
	// adb.Spec.Details.SubnetOCID = ociObj.SubnetId
	// adb.Spec.Details.NsgOCIDs = ociObj.NsgIds
	// adb.Spec.Details.PrivateEndpoint = ociObj.PrivateEndpoint
	// adb.Spec.Details.PrivateEndpointLabel = ociObj.PrivateEndpointLabel
	// adb.Spec.Details.PrivateEndpointIP = ociObj.PrivateEndpointIp
	//
	// // update the subresource as well
	// adb.Status.DisplayName = *ociObj.DisplayName
	// adb.Status.LifecycleState = ociObj.LifecycleState
	// adb.Status.IsDedicated = strconv.FormatBool(*ociObj.IsDedicated)
	// adb.Status.CPUCoreCount = *ociObj.CpuCoreCount
	// adb.Status.DataStorageSizeInTBs = *ociObj.DataStorageSizeInTBs
	// adb.Status.DbWorkload = ociObj.DbWorkload
	// adb.Status.TimeCreated = ociObj.TimeCreated.String()

	return dbmetrics
}

//the following creates deepcopy ...
// +kubebuilder:object:root=true
type DatabaseMetricsList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseMetrics `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseMetrics{}, &DatabaseMetricsList{})
}

// A helper function which is useful for debugging. The function prints out a structural JSON format.
func (adbmetrics *DatabaseMetrics) String() (string, error) {
	out, err := json.MarshalIndent(adbmetrics, "", "    ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}
