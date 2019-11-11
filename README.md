# Ransomware

A nefarious controller that will keep a nuisance pod running unless a secret code is entered.

## Start-from-scratch Guide

This guide demonstrates how this project was built.

```sh
mkdir ransomware
cd ransomware

go mod init ransomware

kubebuilder init --domain meetup.com
kubebuilder create api --group talks --version v1 --kind Ransomware
```

Add fields to RansomwareSpec type (`api/ransomware_types.go`).

```go
	Message    string `json:"message,omitempty"`
	SecretCode string `json:"secretCode,omitempty"`
```

Run `make` to regenerate resource definitions.

Add scheme field to `RansomwareReconciler`: `Scheme *runtime.Scheme` in order to be able to set ownership references. Set scheme in `main()`: `Scheme: mgr.GetScheme()`.

Add watch for pods in `RansomwareReconciler.SetupWithManager()` method: `Owns(&corev1.Pod{})`.

Add logic to `RansomwareReconciler.Reconcile()` method:

|            | correctCode | !correctCode
|------------|-------------|--------------
|  podExists | delete pod  |    no-op
| !podExists |    no-op    |  create pod

Update sample object and seed with incorrect code.

```yaml
spec:
  message: Hahaha
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

In a separate terminal, create an instance of `Ransomware`.

```sh
kubectl apply -f ./config/samples/talks_v1_ransomware.yaml
```

View created nuisance pod.

```sh
kubectl get pods
kubectl logs -f ransomware_sample  # ctrl-c to exit
```

Change `.spec.secretCode` to be `password`. The pod should be removed.

```sh
kubectl edit ransomware ransomware_sample
kubectl get pods
```

## Cleanup

```sh
kind delete cluster --name meetup-talk
```

