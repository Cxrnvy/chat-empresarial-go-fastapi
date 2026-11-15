"""Password hashing and password policy (controls C-03 and part of C-01)."""
import re
from pathlib import Path

from argon2 import PasswordHasher
from argon2.exceptions import InvalidHashError, VerificationError, VerifyMismatchError

# PasswordHasher() uses Argon2id by default and draws a fresh random salt on every call to hash(),
# so two users with the same password end up with different hashes (RF-SEC-02).
_hasher = PasswordHasher()

# A real hash of a throwaway string. When someone logs in with a username that does not exist we
# still verify against this, so the answer takes about as long as for a real account (C-01).
_DUMMY_HASH = _hasher.hash("dummy-value-never-used-by-anyone")

_COMMON_FILE = Path(__file__).resolve().parent.parent / "data" / "common_passwords.txt"
_USERNAME_RE = re.compile(r"^[a-z][a-z0-9_]{2,31}$")


def _load_common() -> frozenset[str]:
    if not _COMMON_FILE.exists():
        return frozenset()
    return frozenset(
        line.strip().lower() for line in _COMMON_FILE.read_text(encoding="utf-8").splitlines() if line.strip()
    )


COMMON_PASSWORDS = _load_common()


def valid_username(name: str) -> bool:
    return bool(_USERNAME_RE.match(name))


def check_password_policy(password: str, username: str, min_len: int, max_len: int) -> list[str]:
    """Return a list of problems; an empty list means the password is acceptable.

    Following NIST SP 800-63B we ask for length and a blocklist check, not for forced mixes of
    symbols, because composition rules help less than length does.
    """
    problems = []
    if len(password) < min_len:
        problems.append(f"must have at least {min_len} characters")
    if len(password) > max_len:
        problems.append(f"must have at most {max_len} characters")
    if password.lower() in COMMON_PASSWORDS:
        problems.append("is in the list of commonly used passwords")
    if username and username.lower() in password.lower():
        problems.append("must not contain the username")
    return problems


def hash_password(password: str) -> str:
    return _hasher.hash(password)


def verify_password(stored_hash: str | None, password: str) -> bool:
    """Compare a password with a stored hash. A missing hash is checked against the dummy one."""
    try:
        return _hasher.verify(stored_hash or _DUMMY_HASH, password) and stored_hash is not None
    except (VerifyMismatchError, VerificationError, InvalidHashError):
        return False


def needs_rehash(stored_hash: str) -> bool:
    return _hasher.check_needs_rehash(stored_hash)
