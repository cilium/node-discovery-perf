#! /bin/bash -e

for count in 1000
do
	echo "Simulating $count nodes"
	sed 's/INIT_COUNT/'"$count"'/g' gke.yaml.sed > gke.yaml
	kubectl apply -f gke.yaml
	kubectl wait --for=condition=complete --timeout=10m -n cilium job/cilium-nodeperf1
	kubectl wait --for=condition=complete --timeout=10m -n cilium job/cilium-nodeperf2
	echo "$count nodes" >> results
	kubectl get pods -n cilium | grep cilium-nodeperf1 | grep Completed | awk '{print $1}' | xargs kubectl logs -n cilium | grep Mean >> results
	kubectl get pods -n cilium | grep cilium-nodeperf2 | grep Completed | awk '{print $1}' | xargs kubectl logs -n cilium | grep Mean >> results
	kubectl delete -f gke.yaml
done
