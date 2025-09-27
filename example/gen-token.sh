#!/bin/bash
# 一键生成 ServiceAccount，并输出可执行的 kubectl 命令

NAMESPACE=default
SA_NAME=my-token-sa

# 获取 API Server 地址
APISERVER=$(kubectl config view -o jsonpath='{.clusters[0].cluster.server}')

# 创建 ServiceAccount
kubectl create sa ${SA_NAME} -n ${NAMESPACE} 2>/dev/null || true

# 绑定 cluster-admin 权限（可改成更细粒度权限）
kubectl create clusterrolebinding ${SA_NAME}-binding \
  --clusterrole=cluster-admin \
  --serviceaccount=${NAMESPACE}:${SA_NAME} 2>/dev/null || true

# 获取 Secret 名称 (v1.24+ 需要显式创建)
SECRET_NAME=$(kubectl get sa ${SA_NAME} -n ${NAMESPACE} -o jsonpath='{.secrets[0].name}')

if [ -z "$SECRET_NAME" ]; then
  # 手动创建 Secret
  cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: ${SA_NAME}-secret
  namespace: ${NAMESPACE}
  annotations:
    kubernetes.io/service-account.name: ${SA_NAME}
type: kubernetes.io/service-account-token
EOF
  SECRET_NAME=${SA_NAME}-secret
  # 等待 secret 生效
  sleep 2
fi

# 输出 Token
TOKEN=$(kubectl get secret ${SECRET_NAME} -n ${NAMESPACE} -o jsonpath='{.data.token}' | base64 --decode)

echo "=================="
echo "生成的 Token:"
echo "${TOKEN}"
echo "=================="
echo ""
echo "你可以直接运行以下命令获取 Pod 列表："
echo ""
echo "kubectl --server=${APISERVER} --token=\"${TOKEN}\" --insecure-skip-tls-verify=true get pods -A"
