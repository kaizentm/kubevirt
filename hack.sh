RESOURCE="devices.kubevirt.io/mshv"
VALUE="1"
echo "Patching nodes"


for node in $(kubectl get nodes -o name); do
  echo "Patching $node"
  kubectl patch $node --subresource status --type=merge -p "{
    \"status\": {
      \"capacity\": { \"${RESOURCE}\": \"${VALUE}\" },
      \"allocatable\": { \"${RESOURCE}\": \"${VALUE}\" }
    }
  }"
done



