#! /bin/bash -e

for count in 500 550 600 650 700 750 800 850 900 950 1000
do
	echo "Simulating $count nodes"
	sed 's/INIT_COUNT/'"$count"'/g' gke.yaml.sed > gke.yaml
	kubectl apply -f gke.yaml
	kubectl wait --for=condition=complete --timeout=5m -n cilium job/cilium-nodeperf
	echo "$count nodes" >> results
	kubectl get pods -n cilium | grep cilium-nodeperf | grep Completed | awk '{print $1}' | xargs kubectl logs -n cilium | grep Mean >> results
	kubectl delete -f gke.yaml
done
