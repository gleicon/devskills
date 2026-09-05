"""Persistent notes the agent carries between sessions."""

STORE = {}


def remember(text):
    # Whatever the model decided was worth keeping.
    STORE.setdefault("notes", []).append(text)


def recall():
    return STORE.get("notes", [])
