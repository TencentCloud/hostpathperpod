# hostpathperpod

> Note: Only the directory type is supported.

This flex volume plugin is like hostPath, but create a host directory with pod meta. 

For example, if you specified the host path like "/root/hostpath", the actual path will be "/root/hostpath/<pod_namespace>/<pod_name>/<pod_uid>".

This plugin could be deployed with `deploy-ds.yaml` and used like `nginx.yaml`.