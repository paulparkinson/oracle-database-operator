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

package controllers

import (
	"context"
	// "fmt"
	// "reflect"
	//
	"github.com/go-logr/logr"
	// "github.com/oracle/oci-go-sdk/v45/database"
	// "github.com/oracle/oci-go-sdk/v45/secrets"
	// "github.com/oracle/oci-go-sdk/v45/workrequests"
	//
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	// "k8s.io/apimachinery/pkg/types"
	// "k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	//
	dbv1alpha1 "github.com/oracle/oracle-database-operator/apis/database/v1alpha1"
	adbutil "github.com/oracle/oracle-database-operator/commons/autonomousdatabase"
	// "github.com/oracle/oracle-database-operator/commons/finalizer"
	// "github.com/oracle/oracle-database-operator/commons/oci"
	"k8s.io/apimachinery/pkg/types"
)

// AutonomousDatabaseReconciler reconciles a AutonomousDatabase object
type DatabaseMetricsReconciler struct {
	KubeClient client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme

	currentLogger logr.Logger
}

// SetupWithManager function
func (r *DatabaseMetricsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dbv1alpha1.DatabaseMetrics{}).
		WithEventFilter(r.eventFilterPredicate()).
		WithOptions(controller.Options{MaxConcurrentReconciles: 50}). // ReconcileHandler is never invoked concurrently with the same object.
		Complete(r)
}

func (r *DatabaseMetricsReconciler) eventFilterPredicate() predicate.Predicate {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldADB := e.ObjectOld.DeepCopyObject().(*dbv1alpha1.DatabaseMetrics)
			newADB := e.ObjectNew.DeepCopyObject().(*dbv1alpha1.DatabaseMetrics)

			// Reconciliation should NOT happen if the lastSuccessfulSpec annotation or status.state changes.
			oldSucSpec := oldADB.GetAnnotations()[dbv1alpha1.LastSuccessfulSpec]
			newSucSpec := newADB.GetAnnotations()[dbv1alpha1.LastSuccessfulSpec]

			lastSucSpecChanged := oldSucSpec != newSucSpec
			// stateChanged := oldADB.Status.LifecycleState != newADB.Status.LifecycleState
			// if lastSucSpecChanged || stateChanged {
			if lastSucSpecChanged {
				// Don't enqueue request
				return false
			}
			// Enqueue request
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Do not trigger reconciliation when the real object is deleted from the cluster.
			return false
		},
	}

	return pred
}

// +kubebuilder:rbac:groups=database.oracle.com,resources=databasemetrics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=database.oracle.com,resources=databasemetrics/status,verbs=update;patch
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=create;get;list;update
// +kubebuilder:rbac:groups="",resources=configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is the funtion that the operator calls every time when the reconciliation loop is triggered.
// It go to the beggining of the reconcile if an error is returned. We won't return a error if it is related
// to OCI, because the issues cannot be solved by re-run the reconcile.
func (r *DatabaseMetricsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.currentLogger = r.Log.WithValues("Namespaced/Name", req.NamespacedName)
	r.currentLogger.Info("PAUL DatabaseMetricsReconciler reconcile successfully")

	// Get the autonomousdatabase instance from the cluster
	adb := &dbv1alpha1.DatabaseMetrics{}
	if err := r.KubeClient.Get(context.TODO(), req.NamespacedName, adb); err != nil {
		// Ignore not-found errors, since they can't be fixed by an immediate requeue.
		// No need to change the since we don't know if we obtain the object.
		if !apiErrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}

	testNamespacedName := types.NamespacedName{
		Namespace: "msdataworkshop",
		Name:      "testwallet",
	}

	r.currentLogger.Info("Creating Deployment...")
	// if err := adbutil.CreateTestSecret(r.currentLogger, r.KubeClient, testNamespacedName, nil); err != nil {
	if err := adbutil.CreateTestDeployment(r.currentLogger, r.KubeClient, testNamespacedName, nil); err != nil {
		r.currentLogger.Error(err, "Fail to create deployment")
		return ctrl.Result{}, nil
	}
	r.currentLogger.Info("Finished Creating Deployment")
	return ctrl.Result{}, nil // todo this is wrong
}
