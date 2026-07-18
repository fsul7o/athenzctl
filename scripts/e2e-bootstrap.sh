#!/usr/bin/env bash
# Bootstrap the e2e environment against a kubernetes-athenz (KinD) stack
# from ctyano/athenz-distribution. Copies the admin mTLS material, waits
# for the workloads to be ready, port-forwards ZMS/ZTS/dex, and writes an
# athenzctl config with three contexts: `local` (mTLS admin), `exec-local`
# (auth-mode: exec against the real ctyano/athenz-user-cert CLI), and
# `exec-broken` (auth-mode: exec against a nonexistent command, for
# error-path coverage).
#
# exec-local runs `athenzusercert` the same way athenz-distribution's own
# `make test-athenz-oauth2` does: via `kubectl exec` into the
# `athenz-user-cert` sidecar container of `oauth2-deployment`, where
# in-cluster service DNS matches the ZTS server cert's SAN and the OIDC
# password grant needs no browser. The resulting cert/key are copied out of
# the pod to local files that athenzctl reads via cert-path/key-path.
#
# Usage: e2e-bootstrap.sh <athenz-distribution-dir> <out-dir>
set -euo pipefail

DIST_DIR="${1:?athenz-distribution dir required}"
OUT_DIR="${2:?output dir required}"

# After `use-kubernetes-crypki-softhsm` (or the equivalent overwrite step)
# runs, the crypki-signed admin material and CA end up under the
# athenz-cli kustomize dir. The openssl-generated files at ${DIST_DIR}/certs
# are stale in that case — the pods present a Crypki SoftHSM Root CA cert
# and only the crypki-signed admin cert will be trusted by ZMS.
ADMIN_CERT_SRC="${DIST_DIR}/kubernetes/athenz-cli/kustomize/certs/athenz_admin.cert.pem"
ADMIN_KEY_SRC="${DIST_DIR}/kubernetes/athenz-cli/kustomize/keys/athenz_admin.private.pem"
ADMIN_CA_SRC="${DIST_DIR}/kubernetes/athenz-cli/kustomize/certs/ca.cert.pem"
# Fallback to the openssl-generated root files if the crypki-signed copies
# were not produced (e.g. the crypki step was skipped).
[ -f "${ADMIN_CERT_SRC}" ] || ADMIN_CERT_SRC="${DIST_DIR}/certs/athenz_admin.cert.pem"
[ -f "${ADMIN_KEY_SRC}"  ] || ADMIN_KEY_SRC="${DIST_DIR}/keys/athenz_admin.private.pem"
[ -f "${ADMIN_CA_SRC}"   ] || ADMIN_CA_SRC="${DIST_DIR}/certs/ca.cert.pem"
for f in "${ADMIN_CERT_SRC}" "${ADMIN_KEY_SRC}" "${ADMIN_CA_SRC}"; do
  if [ ! -f "${f}" ]; then
    echo "expected ${f} — did the kubernetes-athenz deploy finish?" >&2
    exit 1
  fi
done
echo "using admin material from: ${ADMIN_CERT_SRC}"

mkdir -p "${OUT_DIR}/certs" "${OUT_DIR}/keys"
cp "${ADMIN_CA_SRC}"   "${OUT_DIR}/certs/ca.cert.pem"
cp "${ADMIN_CERT_SRC}" "${OUT_DIR}/certs/athenz_admin.cert.pem"
cp "${ADMIN_KEY_SRC}"  "${OUT_DIR}/keys/athenz_admin.private.pem"
chmod 600 "${OUT_DIR}/keys/athenz_admin.private.pem"

CA="${OUT_DIR}/certs/ca.cert.pem"
CERT="${OUT_DIR}/certs/athenz_admin.cert.pem"
KEY="${OUT_DIR}/keys/athenz_admin.private.pem"

mkdir -p "${OUT_DIR}/exec-local"
EXEC_CERT="${OUT_DIR}/exec-local/user.cert.pem"
EXEC_KEY="${OUT_DIR}/exec-local/user.key.pem"

# Local ports for the KinD services.
ZMS_LOCAL_PORT=4443
ZTS_LOCAL_PORT=8443
DEX_LOCAL_PORT=5556
ZMS_URL="https://localhost:${ZMS_LOCAL_PORT}/zms/v1"
ZTS_URL="https://localhost:${ZTS_LOCAL_PORT}/zts/v1"
ZMS_SNI="athenz-zms-server"
ZTS_SNI="athenz-zts-server"

wait_rollout() {
  local dep="$1"
  echo "waiting for deployment/${dep} in namespace athenz ..."
  kubectl -n athenz rollout status "deployment/${dep}" --timeout=180s
}
wait_rollout athenz-zms-server
wait_rollout athenz-zts-server
wait_rollout oauth2-deployment

