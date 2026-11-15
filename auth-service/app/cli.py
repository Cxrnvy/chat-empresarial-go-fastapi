"""Command line helper to create the first administrator.

There is no self-registration (accounts are created by the organisation), so somebody has to make
the first admin. The password is read from a prompt (or AUTH_ADMIN_PASSWORD for scripts) and is
never printed or stored: only its Argon2id hash goes to the database.

Usage:  python -m app.cli create-admin --username admin_ana
"""
import argparse
import getpass
import os
import sqlite3
import sys
import time

from . import security
from .config import load_settings
from .db import connect


def create_admin(username: str) -> int:
    settings = load_settings()
    if not security.valid_username(username):
        print("Invalid username (3-32 chars: lowercase letters, digits, underscore).", file=sys.stderr)
        return 2
    password = os.environ.get("AUTH_ADMIN_PASSWORD") or getpass.getpass("Admin password: ")
    problems = security.check_password_policy(
        password, username, settings.min_password_length, settings.max_password_length)
    if problems:
        print("Password " + "; ".join(problems) + ".", file=sys.stderr)
        return 2
    db = connect(settings.db_path)
    try:
        db.execute(
            "INSERT INTO users (username, password_hash, role, created_at) VALUES (?,?,?,?)",
            (username, security.hash_password(password), "admin", time.time()),
        )
        db.commit()
    except sqlite3.IntegrityError:
        print("That username already exists.", file=sys.stderr)
        return 1
    print(f"Administrator '{username}' created.")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = parser.add_subparsers(dest="cmd", required=True)
    p = sub.add_parser("create-admin")
    p.add_argument("--username", required=True)
    args = parser.parse_args()
    return create_admin(args.username)


if __name__ == "__main__":
    sys.exit(main())
