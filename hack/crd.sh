go run vendor/sigs.k8s.io/controller-tools/cmd/controller-gen/main.go crd

for file in config/crds/*; do
  sed 's/nativelb\.k8s\.native-lb/k8s\.native-lb/' ${file}>  config/base/crds/`basename ${file}`
done

rm -rf config/crds
