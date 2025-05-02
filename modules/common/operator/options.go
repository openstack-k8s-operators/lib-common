/*
Copyright 2025 Red Hat

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

package operator

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// SetManagerOptions - Get options from environment, validate and set controller manager options.
func SetManagerOptions(options *ctrl.Options, setupLog logr.Logger) error {
	leaseDuration, err := getEnvInDuration("LEASE_DURATION")
	if err != nil {
		return err
	} else if leaseDuration != 0 {
		setupLog.Info("manager configured with lease duration", "seconds", int(leaseDuration.Seconds()))
		options.LeaseDuration = &leaseDuration
	}

	renewDeadline, err := getEnvInDuration("RENEW_DEADLINE")
	if err != nil {
		return err
	} else if renewDeadline != 0 {
		setupLog.Info("manager configured with renew deadline", "seconds", int(renewDeadline.Seconds()))
		options.RenewDeadline = &renewDeadline
	}

	retryPeriod, err := getEnvInDuration("RETRY_PERIOD")
	if err != nil {
		return err
	} else if retryPeriod != 0 {
		setupLog.Info("manager configured with retry period", "seconds", int(retryPeriod.Seconds()))
		options.RetryPeriod = &retryPeriod
	}

	return nil
}

func getEnvInDuration(envName string) (time.Duration, error) {
	var durationInt int64
	if durationStr := os.Getenv(envName); durationStr != "" {
		var err error
		if durationInt, err = strconv.ParseInt(durationStr, 10, 64); err != nil {
			return 0, fmt.Errorf("unable to parse provided '%s', err: '%w'", envName, err)
		}
	}
	return time.Duration(durationInt) * time.Second, nil
}
