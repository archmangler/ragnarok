for i in role-binding.yaml role.yaml service-account.yaml
do 
  kubectl apply -f $i
done
