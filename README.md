# README

## Project Overview called "Resource Memutator"

This is the initial push of a project aimed at creating a custom Mutation Webhook called Resource memutator. The webhook is still in its alpha phase, and it is not yet production-ready. I'm learning as I develop it, exploring advanced Kubernetes concepts as an SRE/DevOps and enhancing my Go skills through hands-on experience. At this stage, it does not mutate resources, but it logs useful information about incoming requests and any memory limit mismatches. Any suggestions or help are more than welcome!

My webhook intercepts API requests when resources like Deployments or StatefulSets are created or updated. Instead of modifying those resources, it logs useful details, such as differences between requested and actual memory limits. This helps track configuration issues. In the near future, I plan to implement functionality that will automatically mutate the resource, ensuring that resource limits match the resource requests based on specific rules.

In short kubeapi will know as we will registry this custom MutatingWebhookConfiguration, you can check under the list of MutatingWebhook:

`/apis/admissionregistration.k8s.io/v1/mutatingwebhookconfigurations`

## Upcoming Improvements
 
- Move the Docker build to Docker Hub and automate the image build and release process.
- Transition the deployment to use Helm for easier management and scalability.

Stay tuned for further updates as I push more changes!

# Testing the Mutation Webhook (manually setup atm)

- Create a Kind Cluster
First, you need to create a Kubernetes cluster using Kind. Use the following command to create the cluster based on your specific Kind configuration:

`kind create cluster --config kind_cluster_setup/kind_config.yaml`

- Build docker image locally

`cd resource-memutator/docker buildx build --platform linux/amd64 -t resource-memutator:latest .`

- After building your Docker image locally, you'll need to load it into the Kind cluster

`kind load docker-image resource-memutator:latest`

- Install cert-manager
Use Helm to install cert-manager, which will manage your certificates for the webhook

```
helm repo add jetstack https://charts.jetstack.io
helm repo update
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --version v1.15.3 \
  --set installCRDs=true
```

- Cert-manager will need a configuration to generate the certificate required for your webhook. Apply the certificate YAML file as follows:

`kubectl apply -f mutating_webhook/0_certificate_generate.yaml`

- Deploy the webhook and its configuration:

```
kubectl apply -f deploy/resource_memutaror_deploy.yaml
kubectl apply -f mutating_webhook/1_mutating_webhook_resources.yaml
```

- you should seen the pod up and running

`kubectl get pod`

then follow the log of the webhook pod

`kubectl logs resource-memutator-webhook-123 -f`

- Apply some test resources (a Deployment and StatefulSet) to verify that the webhook works as expected and logs the necessary information.

`kubectl apply -f resource_test/Deployment.yaml`
2nd test 
`kubectl apply -f resource_test/StatefulSet.yaml`

in the follow logs you should seen:

```
2024/10/06 15:42:06 Starting webhook server on port 443...
2024/10/06 15:44:03 Container 'test-container' in Deployment 'test-deployment' has mismatched memory requests and limits. Request: 64Mi, Limit: 128Mi
2024/10/06 15:44:11 Container 'busybox' in StatefulSet 'kube-system' has mismatched memory requests and limits. Request: 64Mi, Limit: 128Mi
```

# cleaup 

delete the kind cluster

`kind delete clusters kind`

