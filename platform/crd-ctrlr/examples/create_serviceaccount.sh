NAME=test
kubectl create serviceaccount $NAME
SECRET_NAME=`kubectl get sa $NAME -o jsonpath='{.secrets[0].name}'`
TOKEN=`kubectl get secret $SECRET_NAME  -o jsonpath='{.data.token}' | base64 -d`

kubectl config view --raw > ~/$NAME.conf
kubectl --kubeconfig ~/$NAME.conf config rename-context `kubectl --kubeconfig ~/$NAME.conf config current-context` $NAME
kubectl --kubeconfig ~/$NAME.conf config set-credentials sa-$NAME --token $TOKEN
kubectl --kubeconfig ~/$NAME.conf config set-context $NAME --user sa-$NAME
