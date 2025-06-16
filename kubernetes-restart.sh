echo "Restarting the pods."

kubectl delete pod -l app=postgres
kubectl delete pod -l app=backend
kubectl delete pod -l app=frontend

kubectl apply -f kubernetes/db
kubectl apply -f kubernetes/backend
kubectl apply -f kubernetes/frontend

echo "Pods restarted"

