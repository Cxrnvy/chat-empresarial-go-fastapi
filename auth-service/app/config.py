"""Settings for the service.

Every security value that Phase 2 fixed (5 attempts, 15 minutes, 30 minutes, 8 hours, 12 characters)
lives here as a default, so tests can shrink them and the rest of the code never hard-codes numbers.
The signing secret is deliberately NOT given a default: control C-10 says secrets come from the
environment, so the service refuses to start without one.
"""
import os
from dataclasses import dataclass


@dataclass(frozen=True)
class Settings:
    secret_key: str
    db_path: str = "data/auth.db"
    log_path: str = "data/security.log"
    max_failed_attempts: int = 5          # RF-SEC-05
    lock_seconds: int = 15 * 60           # RF-SEC-05
    idle_seconds: int = 30 * 60           # RF-SEC-05
    max_session_seconds: int = 8 * 60 * 60  # RF-SEC-05
    min_password_length: int = 12         # RF-SEC-02
    max_password_length: int = 128        # avoids sending huge inputs to the hash function


def load_settings() -> Settings:
    secret = os.environ.get("AUTH_SECRET_KEY", "")
    if len(secret) < 32:
        raise RuntimeError(
            "AUTH_SECRET_KEY is missing or too short (min. 32 characters). "
            "Set it in the environment; it must never live in the repository (C-10)."
        )
    return Settings(
        secret_key=secret,
        db_path=os.environ.get("AUTH_DB_PATH", "data/auth.db"),
        log_path=os.environ.get("AUTH_LOG_PATH", "data/security.log"),
    )
