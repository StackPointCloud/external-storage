# oneandone-provisioner
1&amp;1 Flex Volume Provisioner for Kubernetes

# Requirements

- Running Kuberenetes cluster at 1&1.

# Steps

1. Build the provisioner:
    ```
    make container
    ```
    This step will build the binaries and push them to quay.io/stackpointcloud/oneandone-provisioner.


2. Create Kubernetes secret:
    [secret.yaml](./deploy/kubernetes/secret.yaml)
    ```
    apiVersion: v1
    kind: Secret
    metadata:
        name: oneandone
        namespace: kube-system
    data:
        token: "[base64-encoded-1and-token]"
        credentials-datacenter: "[base64-encoded-1and1-datacenter-id]"
    ```
    Push the secret to the cluster:
    ```

    kubectl create -f secret.yaml
    ```
    Validate that the secret has been successfuly created:

    ```
    kubectl get -n kube-system oneandone -o yaml
    ```

3. Deploy provisioner:
    [oneandone-provisioner.yaml](./deploy/kubernetes/oneandone-provisioner.yaml)
    ```
    kubectl create -f oneandone-provisioner.yaml
    ```

4. Create a storage class:
    [storage-class.yaml](./deploy/kubernetes/storage-class.yaml)
    ```
    kubectl create -f storage-class.yaml
    ```
5. Create a persistent volume claim:
    [pvc.yaml](./deploy/kubernetes/storage-class.yaml)
    ```
    kubectl create -f pvc.yml
    ```
    
By completing the final step the provisioner will create 1&1 block storage. To validate you can do:

```
kubectl get pv -o yaml --kubeconfig=kubeconfig 
```

Or query 1&1 Cloud Server API:

```
curl -sX GET -H 'Content-Type: application/json' -H 'X-TOKEN: [1and1-token]' https://cloudpanel-api.1and1.com/v1/block_storages
```

If a problem occurs you can follow the provisioner logs:
```
$kubectl get pods -n kube-system 
NAME                                      READY     STATUS    RESTARTS   AGE
...
oneandone-provisioner-784bf4dc55-6cptm    1/1       Running   0          55m
...
```

```
kubectl logs  -n kube-system oneandone-provisioner-784bf4dc55-6cptm
```