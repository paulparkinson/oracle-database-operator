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

package autonomousdatabase

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"github.com/oracle/oci-go-sdk/v45/common"
	"github.com/oracle/oci-go-sdk/v45/database"
	"github.com/oracle/oci-go-sdk/v45/secrets"

	dbv1alpha1 "github.com/oracle/oracle-database-operator/apis/database/v1alpha1"
	"github.com/oracle/oracle-database-operator/commons/oci"
)

// SetStatus sets the status subresource.
func SetStatus(kubeClient client.Client, adb *dbv1alpha1.AutonomousDatabase) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		curADB := &dbv1alpha1.AutonomousDatabase{}

		namespacedName := types.NamespacedName{
			Namespace: adb.GetNamespace(),
			Name:      adb.GetName(),
		}

		if err := kubeClient.Get(context.TODO(), namespacedName, curADB); err != nil {
			return err
		}

		curADB.Status = adb.Status
		return kubeClient.Status().Update(context.TODO(), curADB)
	})
}

func createWalletSecret(kubeClient client.Client, namespacedName types.NamespacedName, data map[string][]byte) error {
	// Create the secret with the wallet data
	stringData := map[string]string{}
	for key, val := range data {
		stringData[key] = string(val)
	}

	walletSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespacedName.Namespace,
			Name:      namespacedName.Name,
		},
		StringData: stringData,
	}

	if err := kubeClient.Create(context.TODO(), walletSecret); err != nil {
		return err
	}
	return nil
}

func CreateWalletSecret(logger logr.Logger, kubeClient client.Client, dbClient database.DatabaseClient, secretClient secrets.SecretsClient, adb *dbv1alpha1.AutonomousDatabase) error {
	// Kube Secret which contains Instance Wallet
	walletName := adb.Spec.Details.Wallet.Name
	if walletName == nil {
		walletName = common.String(adb.GetName() + "-instance-wallet")
	}

	// No-op if Wallet is already downloaded
	walletNamespacedName := types.NamespacedName{
		Namespace: adb.GetNamespace(),
		Name:      *walletName,
	}
	walletSecret := &corev1.Secret{}
	if err := kubeClient.Get(context.TODO(), walletNamespacedName, walletSecret); err == nil {
		return nil
	}

	data, err := oci.GetWallet(logger, kubeClient, dbClient, secretClient, adb)
	if err != nil {
		return err
	}

	if err := createWalletSecret(kubeClient, walletNamespacedName, data); err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Wallet is stored in the Secret %s", *walletName))
	return nil
}

// to get secret refer to https://github.com/kubernetes-sigs/controller-runtime/blob/0a3dd2a36dd971ea2fdcb163b0457e6be0d315bd/pkg/client/example_test.go#L57
func CreateTestSecret(logger logr.Logger, kubeClient client.Client, namespacedName types.NamespacedName, data map[string][]byte) error {

	stringData := map[string]string{}
	for key, val := range data {
		stringData[key] = string(val)
	}

	testSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespacedName.Namespace,
			Name:      namespacedName.Name,
		},
		StringData: stringData,
	}

	if err := kubeClient.Create(context.TODO(), testSecret); err != nil {
		return err
	}
	logger.Info(fmt.Sprintf("Test Secret created %s", namespacedName.Name))
	return nil
}

// https://github.com/kubernetes/client-go/blob/master/examples/create-update-delete-deployment/main.go
func CreateMetricsDeployment(logger logr.Logger, kubeClient client.Client, namespacedName types.NamespacedName, data map[string][]byte) error {

	// deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespacedName.Namespace,
			Name:      "metricsexporter-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			// Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demometrics",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demometrics",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	// result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	err := kubeClient.Create(context.TODO(), deployment)
	if err != nil {
		panic(err)
	}
	logger.Info(fmt.Sprintf("Database metrics exporter deployment created"))
	// logger.Info(fmt.Sprintf("Test deployment created %s", result.GetObjectMeta().GetName()))
	return nil
}

func CreateLogsDeployment(logger logr.Logger, kubeClient client.Client, namespacedName types.NamespacedName, data map[string][]byte) error {

	// deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespacedName.Namespace,
			Name:      "logexporter-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			// Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demologs",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demologs",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	// result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	err := kubeClient.Create(context.TODO(), deployment)
	if err != nil {
		panic(err)
	}
	logger.Info(fmt.Sprintf("Database logs exporter deployment created"))
	// logger.Info(fmt.Sprintf("Test deployment created %s", result.GetObjectMeta().GetName()))
	return nil
}
