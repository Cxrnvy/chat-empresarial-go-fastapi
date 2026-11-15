"""Security tests for Phase 3. Test names carry the PS-xx code of the Phase 2 test plan."""
import sqlite3
import jwt

from app import security
from .conftest import ADMIN_PW, SECRET, USER_PW, bearer, login, read_log


# ---------------------------------------------------------------- authentication (C-01, C-02)

def test_ps01_valid_login_issues_token_and_logs(env, make_user):
    make_user("maria_perez")
    r = login(env, "maria_perez", USER_PW)
    assert r.status_code == 200
    body = r.json()
    assert body["token_type"] == "bearer" and body["role"] == "user" and body["expires_in"] == 8 * 3600
    # The token really opens a protected endpoint
    v = env.client.get("/auth/validate", headers=bearer(r))
    assert v.status_code == 200 and v.json() == {"username": "maria_perez", "role": "user"}
    assert any(e["event"] == "login_success" and e["user"] == "maria_perez" for e in read_log(env))


def test_ps02_wrong_password_and_unknown_user_look_identical(env, make_user):
    make_user("maria_perez")
    wrong = login(env, "maria_perez", "Not-The-Right-Phrase-1")
    ghost = login(env, "nobody_here", "Not-The-Right-Phrase-1")
    assert wrong.status_code == ghost.status_code == 401
    assert wrong.json() == ghost.json()          # same message: no account enumeration
    reasons = {e.get("reason") for e in read_log(env) if e["event"] == "login_failure"}
    assert {"bad_password", "unknown_user"} <= reasons   # but the log keeps the real reason


def test_ps03_five_failures_lock_account_for_15_minutes(env, make_user):
    make_user("maria_perez")
    for _ in range(5):
        assert login(env, "maria_perez", "Wrong-Phrase-12345").status_code == 401
    # Even the correct password is refused while the lock is active
    assert login(env, "maria_perez", USER_PW).status_code == 401
    env.clock.advance(14 * 60)
    assert login(env, "maria_perez", USER_PW).status_code == 401
    env.clock.advance(61)                          # now past 15 minutes
    assert login(env, "maria_perez", USER_PW).status_code == 200
    assert any(e["event"] == "account_locked" for e in read_log(env))


def test_ps03b_a_good_login_resets_the_failure_counter(env, make_user):
    make_user("maria_perez")
    for _ in range(4):
        login(env, "maria_perez", "Wrong-Phrase-12345")
    assert login(env, "maria_perez", USER_PW).status_code == 200
    for _ in range(4):
        login(env, "maria_perez", "Wrong-Phrase-12345")
    assert login(env, "maria_perez", USER_PW).status_code == 200   # still not locked


# ---------------------------------------------------------------- tokens and sessions (C-04)

def test_ps04_protected_resource_rejects_missing_tampered_and_expired_tokens(env, make_user):
    make_user("maria_perez")
    good = login(env, "maria_perez", USER_PW).json()["access_token"]
    assert env.client.get("/auth/validate").status_code == 401                                    # no token
    tampered = good[:-3] + ("AAA" if not good.endswith("AAA") else "BBB")
    assert env.client.get("/auth/validate", headers={"Authorization": "Bearer " + tampered}).status_code == 401
    env.clock.advance(8 * 3600 + 1)
    assert env.client.get("/auth/validate", headers={"Authorization": "Bearer " + good}).status_code == 401


def test_ps04b_forged_admin_token_signed_with_another_key_is_rejected(env, make_user):
    make_user("maria_perez")
    sid = env.app.state.db.execute("SELECT id FROM sessions LIMIT 1").fetchone()
    forged = jwt.encode({"sub": "maria_perez", "role": "admin", "sid": sid["id"] if sid else "x",
                         "exp": int(env.clock()) + 3600}, "attacker-key-attacker-key-attacker-key", algorithm="HS256")
    r = env.client.get("/admin/users", headers={"Authorization": "Bearer " + forged})
    assert r.status_code == 401


def test_ps08_logout_revokes_token_and_idle_session_expires(env, make_user):
    make_user("maria_perez")
    first = login(env, "maria_perez", USER_PW)
    h = bearer(first)
    assert env.client.post("/auth/logout", headers=h).status_code == 200
    assert env.client.get("/auth/validate", headers=h).status_code == 401      # reuse after logout

    second = login(env, "maria_perez", USER_PW)
    h2 = bearer(second)
    env.clock.advance(29 * 60)
    assert env.client.get("/auth/validate", headers=h2).status_code == 200     # activity renews idle timer
    env.clock.advance(31 * 60)                                                  # 31 min with no activity
    assert env.client.get("/auth/validate", headers=h2).status_code == 401


def test_ps08b_admin_can_revoke_a_live_session(env, make_user, admin_headers):
    make_user("maria_perez")
    h = bearer(login(env, "maria_perez", USER_PW))
    sessions = env.client.get("/admin/sessions", headers=admin_headers).json()
    sid = next(s["session_id"] for s in sessions if s["username"] == "maria_perez")
    assert env.client.post(f"/admin/sessions/{sid}/revoke", headers=admin_headers).status_code == 200
    assert env.client.get("/auth/validate", headers=h).status_code == 401


