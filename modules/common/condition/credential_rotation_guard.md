# Credential rotation consumer finalizer guard

OpenStack service operators that consume rotating credential secrets
(transport URL, application credentials, or similar) attach a consumer
finalizer to the old secret so it is not deleted until every sub-service
has rolled out with the new credential.

## Shared helpers

Use helpers from these packages instead of duplicating guard logic in each
operator:

- `object.ManageSecretConsumerFinalizer` — add finalizer to current secret
- `object.RemoveSecretConsumerFinalizer` — remove finalizer from old secret
- `object.FinalizeSecretRotation` — rotation guard: hold old finalizer
  until all sub-services are ready, then release
- `object.ManageRotationGracePeriod` — time-based grace period that gives
  sub-CRs time to detect config changes, roll pods, and update conditions
  before the guard evaluates readiness
- `condition.CredentialRotationGuardReady` — compute `guardReady` for the
  rotation guard
- `condition.ServiceInstanceIsReady` — check whether a sub-CR has
  finished reconciling (generation/observedGeneration guard, replica
  counts, and DeploymentReadyCondition)
- `statefulset.IsReady` — check `*Replicas == ReadyReplicas`,
  `*Replicas == UpdatedReplicas`, `Generation == ObservedGeneration`, and
  `CurrentRevision == UpdateRevision` so DeploymentReady is only True when
  all pods have rolled to the new config
- `deployment.IsReady` — check `*Replicas == ReadyReplicas`,
  `*Replicas == UpdatedReplicas`, `Status.Replicas == ReadyReplicas`, and
  `Generation == ObservedGeneration`

## Parent controller integration

### 1. Pass transport URL directly — never through Status

Pass `transportURL.Status.SecretName` as a parameter to sub-CR creation
functions and config generation. Never read from
`instance.Status.TransportURLSecret` when building sub-CR specs — the
status field is only used as the "old" value for `FinalizeSecretRotation`.

Set `instance.Status.TransportURLSecret` early only for first-time setup:

```go
if instance.Status.TransportURLSecret == "" ||
    instance.Status.TransportURLSecret == transportURL.Status.SecretName {
    instance.Status.TransportURLSecret = transportURL.Status.SecretName
}
```

During rotation (old != current), the status is updated solely by
`FinalizeSecretRotation` at the end of reconcile.

### 2. Track stability

Use a simple boolean to track whether all sub-CRs are stable:

```go
allSubCRsStable := true

// After each sub-CR CreateOrPatch:
subCR, op, err := r.subCRCreateOrUpdate(ctx, instance, transportURL.Status.SecretName)
if err != nil { return ctrl.Result{}, err }
if op != controllerutil.OperationResultNone {
    allSubCRsStable = false
}
```

### 3. Transport URL annotation (operators without TransportURLSecret in sub-CR spec)

For operators whose sub-CR specs do not include a `TransportURLSecret`
field (e.g. nova, watcher, glance, designate, octavia, telemetry), add
an annotation inside the `CreateOrPatch` mutate function to force a
spec change when the transport URL rotates:

```go
op, err := controllerutil.CreateOrPatch(ctx, r.Client, subCR, func() error {
    subCR.Spec = spec
    if subCR.Annotations == nil {
        subCR.Annotations = map[string]string{}
    }
    subCR.Annotations["openstack.org/transport-url-secret"] = transportURLSecretName
    return controllerutil.SetControllerReference(instance, subCR, r.Scheme)
})
```

This is not needed for operators that pass `TransportURLSecret` directly
in the sub-CR spec (cinder, manila, ironic, barbican, heat) — the spec
field change already makes `CreateOrPatch` return `"updated"`.

### 4. Compute the rotation guard and finalize

```go
guardReady := condition.CredentialRotationGuardReady(
    allSubCRsStable,
    &instance.Status.Conditions,
)

instance.Status.TransportURLSecret, err = object.FinalizeSecretRotation(
    ctx, helper, instance.Namespace,
    instance.Status.TransportURLSecret,
    transportURL.Status.SecretName,
    myTransportConsumerFinalizer,
    guardReady,
)
```

The same pattern works for application credential secrets:

```go
instance.Status.ApplicationCredentialSecret, err = object.FinalizeSecretRotation(
    ctx, helper, instance.Namespace,
    instance.Status.ApplicationCredentialSecret,
    instance.Spec.Auth.ApplicationCredentialSecret,
    myACConsumerFinalizer,
    guardReady,
)
```

### 5. Rotation grace period (optional)

Use `ManageRotationGracePeriod` to give sub-CRs time to detect config
changes, update their Deployments/StatefulSets, and roll pods before the
guard releases the old secret's consumer finalizer:

```go
rotationPending := instance.Status.TransportURLSecret != "" &&
    instance.Status.TransportURLSecret != transportURL.Status.SecretName

result, graceActive, err := object.ManageRotationGracePeriod(
    ctx, r.Client, instance,
    rotationPending,
    30*time.Second,
)
if err != nil { return ctrl.Result{}, err }
if graceActive {
    return result, nil
}
```

When `rotationPending` is true and no grace period is active, it sets the
`openstack.org/rotation-grace-until` annotation and returns a requeue.
While the grace period is active, it continues to requeue. After the
grace period expires, it returns `graceActive == false` so the caller
can proceed to evaluate the rotation guard. When `rotationPending` is
false, it clears any existing annotation.

## Sub-CR integration

Sub-CR controllers should use `statefulset.IsReady()` or
`deployment.IsReady()` for the `DeploymentReadyCondition` check:

```go
if statefulset.IsReady(sts) {
    instance.Status.Conditions.MarkTrue(
        condition.DeploymentReadyCondition,
        condition.DeploymentReadyMessage)
} else if *instance.Spec.Replicas > 0 {
    instance.Status.Conditions.Set(condition.FalseCondition(
        condition.DeploymentReadyCondition,
        condition.RequestedReason,
        condition.SeverityInfo,
        condition.DeploymentReadyRunningMessage))
}
```

`statefulset.IsReady` takes an `appsv1.StatefulSet` and checks
`*Replicas == ReadyReplicas`, `*Replicas == UpdatedReplicas`,
`Generation == ObservedGeneration`, and `CurrentRevision == UpdateRevision`.

`deployment.IsReady` takes an `appsv1.Deployment` and checks
`*Replicas == ReadyReplicas`, `*Replicas == UpdatedReplicas`,
`Status.Replicas == ReadyReplicas`, and `Generation == ObservedGeneration`.

Both ensure DeploymentReady is only True when all pods have rolled to
the new config.

For parent operators that need to check sub-CR readiness directly (e.g.
to feed into the stability tracker), use `ServiceInstanceIsReady`:

```go
ready := condition.ServiceInstanceIsReady(
    subCR.Generation,
    subCR.Status.ObservedGeneration,
    subCR.Status.ReadyCount,
    *subCR.Spec.Replicas,
    &subCR.Status.Conditions,
)
```
