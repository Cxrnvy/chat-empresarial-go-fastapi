"""FastAPI authentication service.

Implements the Phase 2 controls that live in this service:
C-01 generic credential check, C-02 attempt limit and lock, C-03 Argon2id hashes,
C-04 signed tokens with expiry and revocation, C-05 role check on the server, C-09 event log,
C-10 secrets from the environment.
"""
import secrets
import sqlite3
import time
from typing import Callable

import jwt
from fastapi import Depends, FastAPI, HTTPException, Request
from pydantic import BaseModel

from . import security
from .audit import AuditLog
from .config import Settings
from .db import connect

GENERIC_LOGIN_ERROR = "Incorrect username or password."
INVALID_SESSION = "Invalid or expired session."


class LoginIn(BaseModel):
    username: str
    password: str


class NewUserIn(BaseModel):
    username: str
    password: str
    role: str = "user"


class ActiveIn(BaseModel):
    active: bool


def create_app(settings: Settings, clock: Callable[[], float] = time.time) -> FastAPI:
    app = FastAPI(title="Enterprise chat - auth service", version="0.3.0")
    db = connect(settings.db_path)
    audit = AuditLog(settings.log_path, clock)
    app.state.db = db
    app.state.settings = settings

    def origin(request: Request) -> str:
        return request.client.host if request.client else "unknown"

    # ------------------------------------------------------------------ sessions

    def issue_token(user: sqlite3.Row) -> tuple[str, int]:
        now = clock()
        sid = secrets.token_urlsafe(16)
        exp = now + settings.max_session_seconds
        db.execute(
            "INSERT INTO sessions (id, user_id, created_at, last_seen, expires_at) VALUES (?,?,?,?,?)",
            (sid, user["id"], now, now, exp),
        )
        db.commit()
        claims = {"sub": user["username"], "role": user["role"], "sid": sid, "iat": int(now), "exp": int(exp)}
        return jwt.encode(claims, settings.secret_key, algorithm="HS256"), settings.max_session_seconds

    def current_session(request: Request) -> dict:
        """Validate the bearer token. Used by every protected endpoint (RF-SEC-01, C-04)."""
        header = request.headers.get("authorization", "")
        if not header.lower().startswith("bearer "):
            audit.event("token_rejected", "denied", origin=origin(request), reason="missing_token")
            raise HTTPException(401, INVALID_SESSION)
        token = header[7:].strip()
        try:
            # Expiry and issue time are checked by us with the injected clock, so tests can move time forward.
            claims = jwt.decode(token, settings.secret_key, algorithms=["HS256"], options={"verify_exp": False, "verify_iat": False, "verify_nbf": False})
        except jwt.PyJWTError:
            audit.event("token_rejected", "denied", origin=origin(request), reason="bad_signature_or_format")
            raise HTTPException(401, INVALID_SESSION)
        now = clock()
        row = db.execute(
            "SELECT s.*, u.username, u.role, u.active FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.id = ?",
            (claims.get("sid"),),
        ).fetchone()
        reason = None
        if row is None or row["username"] != claims.get("sub"):
            reason = "unknown_session"
        elif row["revoked"]:
            reason = "revoked"
        elif not row["active"]:
            reason = "user_disabled"
        elif now >= row["expires_at"]:
            reason = "max_lifetime"
        elif now - row["last_seen"] > settings.idle_seconds:
            reason = "idle_timeout"
        if reason:
            audit.event("token_rejected", "denied", user=claims.get("sub"), origin=origin(request), reason=reason)
            raise HTTPException(401, INVALID_SESSION)
        db.execute("UPDATE sessions SET last_seen = ? WHERE id = ?", (now, row["id"]))
        db.commit()
        # Role comes from the database, not from the token, so a role change applies immediately.
        return {"sid": row["id"], "username": row["username"], "role": row["role"]}

    def require_admin(request: Request, session: dict = Depends(current_session)) -> dict:
        """C-05: deny by default. Anything that is not explicitly an admin gets a 403 and a log line."""
        if session["role"] != "admin":
            audit.event(
                "access_denied", "denied", user=session["username"], origin=origin(request),
                resource=request.url.path, role=session["role"],
            )
            raise HTTPException(403, "You do not have permission to do this.")
        return session

    # ------------------------------------------------------------------ authentication

    @app.post("/auth/login")
    def login(body: LoginIn, request: Request):
        now = clock()
        org = origin(request)
        user = db.execute("SELECT * FROM users WHERE username = ?", (body.username,)).fetchone()

        # Always run one hash verification, with a dummy hash if the user is unknown (C-01).
        password_ok = security.verify_password(user["password_hash"] if user else None, body.password)

        if user is None:
            audit.event("login_failure", "denied", user=body.username[:64], origin=org, reason="unknown_user")
            raise HTTPException(401, GENERIC_LOGIN_ERROR)

        if user["locked_until"] > now:
            audit.event("login_failure", "denied", user=user["username"], origin=org, reason="account_locked")
            raise HTTPException(401, GENERIC_LOGIN_ERROR)

        if not user["active"]:
            audit.event("login_failure", "denied", user=user["username"], origin=org, reason="account_disabled")
            raise HTTPException(401, GENERIC_LOGIN_ERROR)

        if not password_ok:
            failed = user["failed_attempts"] + 1
            if failed >= settings.max_failed_attempts:
                db.execute(
                    "UPDATE users SET failed_attempts = 0, locked_until = ? WHERE id = ?",
                    (now + settings.lock_seconds, user["id"]),
                )
                audit.event("account_locked", "denied", user=user["username"], origin=org,
                            lock_seconds=settings.lock_seconds)
            else:
                db.execute("UPDATE users SET failed_attempts = ? WHERE id = ?", (failed, user["id"]))
            db.commit()
            audit.event("login_failure", "denied", user=user["username"], origin=org, reason="bad_password")
            raise HTTPException(401, GENERIC_LOGIN_ERROR)

        # Success: reset the counter, upgrade the hash if the parameters changed, open a session.
        if security.needs_rehash(user["password_hash"]):
            db.execute("UPDATE users SET password_hash = ? WHERE id = ?",
                       (security.hash_password(body.password), user["id"]))
        db.execute("UPDATE users SET failed_attempts = 0, locked_until = 0 WHERE id = ?", (user["id"],))
        db.commit()
        token, expires_in = issue_token(user)
        audit.event("login_success", "allowed", user=user["username"], origin=org)
        return {"access_token": token, "token_type": "bearer", "expires_in": expires_in, "role": user["role"]}

    @app.post("/auth/logout")
    def logout(request: Request, session: dict = Depends(current_session)):
        db.execute("UPDATE sessions SET revoked = 1 WHERE id = ?", (session["sid"],))
        db.commit()
        audit.event("logout", "allowed", user=session["username"], origin=origin(request))
        return {"detail": "Session closed."}

    @app.get("/auth/validate")
    def validate(session: dict = Depends(current_session)):
        """What the Go chat server calls to check a token and learn the user's role."""
        return {"username": session["username"], "role": session["role"]}

    # ------------------------------------------------------------------ administration (admin only)

    @app.post("/admin/users", status_code=201)
    def create_user(body: NewUserIn, request: Request, admin: dict = Depends(require_admin)):
        if body.role not in ("user", "admin"):
            raise HTTPException(422, "Role must be 'user' or 'admin'.")
        if not security.valid_username(body.username):
            raise HTTPException(422, "Username must be 3-32 characters: lowercase letters, digits or underscore.")
        problems = security.check_password_policy(
            body.password, body.username, settings.min_password_length, settings.max_password_length)
        if problems:
            audit.event("user_create", "denied", user=admin["username"], origin=origin(request),
                        target=body.username, reason="weak_password")
            raise HTTPException(422, "Password " + "; ".join(problems) + ".")
        try:
            db.execute(
                "INSERT INTO users (username, password_hash, role, created_at) VALUES (?,?,?,?)",
                (body.username, security.hash_password(body.password), body.role, clock()),
            )
            db.commit()
        except sqlite3.IntegrityError:
            raise HTTPException(409, "That username already exists.")
        audit.event("user_create", "allowed", user=admin["username"], origin=origin(request),
                    target=body.username, role=body.role)
        return {"username": body.username, "role": body.role, "active": True}

    @app.get("/admin/users")
    def list_users(admin: dict = Depends(require_admin)):
        rows = db.execute("SELECT username, role, active FROM users ORDER BY username").fetchall()
        return [{"username": r["username"], "role": r["role"], "active": bool(r["active"])} for r in rows]

    @app.post("/admin/users/{username}/active")
    def set_active(username: str, body: ActiveIn, request: Request, admin: dict = Depends(require_admin)):
        user = db.execute("SELECT id FROM users WHERE username = ?", (username,)).fetchone()
        if user is None:
            raise HTTPException(404, "User not found.")
        db.execute("UPDATE users SET active = ? WHERE id = ?", (int(body.active), user["id"]))
        if not body.active:
            db.execute("UPDATE sessions SET revoked = 1 WHERE user_id = ?", (user["id"],))
        db.commit()
        audit.event("user_active_change", "allowed", user=admin["username"], origin=origin(request),
                    target=username, active=body.active)
        return {"username": username, "active": body.active}

    @app.get("/admin/sessions")
    def list_sessions(admin: dict = Depends(require_admin)):
        now = clock()
        rows = db.execute(
            "SELECT s.id, u.username, s.created_at, s.last_seen FROM sessions s JOIN users u ON u.id = s.user_id "
            "WHERE s.revoked = 0 AND s.expires_at > ? AND ? - s.last_seen <= ?",
            (now, now, settings.idle_seconds),
        ).fetchall()
        return [{"session_id": r["id"], "username": r["username"]} for r in rows]

    @app.post("/admin/sessions/{session_id}/revoke")
    def revoke_session(session_id: str, request: Request, admin: dict = Depends(require_admin)):
        cur = db.execute("UPDATE sessions SET revoked = 1 WHERE id = ?", (session_id,))
        db.commit()
        if cur.rowcount == 0:
            raise HTTPException(404, "Session not found.")
        audit.event("session_revoke", "allowed", user=admin["username"], origin=origin(request), target=session_id)
        return {"detail": "Session revoked."}

    @app.get("/admin/stats")
    def stats(admin: dict = Depends(require_admin)):
        """Aggregated numbers only: no per-person data and no message content (data minimisation)."""
        now = clock()
        total = db.execute("SELECT COUNT(*) FROM users").fetchone()[0]
        admins = db.execute("SELECT COUNT(*) FROM users WHERE role = 'admin'").fetchone()[0]
        live = db.execute(
            "SELECT COUNT(*) FROM sessions WHERE revoked = 0 AND expires_at > ? AND ? - last_seen <= ?",
            (now, now, settings.idle_seconds),
        ).fetchone()[0]
        return {"users": total, "admins": admins, "active_sessions": live}

    return app