def test_disabling_a_user_kills_their_sessions(env, make_user, admin_headers):
    make_user("maria_perez")
    h = bearer(login(env, "maria_perez", USER_PW))
    env.client.post("/admin/users/maria_perez/active", json={"active": False}, headers=admin_headers)
    assert env.client.get("/auth/validate", headers=h).status_code == 401
    assert login(env, "maria_perez", USER_PW).status_code == 401


# ---------------------------------------------------------------- authorization (C-05)

def test_ps05_admin_can_list_users_and_see_stats(env, make_user, admin_headers):
    make_user("maria_perez")
    r = env.client.get("/admin/users", headers=admin_headers)
    assert r.status_code == 200
    names = [u["username"] for u in r.json()]
    assert names == ["admin_ana", "maria_perez"]
    assert all("password" not in u and "password_hash" not in u for u in r.json())
    assert env.client.get("/admin/stats", headers=admin_headers).json()["users"] == 2


def test_ps06_regular_user_is_denied_admin_functions_and_it_is_logged(env, make_user):
    make_user("maria_perez")
    h = bearer(login(env, "maria_perez", USER_PW))
    create = env.client.post("/admin/users", json={"username": "intruder", "password": "Another-Phrase-2026"}, headers=h)
    stats = env.client.get("/admin/stats", headers=h)
    assert create.status_code == stats.status_code == 403
    denied = [e for e in read_log(env) if e["event"] == "access_denied"]
    assert {d["resource"] for d in denied} == {"/admin/users", "/admin/stats"}
    assert all(d["user"] == "maria_perez" for d in denied)
    # and the account was not created
    names = [u["username"] for u in env.client.get("/admin/users", headers=bearer(login(env, "admin_ana", ADMIN_PW))).json()]
    assert "intruder" not in names


# ---------------------------------------------------------------- credential protection (C-03)

def test_ps09_only_argon2id_hashes_are_stored_and_they_differ(env, make_user):
    make_user("maria_perez")
    make_user("jorge_diaz")                          # same password on purpose
    login(env, "maria_perez", "Wrong-Phrase-12345")  # leave a failed attempt in the log too
    con = sqlite3.connect(env.settings.db_path)
    rows = dict(con.execute("SELECT username, password_hash FROM users").fetchall())
    assert all(h.startswith("$argon2id$") for h in rows.values())
    assert rows["maria_perez"] != rows["jorge_diaz"]           # unique salt per hash
    raw_db = open(env.settings.db_path, "rb").read()
    raw_log = open(env.settings.log_path, "rb").read()
    for secret in (USER_PW, ADMIN_PW, "Wrong-Phrase-12345"):
        assert secret.encode() not in raw_db
        assert secret.encode() not in raw_log


def test_password_policy_rejects_short_common_and_username_passwords(env, admin_headers):
    def create(pw, name="nuevo_user"):
        return env.client.post("/admin/users", json={"username": name, "password": pw}, headers=admin_headers)
    assert create("Short-1").status_code == 422                 # fewer than 12 characters
    assert create("password123").status_code == 422             # short and common
    assert create("contrasena1234").status_code == 422          # long enough but on the blocklist
    assert create("nuevo_user-2026-ok").status_code == 422      # contains the username
    assert create("Long-Enough-Phrase-2026").status_code == 201


def test_duplicate_username_is_rejected(env, make_user, admin_headers):
    make_user("maria_perez")
    r = env.client.post("/admin/users", json={"username": "maria_perez", "password": "Another-Phrase-2026"},
                        headers=admin_headers)
    assert r.status_code == 409


def test_hash_verification_unit():
    h = security.hash_password("Some-Phrase-For-Unit-Test")
    assert security.verify_password(h, "Some-Phrase-For-Unit-Test")
    assert not security.verify_password(h, "Some-Phrase-For-Unit-Tesx")
    assert not security.verify_password(None, "anything")


# ---------------------------------------------------------------- logging (C-09)

def test_ps13_log_has_date_user_origin_result_and_no_secrets(env, make_user):
    make_user("maria_perez")
    login(env, "nobody_here", "Wrong-Phrase-12345")
    for _ in range(5):
        login(env, "maria_perez", "Wrong-Phrase-12345")
    h = bearer(login(env, "admin_ana", ADMIN_PW))
    env.client.post("/auth/logout", headers=h)
    events = read_log(env)
    assert {"login_success", "login_failure", "account_locked", "logout", "user_create"} <= {e["event"] for e in events}
    for e in events:
        assert {"ts", "event", "user", "origin", "result"} <= set(e)
        assert "password" not in e and "token" not in e
    text = open(env.settings.log_path, encoding="utf-8").read()
    for secret in (USER_PW, ADMIN_PW, "Wrong-Phrase-12345", SECRET):
        assert secret not in text


def test_service_refuses_to_start_without_a_secret(monkeypatch):
    import pytest
    from app.config import load_settings
    monkeypatch.delenv("AUTH_SECRET_KEY", raising=False)
    with pytest.raises(RuntimeError):
        load_settings()                               # C-10: no secret, no service
