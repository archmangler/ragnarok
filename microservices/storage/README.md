# NFS  Service 
==============

- In the simplest case we use an NFS service on Kubernetes to host data input files for processing by the load generation system.
- The solution documented here usies AWS EKS GP2 EBS volume as a backing for NFSv4 shares
- The following manifests comprise the installation on kubernetes (tested on EKS)

```
-rw-r--r--  1 welcome  staff  769 Dec 24 08:17 nfs-client-example.yaml
-rw-r--r--  1 welcome  staff  230 Dec 24 07:33 nfs-server-storage-pvc.yaml
-rw-r--r--  1 welcome  staff  778 Dec 24 08:12 nfs-server.yaml
-rw-r--r--  1 welcome  staff  262 Dec 24 07:50 nfs-service.yaml
```

# Configuring a client pod to use NFS
====================================

- add the following configuration under your pod/deployment spec:

```
.
.
.
spec:
  # Add the server as an NFS volume for the pod
  volumes:
    - name: nfs-volume
      nfs: 
        # URL for the NFS server
        server: 10.8.0.48  #Change this to the current address of the NFS service , obtained by looking at: kubectl get ep -n <application namespace>
        path: /
  # In this container, we'll mount the NFS volume
  # and write the date to a file inside it.
  containers:
    - name: dataclient 
      image: alpine
      # Mount the NFS volume in the container
      volumeMounts:
        - name: nfs-volume
          mountPath: /data
```


# References
============

- https://stackoverflow.com/questions/56559704/nfs-server-on-kubernetes-minikube-reports-exportfs-does-not-support-nfs-expo
