# Ransomware

A nefarious controller that will keep a nuisance pod running unless a secret code is entered.

```yaml
apiVersion: talks.meetup.com/v1
kind: Ransomware
metadata:
  name: sample
spec:
  message: Hahaha
  # Wrong code   --> Ensures nuisance pod is running.
  # Correct code --> Nuisance pod gets removed.
  secretCode: idk
```

## Start-from-scratch Guide

This guide demonstrates how this project was built.

```sh
mkdir ransomware
cd ransomware

go mod init ransomware

kubebuilder init --domain meetup.com
kubebuilder create api --group talks --version v1 --kind Ransomware
```

Add fields to RansomwareSpec type (`./api/ransomware_types.go`).

```go
Message    string `json:"message,omitempty"`
SecretCode string `json:"secretCode,omitempty"`
```

Run `make` to regenerate resource definitions.

Add scheme field to `RansomwareReconciler` (`./controllers/ransomware_controller.go`): `Scheme *runtime.Scheme`. This will be used to set ownership references later. Set scheme in `main()`: `Scheme: mgr.GetScheme()`.

Add watch for pods in `RansomwareReconciler.SetupWithManager()` method: `Owns(&corev1.Pod{})`.

Add logic to `RansomwareReconciler.Reconcile()` method:

1. `Get` instance of `Ransomware` referenced by `req` which is passed in as an argument.
2. Build the desired `corev1.Pod` instance.
3. Set the controller reference (`controllerutil.SetControllerReference`) on the pod.
4. `Get` the pod and record if it exists.
5. Compare `secretCode` in `Ransomware` instance to expected value.
6. Execute CRUD on pod:

|            | correctCode | !correctCode
|------------|-------------|--------------
|  podExists | delete pod  |    no-op
| !podExists |    no-op    |  create pod

Update the sample object (`./config/samples/talks_v1_ransomware.yaml`) and seed with incorrect code.

```yaml
spec:
  message: Muahaha
  secretCode: idk
```

Create local cluster.

```sh
kind create cluster --name meetup-talk
export KUBECONFIG="$(kind get kubeconfig-path --name="meetup-talk")"
```

Install CRDs.

```sh
make install
```

Run the controller.

```sh
make run
```

In a separate terminal, create an instance of `Ransomware` in the cluster.

```sh
kubectl apply -f ./config/samples/talks_v1_ransomware.yaml
```

View created nuisance pod.

```sh
kubectl get pods
kubectl logs -f sample  # ctrl-c to exit
```

Change `.spec.secretCode` to be `password`. The pod should be removed.

```sh
kubectl edit ransomware sample
kubectl get pods
```

## Cleanup

```sh
kind delete cluster --name meetup-talk
```