# Port-forward each service in the background; record PIDs so teardown can
# kill them cleanly. Skip a service whose local port is already in use —
# athenz-distribution's own test-athenz-oauth2 target port-forwards
# oauth2:5556 in the background, and we must not collide with it.
PID_FILE="${OUT_DIR}/pids"
: > "${PID_FILE}"
port_in_use() {
  local p="$1"
  # -sTCP:LISTEN filters to processes actively bound on the port.
  lsof -nP -iTCP:"${p}" -sTCP:LISTEN >/dev/null 2>&1
}
start_pf() {
  local svc="$1" localp="$2" remote="$3"
  if port_in_use "${localp}"; then
    echo "port ${localp} already in use — assuming it's the intended pf for ${svc}, skipping"
    return 0
  fi
  kubectl -n athenz port-forward "svc/${svc}" "${localp}:${remote}" \
    >"${OUT_DIR}/pf-${svc}.log" 2>&1 &
  echo "$!" >> "${PID_FILE}"
}
start_pf athenz-zms-server "${ZMS_LOCAL_PORT}" 4443
start_pf athenz-zts-server "${ZTS_LOCAL_PORT}" 4443
# dex+envoy live in the oauth2 deployment; the oauth2 service exposes both.
start_pf oauth2 "${DEX_LOCAL_PORT}" 5556

wait_https_status() {
  local name="$1" port="$2" sni="$3"
  echo "waiting for ${name} at https://localhost:${port}/${name}/v1/status ..."
  for _ in $(seq 1 60); do
    if curl -sf --resolve "${sni}:${port}:127.0.0.1" \
        --cacert "${CA}" --cert "${CERT}" --key "${KEY}" \
        "https://${sni}:${port}/${name}/v1/status" >/dev/null; then
      echo "${name} ready"
      return 0
    fi
    sleep 1
  done
  echo "${name} did not become ready within 60s" >&2
  exit 1
}
wait_https_status zms "${ZMS_LOCAL_PORT}" "${ZMS_SNI}"
wait_https_status zts "${ZTS_LOCAL_PORT}" "${ZTS_SNI}"

wait_http() {
  local url="$1" name="$2"
  echo "waiting for ${name} at ${url} ..."
  for _ in $(seq 1 60); do
    if curl -sf "${url}" >/dev/null; then
      echo "${name} ready"
      return 0
    fi
    sleep 1
  done
  echo "${name} did not become ready within 60s" >&2
  exit 1
}
wait_http "http://127.0.0.1:${DEX_LOCAL_PORT}/dex/.well-known/openid-configuration" dex

cat >"${OUT_DIR}/config.yaml" <<EOF
current-context: local
contexts:
  - name: local
    zms-url: ${ZMS_URL}
    zts-url: ${ZTS_URL}
    zms-server-name: ${ZMS_SNI}
    zts-server-name: ${ZTS_SNI}
    cert: ${CERT}
    key: ${KEY}
    ca-cert: ${CA}
  - name: exec-local
    zms-url: ${ZMS_URL}
    zts-url: ${ZTS_URL}
    zms-server-name: ${ZMS_SNI}
    zts-server-name: ${ZTS_SNI}
    ca-cert: ${CA}
    auth-mode: exec
    exec:
      command: /bin/sh
      args:
        - -c
        - "kubectl -n athenz exec deployment/oauth2-deployment -c athenz-user-cert -i -- /bin/sh -c 'echo -n password | athenzusercert -oidc-issuer https://oauth2.athenz/dex -endpoint https://athenz-zts-server.athenz:4443/zts/v1/usercert -signer-tls-ca /etc/ssl/certs/ca-certificates.crt -oidc-user athenz_admin@athenz.io -oidc-password-stdin' && kubectl -n athenz exec deployment/oauth2-deployment -c athenz-user-cert -i -- cat /root/.athenz/user.cert.pem > ${EXEC_CERT} && kubectl -n athenz exec deployment/oauth2-deployment -c athenz-user-cert -i -- cat /root/.athenz/user.key.pem > ${EXEC_KEY}"
      cert-path: ${EXEC_CERT}
      key-path: ${EXEC_KEY}
  - name: exec-broken
    zms-url: ${ZMS_URL}
    zts-url: ${ZTS_URL}
    zms-server-name: ${ZMS_SNI}
    zts-server-name: ${ZTS_SNI}
    ca-cert: ${CA}
    auth-mode: exec
    exec:
      command: /nonexistent/athenzusercert
      cert-path: ${EXEC_CERT}
      key-path: ${EXEC_KEY}
EOF
chmod 600 "${OUT_DIR}/config.yaml"

echo "wrote ${OUT_DIR}/config.yaml"
echo "port-forward PIDs: $(tr '\n' ' ' <"${PID_FILE}")"
echo "athenzctl --config ${OUT_DIR}/config.yaml get domain sys.auth"
