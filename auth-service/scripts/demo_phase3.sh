#!/usr/bin/env bash
# Phase 3 demo: starts the auth service on a throwaway database and walks through the controls,
# printing what a reviewer needs to see. Everything here uses demo-only values; nothing is real.
# Run from the auth-service folder:   bash scripts/demo_phase3.sh
set -u
cd "$(dirname "$0")/.."
WORK="$(mktemp -d)"
export AUTH_DB_PATH="$WORK/auth.db" AUTH_LOG_PATH="$WORK/security.log"
export AUTH_SECRET_KEY="$(python3 -c 'import secrets; print(secrets.token_urlsafe(48))')"
ADMIN_PW="Demo-Admin-Phrase-2026"; USER_PW="Demo-User-Phrase-2026"; BASE="http://127.0.0.1:8000"

say()  { printf '\n=== %s\n' "$*"; }
# call METHOD PATH [TOKEN] [JSON]  -> prints "HTTP <code> <body>" (tokens are never printed in full)
call() {
  local m="$1" p="$2" t="${3:-}" d="${4:-}"
  local args=(-s -o "$WORK/body" -w '%{http_code}' -X "$m" "$BASE$p")
  [ -n "$t" ] && args+=(-H "Authorization: Bearer $t")
  [ -n "$d" ] && args+=(-H 'Content-Type: application/json' -d "$d")
  local code; code="$(curl "${args[@]}")"
  printf '  %s %s -> HTTP %s  %s\n' "$m" "$p" "$code" "$(python3 - "$WORK/body" <<'PY'
import json,sys
raw=open(sys.argv[1]).read()
try:
    d=json.loads(raw)
    if isinstance(d,dict) and "access_token" in d: d["access_token"]=d["access_token"][:14]+"...(masked)"
    print(json.dumps(d))
except Exception: print(raw)
PY
)"
  LAST_CODE="$code"
}
token_of() { python3 -c 'import json,sys;print(json.load(open(sys.argv[1]))["access_token"])' "$WORK/body"; }

say "0. Create the first administrator with the CLI (password read from the environment, never printed)"
AUTH_ADMIN_PASSWORD="$ADMIN_PW" python3 -m app.cli create-admin --username admin_ana

uvicorn app.server:app --port 8000 --log-level warning & SRV=$!
trap 'kill $SRV 2>/dev/null' EXIT
for _ in $(seq 1 40); do curl -s "$BASE/docs" >/dev/null && break; sleep 0.25; done

say "1. AUTHENTICATION: valid login (PS-01)"
call POST /auth/login "" "{\"username\":\"admin_ana\",\"password\":\"$ADMIN_PW\"}"; ADMIN_T="$(token_of)"

say "2. USER CREATION by an administrator, including a weak password (RF-SEC-02)"
call POST /admin/users "$ADMIN_T" '{"username":"maria_perez","password":"corta123"}'
call POST /admin/users "$ADMIN_T" '{"username":"maria_perez","password":"contrasena1234"}'
call POST /admin/users "$ADMIN_T" "{\"username\":\"maria_perez\",\"password\":\"$USER_PW\"}"
call POST /admin/users "$ADMIN_T" "{\"username\":\"jorge_diaz\",\"password\":\"$USER_PW\"}"

say "3. INVALID CREDENTIALS: wrong password vs. unknown user give the same answer (PS-02)"
call POST /auth/login "" '{"username":"maria_perez","password":"Not-The-Phrase-0000"}'
call POST /auth/login "" '{"username":"nobody_here","password":"Not-The-Phrase-0000"}'

say "4. AUTHORIZATION: allowed vs. denied (PS-05, PS-06)"
call POST /auth/login "" "{\"username\":\"maria_perez\",\"password\":\"$USER_PW\"}"; USER_T="$(token_of)"
call GET /auth/validate "$USER_T"
call GET /admin/stats "$USER_T"
call POST /admin/users "$USER_T" '{"username":"intruder","password":"Another-Phrase-2026"}'
call GET /admin/stats "$ADMIN_T"
call GET /admin/users "$ADMIN_T"

say "5. TOKEN CHECKS (PS-04): no token and a tampered token"
call GET /auth/validate
call GET /auth/validate "${USER_T:0:${#USER_T}-4}AAAA"

say "6. LOGOUT, then reuse the closed token (PS-08)"
call POST /auth/logout "$USER_T"
call GET /auth/validate "$USER_T"

say "7. LOCKOUT after 5 failures, then the CORRECT password is still refused (PS-03)"
for i in 1 2 3 4 5; do call POST /auth/login "" '{"username":"jorge_diaz","password":"Wrong-Phrase-12345"}'; done
call POST /auth/login "" "{\"username\":\"jorge_diaz\",\"password\":\"$USER_PW\"}"

say "8. CREDENTIAL STORAGE: what the database really holds (PS-09)"
python3 - "$AUTH_DB_PATH" <<'PY'
import sqlite3, sys
con = sqlite3.connect(sys.argv[1])
for name, h in con.execute("SELECT username, password_hash FROM users ORDER BY username"):
    print(f"  {name:<11} {h[:58]}...")
raw = open(sys.argv[1], "rb").read()
for label, pw in (("admin demo password", "Demo-Admin-Phrase-2026"), ("user demo password", "Demo-User-Phrase-2026")):
    print(f"  '{label}' found inside the database file? {pw.encode() in raw}")
PY
python3 - "$AUTH_DB_PATH" <<'PY'
import sqlite3, sys
r = dict(sqlite3.connect(sys.argv[1]).execute("SELECT username, password_hash FROM users"))
print("  maria_perez and jorge_diaz share a password; same hash? ", r["maria_perez"] == r["jorge_diaz"])
PY

say "9. EVENT LOG (PS-13): selected lines and a search for secrets"
python3 - "$AUTH_LOG_PATH" <<'PY'
import json, sys
lines = [json.loads(l) for l in open(sys.argv[1])]
want = ["login_success", "login_failure", "account_locked", "access_denied", "token_rejected", "logout", "user_create"]
seen = set()
for e in lines:
    if e["event"] in want and (e["event"], e.get("reason", e["result"])) not in seen:
        seen.add((e["event"], e.get("reason", e["result"])))
        print("  " + json.dumps(e, ensure_ascii=False))
text = open(sys.argv[1]).read()
for pw in ("Demo-Admin-Phrase-2026", "Demo-User-Phrase-2026", "Wrong-Phrase-12345", "corta123"):
    print(f"  '{pw}' found in log? {pw in text}")
PY
ls -l "$AUTH_LOG_PATH" | awk '{print "  log file permissions:", $1}'
echo; echo "Demo finished. Temporary data: $WORK"
