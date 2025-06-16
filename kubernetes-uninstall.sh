kubectl delete -f kubernetes/
kubectl delete -f kubernetes/backend
kubectl delete -f kubernetes/frontend
kubectl delete -f kubernetes/db

minikube delete --all

sudo rm -f $(which minikube)

rm -rf ~/.minikube
rm -rf ~/.kube
