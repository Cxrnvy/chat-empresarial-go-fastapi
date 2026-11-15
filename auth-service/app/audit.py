"""Security event log (control C-09, requirement RF-SEC-07).

One JSON object per line with date, user, origin, event and result. Passwords, tokens and message
content are never written here. The file is created with 0600 permissions so only the service
account can read it.
"""
import json
import os
import threading
from datetime import datetime, timezone


class AuditLog:
    def __init__(self, path: str, clock):
        self._path = path
        self._clock = clock
        self._lock = threading.Lock()
        os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
        fd = os.open(path, os.O_CREAT | os.O_APPEND | os.O_WRONLY, 0o600)
        os.close(fd)

    def event(self, event: str, result: str, user: str | None = None, origin: str | None = None, **extra):
        record = {
            "ts": datetime.fromtimestamp(self._clock(), tz=timezone.utc).isoformat(timespec="seconds"),
            "event": event,
            "user": user,
            "origin": origin,
            "result": result,
        }
        record.update(extra)
        line = json.dumps(record, ensure_ascii=False)
        with self._lock, open(self._path, "a", encoding="utf-8") as fh:
            fh.write(line + "\n")
