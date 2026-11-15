import json
import sqlite3

import pytest
from fastapi.testclient import TestClient

from app import security
from app.config import Settings
from app.main import create_app

# Throwaway values for tests only. They protect nothing real.
SECRET = "test-only-signing-secret-0123456789abcdef"
ADMIN_PW = "Admin-Test-Phrase-2026"
USER_PW = "Plain-User-Phrase-2026"


class FakeClock:
    """A clock we can move forward so the 15-minute and 30-minute rules are testable in milliseconds."""

    def __init__(self, start=1_800_000_000.0):
        self.t = start

    def __call__(self):
        return self.t

    def advance(self, seconds):
        self.t += seconds


@pytest.fixture
def env(tmp_path):
    settings = Settings(secret_key=SECRET, db_path=str(tmp_path / "auth.db"), log_path=str(tmp_path / "security.log"))
    clock = FakeClock()
    app = create_app(settings, clock)
    client = TestClient(app)
    # Seed one admin straight into the database (this is what the CLI does).
    app.state.db.execute(
        "INSERT INTO users (username, password_hash, role, created_at) VALUES (?,?,?,?)",
        ("admin_ana", security.hash_password(ADMIN_PW), "admin", clock()),
    )
    app.state.db.commit()
    return type("Env", (), {"client": client, "clock": clock, "settings": settings, "app": app})


def login(env, username, password):
    return env.client.post("/auth/login", json={"username": username, "password": password})


def bearer(resp):
    return {"Authorization": "Bearer " + resp.json()["access_token"]}


@pytest.fixture
def admin_headers(env):
    return bearer(login(env, "admin_ana", ADMIN_PW))


@pytest.fixture
def make_user(env, admin_headers):
    def _make(username, password=USER_PW, role="user"):
        r = env.client.post("/admin/users", json={"username": username, "password": password, "role": role},
                            headers=admin_headers)
        assert r.status_code == 201, r.text
    return _make


def read_log(env):
    with open(env.settings.log_path, encoding="utf-8") as fh:
        return [json.loads(line) for line in fh if line.strip()]
