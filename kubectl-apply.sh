#!/usr/bin/bash

log() {
  echo "$(date '+%Y-%m-%d %H:%M:%S') - $1"
}

secret_name=$DUCKDNS_DOMAIN-tls

log "Start watching for ingresses"
# Run the script continuously every second
while true; do

  cat /etc/letsencrypt/live/${DUCKDNS_DOMAIN}/fullchain.pem > /tmp/fullchain.pem
  cat /etc/letsencrypt/live/${DUCKDNS_DOMAIN}/privkey.pem > /tmp/privkey.pem


  ingress_namespaces=$(kubectl get ingress --all-namespaces -o jsonpath='{.items[*].metadata.namespace}' | tr ' ' '\n' | sort -u)
  all_namespaces=$(kubectl get namespaces -o jsonpath='{.items[*].metadata.name}' | tr ' ' '\n')

  # Create a new secret in dry-run mode for comparison
  new_secret_yaml=$(kubectl create secret tls $secret_name --cert=/tmp/fullchain.pem --key=/tmp/privkey.pem --dry-run=client -o yaml)


  for namespace in ${ingress_namespaces[@]}; do
    correct_ingress=$(kubectl get ingress -n $namespace -o jsonpath='{.items[*].spec.rules[*].host}')
    if [[ $correct_ingress =~ "$DUCKDNS_DOMAIN" ]]; then 

      # Check if the secret already exists in the namespace
      if kubectl get secret $secret_name -n $namespace &> /dev/null; then
        # Estrarre il secret esistente per confronto
        existing_secret_data=$(kubectl get secret $secret_name -n $namespace -o jsonpath='{.data}')

        # Check if the existing secret type matches the desired type
        new_secret_data=$(echo "$new_secret_yaml" | kubectl create -f - --dry-run=client -o jsonpath='{.data}')
        
        # Compare the data of the existing secret with the new secret
        if [[ "$existing_secret_data" != "$new_secret_data" ]]; then
          # The data is different, delete the existing and apply the update
          kubectl delete secret $secret_name
          log_output=$(echo "$new_secret_yaml" | kubectl apply -n $namespace -f - 2>&1)
          log "Secret updated in $namespace: $log_output"
        fi
      else
        # Secret does not exist, create it
        log_output=$(echo "$new_secret_yaml" | kubectl apply -n $namespace -f - 2>&1)
        log "Secret created in $namespace: $log_output"
      fi
    fi
  done
  
  # duckdns ingress does not exist, delete it
  for namespace in $all_namespaces; do
    correct_ingress=$(kubectl get ingress -n $namespace -o jsonpath='{.items[*].spec.rules[*].host}')
    if [[ ! $correct_ingress =~ "$DUCKDNS_DOMAIN" ]]; then 
      if kubectl get secret $secret_name -n "$namespace" &> /dev/null; then
        kubectl delete secret $secret_name -n "$namespace" &> /dev/null
        log "Secret $secret_name has been deleted from $namespace namespace"
      fi
    fi
  done

  # Wait for 1 second before running the script again
  sleep 10

  rm /tmp/fullchain.pem /tmp/privkey.pem
done

