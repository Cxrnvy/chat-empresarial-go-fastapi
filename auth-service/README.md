# auth-service (FastAPI)

Authentication and authorization service for the enterprise chat. It implements the Phase 2 security
controls C-01 to C-05, C-09 and C-10 (see `docs/fase3-seguridad/`).

## What it does
- Accounts are created by an administrator (no self-registration).
- `POST /auth/login` checks credentials with Argon2id hashes, answers with a generic error, locks an
  account for 15 minutes after 5 failed attempts and issues a signed token.
- Sessions expire after 30 minutes without activity or 8 hours in total, and can be revoked.
- Admin endpoints (`/admin/...`) check the role on the server and deny by default.
- Every security event goes to a JSON Lines log without passwords, tokens or message content.

## Run it (Debian / Codespaces)
```bash
cd auth-service
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
export AUTH_SECRET_KEY="$(python3 -c 'import secrets; print(secrets.token_urlsafe(48))')"
python -m app.cli create-admin --username admin_ana     # asks for the password
uvicorn app.server:app --port 8000
```
The interactive docs are at http://127.0.0.1:8000/docs.

## Test it
```bash
python -m pytest -v          # 17 tests, each named with its PS-xx code
bash scripts/demo_phase3.sh  # real server + curl walk-through (uses demo-only passwords)
```

## Rules
Never commit `.env`, `*.db` or `*.log` (already in `.gitignore`). Real secrets only live in environment variables.
