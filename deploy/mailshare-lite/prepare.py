#!/usr/bin/env python3
"""Create an isolated local-only Lite test workspace from the existing dev test.

Creates the target directory only when it does not exist. Secrets remain on
the OCI host; this script does not print them or commit them to Git.
"""
import argparse
import pathlib
import shutil
import sqlite3
import sys


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", default="/opt/lanqin-mailshare-test")
    parser.add_argument("--target", default="/opt/lanqin-mailshare-lite")
    args = parser.parse_args()
    src = pathlib.Path(args.source).resolve()
    dst = pathlib.Path(args.target).resolve()
    repo_dir = pathlib.Path(__file__).resolve().parent

    if dst.exists():
        raise SystemExit(f"Refusing to overwrite existing Lite workspace: {dst}")
    for path in (src / ".env", src / "data" / "lanqin.db", repo_dir / "compose.yaml"):
        if not path.is_file():
            raise SystemExit(f"Missing required file: {path}")

    # Copy only local test data, not the live/original LanQin deployment.
    # Do not copy SQLite -wal/-shm: use the online backup API instead.
    dst.mkdir(parents=True)
    def ignore_database(folder, names):
        return [n for n in names if n in ("lanqin.db", "lanqin.db-shm", "lanqin.db-wal")]
    shutil.copytree(src / "data", dst / "data", ignore=ignore_database)
    if (src / "mail").is_dir():
        shutil.copytree(src / "mail", dst / "mail")
    else:
        (dst / "mail").mkdir()
    shutil.copy2(src / ".env", dst / ".env")
    (dst / ".env").chmod(0o600)
    shutil.copy2(repo_dir / "compose.yaml", dst / "compose.yaml")

    source_db = sqlite3.connect("file:" + str((src / "data" / "lanqin.db").resolve()) + "?mode=ro", uri=True)
    target_db = sqlite3.connect(dst / "data" / "lanqin.db")
    try:
        source_db.backup(target_db)
        result = target_db.execute("PRAGMA quick_check").fetchone()
        if result is None or result[0] != "ok":
            raise RuntimeError(f"SQLite snapshot failed quick_check: {result}")
    finally:
        target_db.close()
        source_db.close()
    print(f"Lite test workspace ready: {dst}")
    print("The original and development databases were not modified.")
    print("Local test port: 127.0.0.1:18090")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"Could not prepare Lite test workspace: {exc}", file=sys.stderr)
        sys.exit(1)
