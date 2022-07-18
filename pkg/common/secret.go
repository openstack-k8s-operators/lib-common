/*
Copyright 2020 Red Hat

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

package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openstack-k8s-operators/lib-common/pkg/condition"
	"github.com/openstack-k8s-operators/lib-common/pkg/helper"

	corev1 "k8s.io/api/core/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// GetSecret -
func GetSecret(
	ctx context.Context,
	h *helper.Helper,
	secretName string,
	secretNamespace string,
) (*corev1.Secret, string, error) {
	secret := &corev1.Secret{}

	err := h.GetClient().Get(ctx, types.NamespacedName{Name: secretName, Namespace: secretNamespace}, secret)
	if err != nil {
		return nil, "", err
	}

	secretHash, err := ObjectHash(secret)
	if err != nil {
		return nil, "", fmt.Errorf("error calculating configuration hash: %v", err)
	}
	return secret, secretHash, nil
}

// GetSecrets -
func GetSecrets(
	ctx context.Context,
	h *helper.Helper,
	secretNamespace string,
	labelSelectorMap map[string]string,
) (*corev1.SecretList, error) {
	var secrets *corev1.SecretList

	secrets, err := h.GetKClient().CoreV1().Secrets(secretNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.FormatLabels(labelSelectorMap),
	})

	if err != nil {
		return secrets, err
	}

	return secrets, nil
}

// CreateOrPatchSecret -
func CreateOrPatchSecret(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	secret *corev1.Secret,
) (string, controllerutil.OperationResult, error) {

	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), secret, func() error {

		err := controllerutil.SetControllerReference(obj, secret, h.GetScheme())
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", op, fmt.Errorf("error create/updating secret: %v", err)
	}

	secretHash, err := ObjectHash(secret)
	if err != nil {
		return "", "", fmt.Errorf("error calculating configuration hash: %v", err)
	}

	return secretHash, op, err
}

// createOrUpdateSecret -
func createOrUpdateSecret(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	st Template,
) (string, controllerutil.OperationResult, error) {
	data := make(map[string][]byte)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        st.Name,
			Namespace:   st.Namespace,
			Annotations: st.Annotations,
		},
		Data: data,
	}

	if st.SecretType != "" {
		secret.Type = st.SecretType
	}

	// create or update the CM
	op, err := controllerutil.CreateOrPatch(ctx, h.GetClient(), secret, func() error {
		secret.Labels = st.Labels
		// add data from templates
		renderedTemplateData, err := GetTemplateData(st)
		if err != nil {
			return err
		}
		dataString := renderedTemplateData

		// add provided custom data to dataString
		// Note: this can overwrite data rendered from GetTemplateData() if key is same
		if len(st.CustomData) > 0 {
			for k, v := range st.CustomData {
				dataString[k] = v
			}
		}

		for k, d := range dataString {
			data[k] = []byte(d)
		}
		secret.Data = data

		// Only set controller ref if namespaces are equal, else we hit an error
		if obj.GetNamespace() == secret.Namespace {
			if !st.SkipSetOwner {
				err := controllerutil.SetControllerReference(obj, secret, h.GetScheme())
				if err != nil {
					return err
				}
			}
		} else {
			// Set ownership labels that can be found by the respective controller kind
			ownerLabel := fmt.Sprintf("%s.%s", strings.ToLower(st.InstanceType), obj.GetObjectKind().GroupVersionKind().Group)
			labelSelector := map[string]string{
				ownerLabel + "/uid":       string(obj.GetUID()),
				ownerLabel + "/namespace": obj.GetNamespace(),
				ownerLabel + "/name":      obj.GetName(),
			}

			secret.GetObjectMeta().SetLabels(labels.Merge(secret.GetObjectMeta().GetLabels(), labelSelector))
		}

		return nil
	})

	if err != nil {
		return "", op, err
	}

	secretHash, err := ObjectHash(secret)
	if err != nil {
		return "", op, fmt.Errorf("error calculating configuration hash: %v", err)
	}

	return secretHash, op, nil
}

// createOrGetCustomSecret -
func createOrGetCustomSecret(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	st Template,
) (string, error) {
	// Check if this secret already exists
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        st.Name,
			Namespace:   st.Namespace,
			Labels:      st.Labels,
			Annotations: st.Annotations,
		},
		Data: map[string][]byte{},
	}

	if st.SecretType != "" {
		secret.Type = st.SecretType
	}

	foundSecret := &corev1.Secret{}
	err := h.GetClient().Get(ctx, types.NamespacedName{Name: st.Name, Namespace: st.Namespace}, foundSecret)
	if err != nil && k8s_errors.IsNotFound(err) {
		err := controllerutil.SetControllerReference(obj, secret, h.GetScheme())
		if err != nil {
			return "", err
		}

		h.GetLogger().Info(fmt.Sprintf("Creating a new Secret %s in namespace %s", st.Namespace, st.Name))
		err = h.GetClient().Create(ctx, secret)
		if err != nil {
			return "", err
		}
	} else {
		// use data from already existing custom secret
		secret.Data = foundSecret.Data
	}

	secretHash, err := ObjectHash(secret)
	if err != nil {
		return "", fmt.Errorf("error calculating configuration hash: %v", err)
	}

	return secretHash, nil
}

// EnsureSecrets - get all secrets required, verify they exist and add the hash to env and status
func EnsureSecrets(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	sts []Template,
	envVars *map[string]EnvSetter,
) error {
	var err error

	for _, s := range sts {
		var hash string
		var op controllerutil.OperationResult

		if s.Type != TemplateTypeCustom {
			hash, op, err = createOrUpdateSecret(ctx, h, obj, s)
		} else {
			hash, err = createOrGetCustomSecret(ctx, h, obj, s)
			// set op to OperationResultNone because createOrGetCustomSecret does not return an op
			// and it will add log entries bellow with none operation
			op = controllerutil.OperationResult(controllerutil.OperationResultNone)
		}
		if err != nil {
			return err
		}
		if op != controllerutil.OperationResultNone {
			h.GetLogger().Info(fmt.Sprintf("Secret %s successfully reconciled - operation: %s", s.Name, string(op)))
		}
		if envVars != nil {
			(*envVars)[s.Name] = EnvValue(hash)
		}
	}

	return nil
}

// DeleteSecretsWithLabel - Delete all secrets in namespace of the obj matching label selector
func DeleteSecretsWithLabel(
	ctx context.Context,
	h *helper.Helper,
	obj client.Object,
	labelSelectorMap map[string]string,
) error {
	err := h.GetClient().DeleteAllOf(
		ctx,
		&corev1.Secret{},
		client.InNamespace(obj.GetNamespace()),
		client.MatchingLabels(labelSelectorMap),
	)
	if err != nil && !k8s_errors.IsNotFound(err) {
		err = fmt.Errorf("Error DeleteAllOf Secret: %v", err)
		return err
	}

	return nil
}

//
// DeleteSecretsWithName - Delete names secret object in namespace
//
func DeleteSecretsWithName(
	ctx context.Context,
	h *helper.Helper,
	cond *condition.Condition,
	name string,
	namespace string,
) (condition.Condition, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	err := h.GetClient().Delete(ctx, secret, &client.DeleteOptions{})
	if err != nil && !k8s_errors.IsNotFound(err) {
		msg := fmt.Sprintf("Failed to delete %s %s", secret.Kind, secret.Name)
		cond := condition.NewCondition(
			condition.TypeError,
			corev1.ConditionTrue,
			ReasonSecretDeleteError,
			msg)

		return cond, err
	}

	LogForObject(
		h,
		fmt.Sprintf("Secret %s in namespace %s deleted", secret.Name, secret.Namespace),
		secret,
	)

	return condition.Condition{}, nil
}

//
// GetDataFromSecret - Get data from Secret
//
// if the secret or data is not found, requeue after requeueTimeout in seconds
func GetDataFromSecret(
	ctx context.Context,
	h *helper.Helper,
	secretName string,
	requeueTimeout int,
	key string,
) (string, condition.Condition, ctrl.Result, error) {

	data := ""

	secret, _, err := GetSecret(ctx, h, secretName, h.GetBeforeObject().GetNamespace())
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			msg := fmt.Sprintf("%s secret does not exist", secretName)
			cond := condition.NewCondition(
				condition.TypeError,
				corev1.ConditionTrue,
				ReasonSecretMissing,
				msg)

			return data, cond, ctrl.Result{RequeueAfter: time.Duration(requeueTimeout) * time.Second}, nil
		}
		msg := fmt.Sprintf("Error getting %s secret", secretName)
		cond := condition.NewCondition(
			condition.TypeError,
			corev1.ConditionTrue,
			ReasonSecretDeleteError,
			msg)

		return data, cond, ctrl.Result{}, err
	}

	if key != "" {
		val, ok := secret.Data[key]
		if !ok {
			msg := fmt.Sprintf("%s not found in secret %s",
				key,
				secretName)
			cond := condition.NewCondition(
				condition.TypeError,
				corev1.ConditionTrue,
				ReasonSecretError,
				msg)

			return data, cond, ctrl.Result{}, fmt.Errorf(cond.Message)
		}
		data = strings.TrimSuffix(string(val), "\n")
	}

	return data, condition.Condition{}, ctrl.Result{}, nil
}
