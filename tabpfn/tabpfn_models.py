"""Resolve local TabPFN checkpoints before falling back to a download."""

import os
from pathlib import Path


CLASSIFIER_MODEL_FILENAME = "tabpfn-v3-classifier-v3_default.ckpt"
REGRESSOR_MODEL_FILENAME = "tabpfn-v3-regressor-v3_default.ckpt"
PROJECT_DIRECTORY = Path(__file__).resolve().parent


def resolve_model_path(
    requested_path: Path | None,
    default_filename: str,
) -> Path | None:
    """Return an explicit or discovered checkpoint, or None for TabPFN auto."""
    if requested_path is not None:
        resolved_path = requested_path.expanduser().resolve()
        if not resolved_path.is_file():
            raise FileNotFoundError(f"TabPFN checkpoint not found: {resolved_path}")
        return resolved_path

    configured_directory = os.environ.get("TABPFN_MODEL_DIR")
    candidate_directories = []
    if configured_directory:
        candidate_directories.append(Path(configured_directory).expanduser())
    candidate_directories.extend(
        (
            PROJECT_DIRECTORY / "models",
            PROJECT_DIRECTORY.parent / "tabpfn_3",
        )
    )

    for directory in candidate_directories:
        candidate = (directory / default_filename).resolve()
        if candidate.is_file():
            return candidate

    return None

