"""Windows compatibility for TabPFN 8.2's interactive browser login."""

from collections.abc import Iterator
from contextlib import contextmanager
import sys
import threading


def _poll_for_token_on_windows(
    auth_event: threading.Event,
    received_token: list[str | None],
) -> str | None:
    """Wait for the browser callback while accepting masked console input."""
    import msvcrt

    prompt = "API key (or press Enter to keep waiting): "
    sys.stdout.write(prompt)
    sys.stdout.flush()
    characters: list[str] = []

    while not auth_event.wait(0.05):
        while msvcrt.kbhit():
            character = msvcrt.getwch()

            if character in ("\x00", "\xe0"):
                # Consume the second code unit emitted by function/arrow keys.
                msvcrt.getwch()
                continue
            if character == "\x03":
                raise KeyboardInterrupt
            if character in ("\r", "\n"):
                sys.stdout.write("\n")
                sys.stdout.flush()
                token = "".join(characters).strip()
                if token:
                    return token
                sys.stdout.write(prompt)
                sys.stdout.flush()
                continue
            if character == "\b":
                if characters:
                    characters.pop()
                    sys.stdout.write("\b \b")
                    sys.stdout.flush()
                continue
            if character.isprintable():
                characters.append(character)
                # Avoid echoing an API key into terminal logs.
                sys.stdout.write("*")
                sys.stdout.flush()

    return received_token[0]


@contextmanager
def windows_browser_auth_workaround() -> Iterator[None]:
    """Temporarily replace TabPFN's Unix-only stdin poll on Windows."""
    if sys.platform != "win32":
        yield
        return

    from tabpfn import browser_auth

    original_poll = browser_auth._poll_for_token
    browser_auth._poll_for_token = _poll_for_token_on_windows
    try:
        yield
    finally:
        browser_auth._poll_for_token = original_poll

